// This file verifies volume selection, admission patches and export-mode transitions with fake clients.

package exportpoc

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	jsonpatch "github.com/evanphx/json-patch/v5"
	"github.com/stretchr/testify/require"
	admissionv1 "k8s.io/api/admission/v1"
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	ktesting "k8s.io/client-go/testing"
	"k8s.io/utils/ptr"
)

// releaseSecret builds a Helm-compatible compressed release containing one owned PVC for the discovery tests.
func releaseSecret(t *testing.T) *corev1.Secret {
	t.Helper()
	release, err := json.Marshal(map[string]string{"manifest": "apiVersion: v1\nkind: PersistentVolumeClaim\nmetadata:\n  name: data\n"})
	require.NoError(t, err)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	_, err = writer.Write(release)
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "release", Namespace: "test", Labels: map[string]string{"owner": "helm", "status": "deployed", "name": "test"}}, Data: map[string][]byte{"release": []byte(base64.StdEncoding.EncodeToString(compressed.Bytes()))}}
}

// fixture creates a V3 dogu, its Deployment and pod, an external claim reference and a ready webhook registration.
func fixture(t *testing.T, active bool) (*Runtime, *fake.Clientset, *corev1.Pod) {
	t.Helper()
	deployment := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "app", Namespace: "test", UID: "deployment", Labels: map[string]string{DoguLabel: "example"}, Annotations: map[string]string{"meta.helm.sh/release-name": "test", "meta.helm.sh/release-namespace": "test"}}}
	rs := &appsv1.ReplicaSet{ObjectMeta: metav1.ObjectMeta{Name: "app-rs", Namespace: "test", UID: "replicaset", OwnerReferences: []metav1.OwnerReference{{Kind: "Deployment", Name: "app", UID: deployment.UID, Controller: ptr.To(true)}}}}
	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "app-old", Namespace: "test", UID: "old", OwnerReferences: []metav1.OwnerReference{{Kind: "ReplicaSet", Name: rs.Name, UID: rs.UID, Controller: ptr.To(true)}}}, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "application", Image: "app"}}, Volumes: []corev1.Volume{{Name: "app-data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "data"}}}, {Name: "external", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "external"}}}}}, Status: corev1.PodStatus{Conditions: []corev1.PodCondition{{Type: corev1.PodReady, Status: corev1.ConditionTrue}}}}
	webhook := &admissionregv1.MutatingWebhookConfiguration{ObjectMeta: metav1.ObjectMeta{Name: "webhook"}, Webhooks: []admissionregv1.MutatingWebhook{{FailurePolicy: ptr.To(admissionregv1.Ignore), ClientConfig: admissionregv1.WebhookClientConfig{CABundle: []byte("ca"), Service: &admissionregv1.ServiceReference{Namespace: "test", Name: "webhook", Path: ptr.To("/mutate-pods")}}}}}
	endpoints := &discoveryv1.EndpointSlice{ObjectMeta: metav1.ObjectMeta{Name: "webhook", Namespace: "test", Labels: map[string]string{"kubernetes.io/service-name": "webhook"}}, AddressType: discoveryv1.AddressTypeIPv4, Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"127.0.0.1"}, Conditions: discoveryv1.EndpointConditions{Ready: ptr.To(true)}}}}
	key := &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "key", Namespace: "test"}, Data: map[string]string{"publicKey": "ssh-ed25519 test"}}
	objects := []runtime.Object{deployment, rs, pod, webhook, endpoints, key, releaseSecret(t), &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "data", Namespace: "test"}}}
	if active {
		objects = append(objects, &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: TriggerConfigMap, Namespace: "test", UID: "activation"}})
	}
	kube := fake.NewClientset(objects...)
	dogu := &unstructured.Unstructured{Object: map[string]any{"apiVersion": "k8s.cloudogu.com/v3beta1", "kind": "Dogu", "metadata": map[string]any{"name": "example", "namespace": "test"}, "spec": map[string]any{"name": "example", "doguApiVersion": "v3"}}}
	return newFixtureRuntime(kube, dogu), kube, pod
}

// newFixtureRuntime connects fake clients to the PoC with a short replacement timeout for deterministic tests.
func newFixtureRuntime(kube *fake.Clientset, dogu *unstructured.Unstructured) *Runtime {
	dyn := dynamicfake.NewSimpleDynamicClient(runtime.NewScheme(), dogu)
	return &Runtime{config: Config{Namespace: "test", SidecarImage: "exporter:test", PublicKeyConfigMap: "key", PublicKeyDataKey: "publicKey", WebhookName: "webhook", ReplacementTimeout: 30 * time.Millisecond}, discovery: &discovery{kube: kube, dynamic: dyn, namespace: "test"}}
}

