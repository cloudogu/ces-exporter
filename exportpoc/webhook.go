// This file handles pod admission and builds patches for read-only exporter sidecars.

package exportpoc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// These names identify the trigger, webhook state and resources owned by the PoC.
const (
	TriggerConfigMap   = "ces-exporter-export-mode"
	WebhookConfigMap   = "ces-exporter-webhook-config"
	SidecarName        = "ces-exporter-sidecar"
	keyVolume          = "ces-exporter-public-key"
	injectedAnnotation = "k8s.cloudogu.com/exporter-poc-injected"
)

// patchOperation represents a JSON Patch operation in the admission response.
type patchOperation struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value"`
}

// admission returns an AdmissionReview containing the injection patch; injection errors allow the pod and produce a visible warning.
func (r *Runtime) admission(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "POST required", http.StatusMethodNotAllowed)
		return
	}
	var review admissionv1.AdmissionReview
	if err := json.NewDecoder(http.MaxBytesReader(w, req.Body, 4<<20)).Decode(&review); err != nil || review.Request == nil {
		http.Error(w, "invalid AdmissionReview", http.StatusBadRequest)
		return
	}
	response := &admissionv1.AdmissionResponse{UID: review.Request.UID, Allowed: true}
	patch, err := r.mutate(req.Context(), review.Request)
	if err != nil {
		slog.Error("PoC admission failed; allowing pod without exporter", "error", err)
		response.Warnings = []string{"exporter injection failed; inspect ces-exporter logs"}
	}
	if len(patch) > 0 && err == nil {
		response.PatchType = ptr.To(admissionv1.PatchTypeJSONPatch)
		response.Patch = patch
	}
	review.Request = nil
	review.Response = response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(review); err != nil {
		slog.Error("write admission response", "error", err)
	}
}

// mutate restricts injection to eligible pod creations in this namespace and reads the current mode directly from Kubernetes.
func (r *Runtime) mutate(ctx context.Context, req *admissionv1.AdmissionRequest) ([]byte, error) {
	if req.Operation != admissionv1.Create || req.Resource.Group != "" || req.Resource.Resource != "pods" || req.SubResource != "" || req.Namespace != r.config.Namespace {
		return nil, nil
	}
	var pod corev1.Pod
	if err := json.Unmarshal(req.Object.Raw, &pod); err != nil {
		return nil, err
	}
	if pod.Namespace == "" {
		pod.Namespace = req.Namespace
	}
	mode, err := r.discovery.kube.CoreV1().ConfigMaps(r.config.Namespace).Get(ctx, WebhookConfigMap, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if mode.Data["exportMode"] != "true" {
		return nil, nil
	}
	workloads, err := r.discovery.workloads(ctx)
	if err != nil {
		return nil, err
	}
	workload, err := r.discovery.podWorkload(ctx, &pod, workloads)
	if err != nil || workload == nil {
		return nil, err
	}
	volumes, err := r.discovery.volumes(ctx, &pod, workload)
	if err != nil || len(volumes) == 0 {
		return nil, err
	}
	key, err := r.discovery.kube.CoreV1().ConfigMaps(r.config.Namespace).Get(ctx, r.config.PublicKeyConfigMap, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	if key.Data[r.config.PublicKeyDataKey] == "" {
		return nil, fmt.Errorf("importer public key is empty")
	}
	return injectionPatch(&pod, volumes, r.config, workload.object.GetLabels()[DoguLabel])
}

// injectionPatch appends the exporter, read-only mounts and injection marker while preserving application containers and volumes.
func injectionPatch(pod *corev1.Pod, volumes []corev1.Volume, config Config, doguName string) ([]byte, error) {
	for _, container := range pod.Spec.Containers {
		if container.Name == SidecarName {
			if pod.Annotations[injectedAnnotation] == "true" {
				return nil, nil
			}
			return nil, fmt.Errorf("reserved sidecar name already exists")
		}
	}
	for _, volume := range pod.Spec.Volumes {
		if volume.Name == keyVolume {
			return nil, fmt.Errorf("reserved key volume already exists")
		}
	}
	mounts := []corev1.VolumeMount{{Name: keyVolume, MountPath: "/root/.ssh/authorized_keys", SubPath: config.PublicKeyDataKey, ReadOnly: true}}
	for _, volume := range volumes {
		mounts = append(mounts, corev1.VolumeMount{Name: volume.Name, MountPath: path.Join("/export", volume.PersistentVolumeClaim.ClaimName), ReadOnly: true})
	}
	container := corev1.Container{
		Name: SidecarName, Image: config.SidecarImage, ImagePullPolicy: corev1.PullIfNotPresent, VolumeMounts: mounts,
		Env:             []corev1.EnvVar{{Name: "DOGU_NAME", Value: doguName}},
		ReadinessProbe:  &corev1.Probe{ProbeHandler: corev1.ProbeHandler{Exec: &corev1.ExecAction{Command: []string{"sh", "-c", "nc -z 127.0.0.1 22"}}}, PeriodSeconds: 2, TimeoutSeconds: 1},
		SecurityContext: &corev1.SecurityContext{RunAsUser: ptr.To(int64(0)), RunAsNonRoot: ptr.To(false), Capabilities: &corev1.Capabilities{Drop: []corev1.Capability{"ALL"}, Add: []corev1.Capability{"DAC_OVERRIDE", "SYS_CHROOT", "NET_BIND_SERVICE", "SETGID", "SETUID"}}},
	}
	key := corev1.Volume{Name: keyVolume, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: config.PublicKeyConfigMap}, Items: []corev1.KeyToPath{{Key: config.PublicKeyDataKey, Path: config.PublicKeyDataKey}}}}}
	// Patching only appended items preserves application containers and mounts.
	patch := []patchOperation{{Op: "add", Path: "/spec/containers/-", Value: container}}
	if len(pod.Spec.Volumes) == 0 {
		patch = append(patch, patchOperation{Op: "add", Path: "/spec/volumes", Value: []corev1.Volume{key}})
	} else {
		patch = append(patch, patchOperation{Op: "add", Path: "/spec/volumes/-", Value: key})
	}
	if len(pod.Annotations) == 0 {
		patch = append(patch, patchOperation{Op: "add", Path: "/metadata/annotations", Value: map[string]string{injectedAnnotation: "true"}})
	} else {
		patch = append(patch, patchOperation{Op: "add", Path: "/metadata/annotations/k8s.cloudogu.com~1exporter-poc-injected", Value: "true"})
	}
	return json.Marshal(patch)
}
