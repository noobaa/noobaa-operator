package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetDesiredCoreReplicas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		disableCoreHA bool
		want          int32
	}{
		{name: "default means on", disableCoreHA: false, want: options.CoreHAReplicaCount},
		{name: "disabled", disableCoreHA: true, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{}
			nb.Spec.DisableCoreHA = tt.disableCoreHA
			if got := getDesiredCoreReplicas(nb); got != tt.want {
				t.Fatalf("getDesiredCoreReplicas() = %d, want %d", got, tt.want)
			}
			if got := isCoreHAEnabled(nb); got != (tt.want == options.CoreHAReplicaCount) {
				t.Fatalf("isCoreHAEnabled() = %v, want %v", got, tt.want == options.CoreHAReplicaCount)
			}
		})
	}
}

func TestIsCorePodStale(t *testing.T) {
	t.Parallel()

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: "noobaa-core-1",
			Annotations: map[string]string{
				coreConfigMapHashAnnotation: "oldhash",
			},
			Labels: map[string]string{
				appsv1.ControllerRevisionHashLabelKey: "rev-old",
			},
		},
	}

	if !isCorePodStale(pod, "newhash", "rev-new") {
		t.Fatal("expected stale on hash mismatch")
	}
	pod.Annotations[coreConfigMapHashAnnotation] = "newhash"
	if !isCorePodStale(pod, "newhash", "rev-new") {
		t.Fatal("expected stale on revision mismatch")
	}
	pod.Labels[appsv1.ControllerRevisionHashLabelKey] = "rev-new"
	if isCorePodStale(pod, "newhash", "rev-new") {
		t.Fatal("expected not stale when hash and revision match")
	}
}
