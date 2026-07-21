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

func staleCorePod(name string, ready bool, deleting bool) corev1.Pod {
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
			Annotations: map[string]string{
				coreConfigMapHashAnnotation: "oldhash",
			},
			Labels: map[string]string{
				appsv1.ControllerRevisionHashLabelKey: "rev-old",
			},
		},
	}
	if ready {
		pod.Status.Conditions = []corev1.PodCondition{{
			Type:   corev1.PodReady,
			Status: corev1.ConditionTrue,
		}}
	}
	if deleting {
		now := metav1.Now()
		pod.DeletionTimestamp = &now
	}
	return pod
}

func TestPickStaleCorePodToDelete(t *testing.T) {
	t.Parallel()

	const desiredHash = "newhash"
	const updateRevision = "rev-new"

	tests := []struct {
		name string
		pods []corev1.Pod
		want string // empty means nil
	}{
		{
			name: "prefer standby when both stale and leader listed first",
			pods: []corev1.Pod{
				staleCorePod("noobaa-core-0", true, false),
				staleCorePod("noobaa-core-1", false, false),
			},
			want: "noobaa-core-1",
		},
		{
			name: "prefer standby when both stale and standby listed first",
			pods: []corev1.Pod{
				staleCorePod("noobaa-core-1", false, false),
				staleCorePod("noobaa-core-0", true, false),
			},
			want: "noobaa-core-1",
		},
		{
			name: "delete leader when only leader is stale",
			pods: []corev1.Pod{
				staleCorePod("noobaa-core-0", true, false),
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "noobaa-core-1",
						Annotations: map[string]string{
							coreConfigMapHashAnnotation: desiredHash,
						},
						Labels: map[string]string{
							appsv1.ControllerRevisionHashLabelKey: updateRevision,
						},
					},
				},
			},
			want: "noobaa-core-0",
		},
		{
			name: "wait when any pod is deleting",
			pods: []corev1.Pod{
				staleCorePod("noobaa-core-0", false, true),
				staleCorePod("noobaa-core-1", true, false),
			},
			want: "",
		},
		{
			name: "none stale",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "noobaa-core-0",
						Annotations: map[string]string{
							coreConfigMapHashAnnotation: desiredHash,
						},
						Labels: map[string]string{
							appsv1.ControllerRevisionHashLabelKey: updateRevision,
						},
					},
					Status: corev1.PodStatus{
						Conditions: []corev1.PodCondition{{
							Type:   corev1.PodReady,
							Status: corev1.ConditionTrue,
						}},
					},
				},
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := pickStaleCorePodToDelete(tt.pods, desiredHash, updateRevision)
			if tt.want == "" {
				if got != nil {
					t.Fatalf("pickStaleCorePodToDelete() = %q, want nil", got.Name)
				}
				return
			}
			if got == nil {
				t.Fatalf("pickStaleCorePodToDelete() = nil, want %q", tt.want)
			}
			if got.Name != tt.want {
				t.Fatalf("pickStaleCorePodToDelete() = %q, want %q", got.Name, tt.want)
			}
		})
	}
}
