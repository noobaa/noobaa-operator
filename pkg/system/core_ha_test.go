package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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

func TestDesiredCoreAffinity(t *testing.T) {
	t.Parallel()

	const coreName = "noobaa"
	zoneKey := "topology.kubernetes.io/zone"

	reconcilerFor := func(nb *nbv1.NooBaa) *Reconciler {
		return &Reconciler{
			NooBaa:  nb,
			Request: types.NamespacedName{Name: nb.Name},
		}
	}

	t.Run("HA off no affinity", func(t *testing.T) {
		t.Parallel()
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.DisableCoreHA = true
		if got := reconcilerFor(nb).desiredCoreAffinity(); got != nil {
			t.Fatalf("expected nil affinity, got %#v", got)
		}
	})

	t.Run("HA on no affinity hostname only", func(t *testing.T) {
		t.Parallel()
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		got := reconcilerFor(nb).desiredCoreAffinity()
		if got == nil || got.PodAntiAffinity == nil {
			t.Fatal("expected pod anti-affinity")
		}
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 1 {
			t.Fatalf("preferred terms = %d, want 1", len(prefs))
		}
		assertPreferredCoreTerm(t, prefs[0], coreName, corev1.LabelHostname)
	})

	t.Run("HA on with topologyKey zone", func(t *testing.T) {
		t.Parallel()
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.Affinity = &nbv1.AffinitySpec{TopologyKey: zoneKey}
		got := reconcilerFor(nb).desiredCoreAffinity()
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 2 {
			t.Fatalf("preferred terms = %d, want 2", len(prefs))
		}
		assertPreferredCoreTerm(t, prefs[0], coreName, corev1.LabelHostname)
		assertPreferredCoreTerm(t, prefs[1], coreName, zoneKey)
	})

	t.Run("HA on topologyKey hostname no duplicate", func(t *testing.T) {
		t.Parallel()
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.Affinity = &nbv1.AffinitySpec{TopologyKey: corev1.LabelHostname}
		got := reconcilerFor(nb).desiredCoreAffinity()
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 1 {
			t.Fatalf("preferred terms = %d, want 1", len(prefs))
		}
		assertPreferredCoreTerm(t, prefs[0], coreName, corev1.LabelHostname)
	})

	t.Run("HA on preserves nodeAffinity", func(t *testing.T) {
		t.Parallel()
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.Affinity = &nbv1.AffinitySpec{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "node-role.kubernetes.io/worker",
							Operator: corev1.NodeSelectorOpExists,
						}},
					}},
				},
			},
		}
		got := reconcilerFor(nb).desiredCoreAffinity()
		if got.NodeAffinity == nil {
			t.Fatal("expected nodeAffinity preserved")
		}
		if got.PodAntiAffinity == nil || len(got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution) != 1 {
			t.Fatal("expected injected hostname preferred term")
		}
		// Must not mutate the CR affinity object.
		if nb.Spec.Affinity.PodAntiAffinity != nil {
			t.Fatal("expected CR PodAntiAffinity to remain nil (deep copy)")
		}
	})

	t.Run("HA on existing preferred term no duplicate", func(t *testing.T) {
		t.Parallel()
		existing := corev1.WeightedPodAffinityTerm{
			Weight: 100,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: corev1.LabelHostname,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"noobaa-core": coreName},
				},
			},
		}
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.Affinity = &nbv1.AffinitySpec{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{existing},
			},
		}
		got := reconcilerFor(nb).desiredCoreAffinity()
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 1 {
			t.Fatalf("preferred terms = %d, want 1 (no duplicate)", len(prefs))
		}
	})

	t.Run("HA on stricter user selector still injects term", func(t *testing.T) {
		t.Parallel()
		// User term includes noobaa-core but also an extra label — must not suppress ours.
		userTerm := corev1.WeightedPodAffinityTerm{
			Weight: 50,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: corev1.LabelHostname,
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{
						"noobaa-core": coreName,
						"app":         "special",
					},
				},
			},
		}
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.Affinity = &nbv1.AffinitySpec{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{userTerm},
			},
		}
		got := reconcilerFor(nb).desiredCoreAffinity()
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 2 {
			t.Fatalf("preferred terms = %d, want 2 (user + generated)", len(prefs))
		}
		assertPreferredCoreTerm(t, prefs[1], coreName, corev1.LabelHostname)
	})

	t.Run("HA off preserves user podAntiAffinity", func(t *testing.T) {
		t.Parallel()
		userTerm := corev1.WeightedPodAffinityTerm{
			Weight: 50,
			PodAffinityTerm: corev1.PodAffinityTerm{
				TopologyKey: "topology.kubernetes.io/region",
				LabelSelector: &metav1.LabelSelector{
					MatchLabels: map[string]string{"app": "custom"},
				},
			},
		}
		nb := &nbv1.NooBaa{}
		nb.Name = coreName
		nb.Spec.DisableCoreHA = true
		nb.Spec.Affinity = &nbv1.AffinitySpec{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{userTerm},
			},
		}
		got := reconcilerFor(nb).desiredCoreAffinity()
		if got == nil || got.PodAntiAffinity == nil {
			t.Fatal("expected user podAntiAffinity")
		}
		prefs := got.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution
		if len(prefs) != 1 || prefs[0].Weight != 50 || prefs[0].PodAffinityTerm.TopologyKey != "topology.kubernetes.io/region" {
			t.Fatalf("user anti-affinity changed: %#v", prefs)
		}
	})
}

func assertPreferredCoreTerm(t *testing.T, term corev1.WeightedPodAffinityTerm, coreName, topologyKey string) {
	t.Helper()
	if term.Weight != 100 {
		t.Fatalf("weight = %d, want 100", term.Weight)
	}
	if term.PodAffinityTerm.TopologyKey != topologyKey {
		t.Fatalf("topologyKey = %q, want %q", term.PodAffinityTerm.TopologyKey, topologyKey)
	}
	if term.PodAffinityTerm.LabelSelector == nil || term.PodAffinityTerm.LabelSelector.MatchLabels["noobaa-core"] != coreName {
		t.Fatalf("labelSelector = %#v, want noobaa-core=%s", term.PodAffinityTerm.LabelSelector, coreName)
	}
}
