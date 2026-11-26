// This file reconciles the ConfigMap trigger and replaces migration pods sequentially.

package exportpoc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"time"

	admissionregv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// errTriggerChanged represents a normal desired-state change, not an activation failure.
var errTriggerChanged = errors.New("export-mode trigger changed during reconciliation")

// desired reads the trigger and uses its UID to distinguish successive activations.
func (r *Runtime) desired(ctx context.Context) (bool, string, error) {
	cm, err := r.discovery.kube.CoreV1().ConfigMaps(r.config.Namespace).Get(ctx, TriggerConfigMap, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, "inactive", nil
	}
	if err != nil {
		return false, "", err
	}
	return true, string(cm.UID), nil
}

// control periodically reconciles the desired mode until the exporter shuts down.
func (r *Runtime) control(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	for {
		if err := r.reconcile(ctx); err != nil {
			slog.Error("PoC reconciliation failed", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// registrationReady requires a registered webhook with a CA and a ready serving endpoint before any pod replacement.
func (r *Runtime) registrationReady(ctx context.Context) error {
	configuration, err := r.discovery.kube.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(ctx, r.config.WebhookName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	for _, webhook := range configuration.Webhooks {
		svc := webhook.ClientConfig.Service
		if svc != nil && svc.Namespace == r.config.Namespace && svc.Path != nil && *svc.Path == "/mutate-pods" && len(webhook.ClientConfig.CABundle) > 0 && webhook.FailurePolicy != nil && *webhook.FailurePolicy == admissionregv1.Ignore {
			endpoints, err := r.discovery.kube.DiscoveryV1().EndpointSlices(r.config.Namespace).List(ctx, metav1.ListOptions{LabelSelector: "kubernetes.io/service-name=" + svc.Name})
			if err != nil {
				return err
			}
			for _, slice := range endpoints.Items {
				for _, endpoint := range slice.Endpoints {
					if endpoint.Conditions.Ready != nil && *endpoint.Conditions.Ready && len(endpoint.Addresses) > 0 {
						return nil
					}
				}
			}
		}
	}
	return fmt.Errorf("webhook registration, CA or serving endpoint not ready")
}

// reconcile publishes the mode before replacing pods and persists failures to prevent repeated destructive retries.
func (r *Runtime) reconcile(ctx context.Context) error {
	desired, transition, err := r.desired(ctx)
	if err != nil {
		return err
	}
	if err = r.registrationReady(ctx); err != nil {
		return err
	}
	cms := r.discovery.kube.CoreV1().ConfigMaps(r.config.Namespace)
	mode, err := cms.Get(ctx, WebhookConfigMap, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		mode = &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: WebhookConfigMap, Namespace: r.config.Namespace}, Data: map[string]string{}}
		mode, err = cms.Create(ctx, mode, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}
	if mode.Data == nil {
		mode.Data = map[string]string{}
	}
	if mode.Data["transition"] == transition && mode.Data["phase"] == "failed" {
		return nil
	}
	alreadyReady := mode.Data["transition"] == transition && mode.Data["exportMode"] == strconv.FormatBool(desired) && mode.Data["phase"] == "ready"
	mode.Data["exportMode"] = strconv.FormatBool(desired)
	mode.Data["transition"] = transition
	mode.Data["phase"] = "reconciling"
	delete(mode.Data, "error")
	if !alreadyReady {
		mode, err = cms.Update(ctx, mode, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	// All admission reads are uncached. The successful update above is the barrier
	// before any pod deletion, including deactivation.
	err = r.replacePods(ctx, desired, transition)
	if errors.Is(err, errTriggerChanged) {
		return nil
	}
	if err == nil && alreadyReady {
		return nil
	}
	if err != nil {
		mode.Data["phase"] = "failed"
		mode.Data["error"] = err.Error()
		r.event(ctx, "Warning", "ExportReplacementFailed", err.Error())
	} else {
		mode.Data["phase"] = "ready"
	}
	// Reread to avoid overwriting changes made while waiting for replacement pods.
	current, getErr := cms.Get(ctx, WebhookConfigMap, metav1.GetOptions{})
	if getErr != nil {
		return getErr
	}
	if current.Data["transition"] == transition {
		current.Data = mode.Data
		if _, updateErr := cms.Update(ctx, current, metav1.UpdateOptions{}); updateErr != nil {
			return updateErr
		}
	}
	return err
}

// replacePods replaces eligible pods one at a time, waiting for application and exporter readiness after each deletion.
func (r *Runtime) replacePods(ctx context.Context, desired bool, transition string) error {
	if desired {
		key, err := r.discovery.kube.CoreV1().ConfigMaps(r.config.Namespace).Get(ctx, r.config.PublicKeyConfigMap, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("importer key: %w", err)
		}
		if key.Data[r.config.PublicKeyDataKey] == "" {
			return fmt.Errorf("importer public key is empty")
		}
	}
	workloads, err := r.discovery.workloads(ctx)
	if err != nil {
		return err
	}
	pods, err := r.discovery.kube.CoreV1().Pods(r.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return err
	}
	sort.Slice(pods.Items, func(i, j int) bool { return pods.Items[i].Name < pods.Items[j].Name })
	for i := range pods.Items {
		pod := &pods.Items[i]
		if pod.Status.Phase == corev1.PodSucceeded || pod.Status.Phase == corev1.PodFailed {
			continue
		}
		w, err := r.discovery.podWorkload(ctx, pod, workloads)
		if err != nil {
			return err
		}
		if w == nil {
			continue
		}
		volumes, err := r.discovery.volumes(ctx, pod, w)
		if err != nil {
			return err
		}
		if len(volumes) == 0 {
			continue
		}
		if hasSidecar(pod) == desired && pod.DeletionTimestamp == nil {
			if err = r.waitForPod(ctx, pod, w, desired, transition, false); err != nil {
				return err
			}
			continue
		}
		_, current, err := r.desired(ctx)
		if err != nil {
			return err
		}
		if current != transition {
			return errTriggerChanged
		}
		slog.Info("replacing PoC pod", "pod", pod.Name, "exportMode", desired)
		uid := pod.UID
		err = r.discovery.kube.CoreV1().Pods(r.config.Namespace).Delete(ctx, pod.Name, metav1.DeleteOptions{Preconditions: &metav1.Preconditions{UID: &uid}})
		if err != nil && !apierrors.IsNotFound(err) {
			return err
		}
		if err = r.waitForPod(ctx, pod, w, desired, transition, true); err != nil {
			return err
		}
	}
	return nil
}

// waitForPod waits for the expected mode and readiness within the replacement timeout, stopping when the trigger changes.
func (r *Runtime) waitForPod(ctx context.Context, old *corev1.Pod, w *workload, desired bool, transition string, replaced bool) error {
	waitCtx, cancel := context.WithTimeout(ctx, r.config.ReplacementTimeout)
	defer cancel()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		_, current, err := r.desired(waitCtx)
		if err != nil {
			return err
		}
		if current != transition {
			return errTriggerChanged
		}
		pods, err := r.discovery.kube.CoreV1().Pods(r.config.Namespace).List(waitCtx, metav1.ListOptions{})
		if err != nil {
			return err
		}
		for i := range pods.Items {
			pod := &pods.Items[i]
			if pod.DeletionTimestamp != nil || hasSidecar(pod) != desired {
				continue
			}
			if replaced && pod.UID == old.UID {
				continue
			}
			if !replaced && pod.UID != old.UID {
				continue
			}
			if w.kind == "StatefulSet" && pod.Name != old.Name {
				continue
			}
			// For Deployments the controller ReplicaSet identifies the replacement.
			if !sameController(pod, old) {
				continue
			}
			if podReady(pod) {
				return nil
			}
		}
		select {
		case <-waitCtx.Done():
			return fmt.Errorf("waiting for %s replacement/readiness: %w", old.Name, waitCtx.Err())
		case <-ticker.C:
		}
	}
}

// sameController identifies pods managed by the same controller UID, including replacement StatefulSet pods.
func sameController(a, b *corev1.Pod) bool {
	ar, br := metav1.GetControllerOf(a), metav1.GetControllerOf(b)
	return ar != nil && br != nil && ar.UID == br.UID
}

// hasSidecar recognizes only a sidecar marked as injected by this PoC.
func hasSidecar(pod *corev1.Pod) bool {
	for _, container := range pod.Spec.Containers {
		if container.Name == SidecarName && pod.Annotations[injectedAnnotation] == "true" {
			return true
		}
	}
	return false
}

// podReady reports the Kubernetes PodReady condition, which includes the injected exporter probe.
func podReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// event records a Kubernetes event for operator-visible failures without hiding the original error.
func (r *Runtime) event(ctx context.Context, eventType, reason, message string) {
	_, err := r.discovery.kube.CoreV1().Events(r.config.Namespace).Create(ctx, &corev1.Event{ObjectMeta: metav1.ObjectMeta{GenerateName: "ces-exporter-poc-", Namespace: r.config.Namespace}, InvolvedObject: corev1.ObjectReference{Kind: "ConfigMap", Namespace: r.config.Namespace, Name: WebhookConfigMap}, Type: eventType, Reason: reason, Message: message, Source: corev1.EventSource{Component: "ces-exporter"}, FirstTimestamp: metav1.Now(), LastTimestamp: metav1.Now(), Count: 1}, metav1.CreateOptions{})
	if err != nil {
		slog.Error("record PoC event", "error", err)
	}
}