// TestAdmissionMountsOnlyOwnedClaimsReadOnly verifies read-only owned mounts, preserved application containers and idempotent admission.
func TestAdmissionMountsOnlyOwnedClaimsReadOnly(t *testing.T) {
	r, kube, pod := fixture(t, true)
	_, err := kube.CoreV1().ConfigMaps("test").Create(t.Context(), &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: WebhookConfigMap, Namespace: "test"}, Data: map[string]string{"exportMode": "true"}}, metav1.CreateOptions{})
	require.NoError(t, err)
	original, err := json.Marshal(pod)
	require.NoError(t, err)
	request := &admissionv1.AdmissionRequest{Operation: admissionv1.Create, Namespace: "test", Resource: metav1.GroupVersionResource{Version: "v1", Resource: "pods"}, Object: runtime.RawExtension{Raw: original}}
	patch, err := r.mutate(t.Context(), request)
	require.NoError(t, err)
	decoded, err := jsonpatch.DecodePatch(patch)
	require.NoError(t, err)
	mutated, err := decoded.Apply(original)
	require.NoError(t, err)
	var result corev1.Pod
	require.NoError(t, json.Unmarshal(mutated, &result))
	require.Equal(t, pod.Spec.Containers[0], result.Spec.Containers[0])
	require.Len(t, result.Spec.Containers, 2)
	exporter := result.Spec.Containers[1]
	require.Equal(t, SidecarName, exporter.Name)
	require.Len(t, exporter.VolumeMounts, 2)
	require.Equal(t, "/export/data", exporter.VolumeMounts[1].MountPath)
	for _, mount := range exporter.VolumeMounts {
		require.True(t, mount.ReadOnly)
		require.NotEqual(t, "external", mount.Name)
	}
	require.Equal(t, pod.Spec.Volumes, result.Spec.Volumes[:len(pod.Spec.Volumes)])
	request.Object.Raw = mutated
	patch, err = r.mutate(t.Context(), request)
	require.NoError(t, err)
	require.Empty(t, patch)
	for _, mode := range []string{"false", ""} {
		cm, err := kube.CoreV1().ConfigMaps("test").Get(t.Context(), WebhookConfigMap, metav1.GetOptions{})
		require.NoError(t, err)
		cm.Data["exportMode"] = mode
		_, err = kube.CoreV1().ConfigMaps("test").Update(t.Context(), cm, metav1.UpdateOptions{})
		require.NoError(t, err)
		request.Object.Raw = original
		patch, err = r.mutate(t.Context(), request)
		require.NoError(t, err)
		require.Empty(t, patch)
	}
}

// TestV2WorkloadsAreExcluded ensures that the new controller cannot select workloads belonging to a V2 dogu.
func TestV2WorkloadsAreExcluded(t *testing.T) {
	r, _, _ := fixture(t, true)
	dogu, err := r.discovery.dynamic.Resource(doguResource).Namespace("test").Get(t.Context(), "example", metav1.GetOptions{})
	require.NoError(t, err)
	require.NoError(t, unstructured.SetNestedField(dogu.Object, "v2", "spec", "doguApiVersion"))
	_, err = r.discovery.dynamic.Resource(doguResource).Namespace("test").Update(t.Context(), dogu, metav1.UpdateOptions{})
	require.NoError(t, err)
	workloads, err := r.discovery.workloads(t.Context())
	require.NoError(t, err)
	require.Empty(t, workloads)
}

