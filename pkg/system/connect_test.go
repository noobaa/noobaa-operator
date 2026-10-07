package system

import (
	"strings"
	"testing"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func podWithReady(name string, phase corev1.PodPhase, ready bool, deleting bool) corev1.Pod {
	status := corev1.ConditionFalse
	if ready {
		status = corev1.ConditionTrue
	}
	pod := corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: name},
		Status: corev1.PodStatus{
			Phase: phase,
			Conditions: []corev1.PodCondition{{
				Type:   corev1.PodReady,
				Status: status,
			}},
		},
	}
	if deleting {
		now := metav1.NewTime(time.Now())
		pod.DeletionTimestamp = &now
	}
	return pod
}

func TestPickReadyPodName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pods    []corev1.Pod
		want    string
		wantErr string
	}{
		{
			name:    "empty list",
			pods:    nil,
			wantErr: "no Ready mgmt/core pod",
		},
		{
			name: "only not ready",
			pods: []corev1.Pod{
				podWithReady("noobaa-core-0", corev1.PodRunning, false, false),
				podWithReady("noobaa-core-1", corev1.PodRunning, false, false),
			},
			wantErr: "no Ready mgmt/core pod",
		},
		{
			name: "ready standby skipped pending",
			pods: []corev1.Pod{
				podWithReady("noobaa-core-0", corev1.PodPending, false, false),
				podWithReady("noobaa-core-1", corev1.PodRunning, true, false),
			},
			want: "noobaa-core-1",
		},
		{
			name: "first ready in list order",
			pods: []corev1.Pod{
				podWithReady("noobaa-core-1", corev1.PodRunning, true, false),
				podWithReady("noobaa-core-0", corev1.PodRunning, true, false),
			},
			want: "noobaa-core-1",
		},
		{
			name: "deleting ready pod skipped",
			pods: []corev1.Pod{
				podWithReady("noobaa-core-0", corev1.PodRunning, true, true),
				podWithReady("noobaa-core-1", corev1.PodRunning, true, false),
			},
			want: "noobaa-core-1",
		},
		{
			name: "running but not ready skipped",
			pods: []corev1.Pod{
				podWithReady("noobaa-core-0", corev1.PodRunning, false, false),
				podWithReady("noobaa-core-1", corev1.PodRunning, true, false),
			},
			want: "noobaa-core-1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := pickReadyPodName(tt.pods)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("pickReadyPodName() = %q, want %q", got, tt.want)
			}
		})
	}
}
