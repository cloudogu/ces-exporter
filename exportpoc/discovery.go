// This file resolves V3 workload ownership and selects Helm-managed PVCs.

// Package exportpoc provides the opt-in V3 sidecar proof of concept.
package exportpoc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

// DoguLabel associates Helm-managed workloads with their V3 dogu.
const DoguLabel = "k8s.cloudogu.com/dogu.name"

// doguResource addresses the V3 API without changing the legacy typed Dogu client.
var doguResource = schema.GroupVersionResource{Group: "k8s.cloudogu.com", Version: "v3beta1", Resource: "dogus"}

// discovery shares uncached ownership and volume lookup between controller and webhook.
type discovery struct {
	kube      kubernetes.Interface
	dynamic   dynamic.Interface
	namespace string
}

// workload retains the controller identity and claim templates needed for pod attribution.
type workload struct {
	kind, name string
	object     metav1.Object
	template   corev1.PodTemplateSpec
	claims     []corev1.PersistentVolumeClaim
}

// workloads finds Deployments and StatefulSets belonging to V3 dogus through the shared dogu-name label.
func (d *discovery) workloads(ctx context.Context) ([]workload, error) {
	dogus, err := d.dynamic.Resource(doguResource).Namespace(d.namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list V3 dogus: %w", err)
	}
	names := map[string]bool{}
	for _, dogu := range dogus.Items {
		version, _, _ := unstructured.NestedString(dogu.Object, "spec", "doguApiVersion")
		name, _, _ := unstructured.NestedString(dogu.Object, "spec", "name")
		if version == "v3" {
			parts := strings.Split(name, "/")
			names[parts[len(parts)-1]] = true
		}
	}
	deployments, err := d.kube.AppsV1().Deployments(d.namespace).List(ctx, metav1.ListOptions{LabelSelector: DoguLabel})
	if err != nil {
		return nil, err
	}
	sets, err := d.kube.AppsV1().StatefulSets(d.namespace).List(ctx, metav1.ListOptions{LabelSelector: DoguLabel})
	if err != nil {
		return nil, err
	}
	result := []workload{}
	for i := range deployments.Items {
		w := &deployments.Items[i]
		if names[w.Labels[DoguLabel]] {
			result = append(result, workload{kind: "Deployment", name: w.Name, object: w, template: w.Spec.Template})
		}
	}
	for i := range sets.Items {
		w := &sets.Items[i]
		if names[w.Labels[DoguLabel]] {
			result = append(result, workload{kind: "StatefulSet", name: w.Name, object: w, template: w.Spec.Template, claims: w.Spec.VolumeClaimTemplates})
		}
	}
	return result, nil
}

// podWorkload follows controller references, including the ReplicaSet hop, and checks UIDs against the discovered workloads.
func (d *discovery) podWorkload(ctx context.Context, pod *corev1.Pod, workloads []workload) (*workload, error) {
	owner := metav1.GetControllerOf(pod)
	if owner == nil {
		return nil, nil
	}
	kind, name, uid := owner.Kind, owner.Name, owner.UID
	if kind == "ReplicaSet" {
		rs, err := d.kube.AppsV1().ReplicaSets(d.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if rs.UID != uid {
			return nil, fmt.Errorf("replicaset UID changed")
		}
		parent := metav1.GetControllerOf(rs)
		if parent == nil {
			return nil, nil
		}
		kind, name, uid = parent.Kind, parent.Name, parent.UID
	}
	for i := range workloads {
		w := &workloads[i]
		if w.kind == kind && w.name == name && w.object.GetUID() == uid {
			return w, nil
		}
	}
	return nil, nil
}

// releaseClaims uses the deployed Helm manifest as evidence of ownership. Merely
// mounting an external claim (existingClaim) must never make it exportable.
func (d *discovery) releaseClaims(ctx context.Context, w *workload) (map[string]bool, error) {
	annotations := w.object.GetAnnotations()
	release := annotations["meta.helm.sh/release-name"]
	if release == "" || annotations["meta.helm.sh/release-namespace"] != d.namespace {
		return nil, fmt.Errorf("workload %s has no local Helm release", w.name)
	}
	secrets, err := d.kube.CoreV1().Secrets(d.namespace).List(ctx, metav1.ListOptions{LabelSelector: "owner=helm,status=deployed,name=" + release})
	if err != nil {
		return nil, err
	}
	if len(secrets.Items) != 1 {
		return nil, fmt.Errorf("expected one deployed Helm release for %s, got %d", release, len(secrets.Items))
	}
	return claimsFromRelease(secrets.Items[0].Data["release"], d.namespace)
}

// claimsFromRelease decodes Helm secret storage and extracts standalone PVC names from the release manifest in the target namespace.
func claimsFromRelease(data []byte, namespace string) (map[string]bool, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("decode Helm release: %w", err)
	}
	if len(decoded) > 2 && decoded[0] == 0x1f && decoded[1] == 0x8b {
		reader, err := gzip.NewReader(bytes.NewReader(decoded))
		if err != nil {
			return nil, err
		}
		decoded, err = io.ReadAll(reader)
		_ = reader.Close()
		if err != nil {
			return nil, err
		}
	}
	var release struct {
		Manifest string `json:"manifest"`
	}
	if err = json.Unmarshal(decoded, &release); err != nil {
		return nil, err
	}
	claims := map[string]bool{}
	decoder := yaml.NewYAMLOrJSONDecoder(strings.NewReader(release.Manifest), 4096)
	for {
		var resource unstructured.Unstructured
		err = decoder.Decode(&resource)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if resource.GetKind() == "PersistentVolumeClaim" && (resource.GetNamespace() == "" || resource.GetNamespace() == namespace) {
			claims[resource.GetName()] = true
		}
	}
	return claims, nil
}

// volumes selects mounted claims created by the Helm release or StatefulSet claim templates, excluding external claims and duplicate mounts.
func (d *discovery) volumes(ctx context.Context, pod *corev1.Pod, w *workload) ([]corev1.Volume, error) {
	claims, err := d.releaseClaims(ctx, w)
	if err != nil {
		return nil, err
	}
	result := []corev1.Volume{}
	seen := map[string]bool{}
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		name := volume.PersistentVolumeClaim.ClaimName
		owned := claims[name]
		if w.kind == "StatefulSet" {
			for _, template := range w.claims {
				if name == template.Name+"-"+pod.Name {
					owned = true
				}
			}
		}
		if !owned || seen[name] {
			continue
		}
		pvc, err := d.kube.CoreV1().PersistentVolumeClaims(d.namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return nil, err
		}
		if pvc.Spec.VolumeMode != nil && *pvc.Spec.VolumeMode == corev1.PersistentVolumeBlock {
			return nil, fmt.Errorf("block PVC %s is unsupported", name)
		}
		seen[name] = true
		result = append(result, volume)
	}
	return result, nil
}