// TestReconciliationOrdersConfigBeforeDeletionAndIsRepeatable checks the mode-update barrier, UID preconditions and absence of repeated replacements.
func TestReconciliationOrdersConfigBeforeDeletionAndIsRepeatable(t *testing.T) {
	for _, active := range []bool{true, false} {
		t.Run(fmt.Sprint(active), func(t *testing.T) {
			r, kube, pod := fixture(t, active)
			if !active {
				pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: SidecarName})
				pod.Annotations = map[string]string{injectedAnnotation: "true"}
				_, err := kube.CoreV1().Pods("test").Update(t.Context(), pod, metav1.UpdateOptions{})
				require.NoError(t, err)
			}
			deletes := 0
			kube.PrependReactor("delete", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
				deletes++
				object, err := kube.Tracker().Get(corev1.SchemeGroupVersion.WithResource("configmaps"), "test", WebhookConfigMap)
				require.NoError(t, err)
				cm := object.(*corev1.ConfigMap)
				require.Equal(t, fmt.Sprint(active), cm.Data["exportMode"])
				deletion := action.(ktesting.DeleteAction)
				require.Equal(t, pod.UID, *deletion.GetDeleteOptions().Preconditions.UID)
				replacement := pod.DeepCopy()
				replacement.Name = "app-new"
				replacement.UID = "new"
				replacement.Annotations = nil
				replacement.Spec.Containers = pod.Spec.Containers[:1]
				if active {
					replacement.Spec.Containers = append(replacement.Spec.Containers, corev1.Container{Name: SidecarName})
					replacement.Annotations = map[string]string{injectedAnnotation: "true"}
				}
				require.NoError(t, kube.Tracker().Delete(corev1.SchemeGroupVersion.WithResource("pods"), "test", pod.Name))
				require.NoError(t, kube.Tracker().Create(corev1.SchemeGroupVersion.WithResource("pods"), replacement, "test"))
				return true, nil, nil
			})
			require.NoError(t, r.reconcile(t.Context()))
			require.Equal(t, 1, deletes)
			require.NoError(t, r.reconcile(t.Context()))
			require.Equal(t, 1, deletes)
			cm, err := kube.CoreV1().ConfigMaps("test").Get(t.Context(), WebhookConfigMap, metav1.GetOptions{})
			require.NoError(t, err)
			require.Equal(t, "ready", cm.Data["phase"])
		})
	}
}

// TestFailedReplacementIsLatchedAcrossReconciliation ensures that a timed-out replacement is persisted instead of causing a restart loop.
func TestFailedReplacementIsLatchedAcrossReconciliation(t *testing.T) {
	r, kube, _ := fixture(t, true)
	require.Error(t, r.reconcile(t.Context()))
	cm, err := kube.CoreV1().ConfigMaps("test").Get(t.Context(), WebhookConfigMap, metav1.GetOptions{})
	require.NoError(t, err)
	require.Equal(t, "failed", cm.Data["phase"])
	require.NoError(t, r.reconcile(context.Background()))
	deletes := 0
	for _, action := range kube.Actions() {
		if action.GetVerb() == "delete" && action.GetResource().Resource == "pods" {
			deletes++
		}
	}
	require.Equal(t, 1, deletes)
}

// TestStatefulSetClaimsAreExportedButExternalClaimsAreNot verifies claim-template ownership without including external claims.
func TestStatefulSetClaimsAreExportedButExternalClaimsAreNot(t *testing.T) {
	r, _, pod := fixture(t, true)
	pod.Name = "database-0"
	pod.Spec.Volumes = append(pod.Spec.Volumes, corev1.Volume{Name: "db", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "db-database-0"}}})
	_, err := r.discovery.kube.CoreV1().PersistentVolumeClaims("test").Create(t.Context(), &corev1.PersistentVolumeClaim{ObjectMeta: metav1.ObjectMeta{Name: "db-database-0", Namespace: "test"}}, metav1.CreateOptions{})
	require.NoError(t, err)
	workloads, err := r.discovery.workloads(t.Context())
	require.NoError(t, err)
	w := workloads[0]
	w.kind = "StatefulSet"
	w.claims = []corev1.PersistentVolumeClaim{{ObjectMeta: metav1.ObjectMeta{Name: "db"}}}
	volumes, err := r.discovery.volumes(t.Context(), pod, &w)
	require.NoError(t, err)
	require.Len(t, volumes, 2)
	require.Equal(t, "db-database-0", volumes[1].PersistentVolumeClaim.ClaimName)
}

// TestTriggerChangeDoesNotReportFailure verifies that a normal mode switch interrupts reconciliation without a failure event.
func TestTriggerChangeDoesNotReportFailure(t *testing.T) {
	r, kube, _ := fixture(t, true)
	kube.PrependReactor("delete", "pods", func(action ktesting.Action) (bool, runtime.Object, error) {
		require.NoError(t, kube.Tracker().Delete(corev1.SchemeGroupVersion.WithResource("configmaps"), "test", TriggerConfigMap))
		return false, nil, nil
	})
	require.NoError(t, r.reconcile(t.Context()))
	for _, action := range kube.Actions() {
		require.False(t, action.GetVerb() == "create" && action.GetResource().Resource == "events", "a normal mode change must not generate a failure event")
	}
}
