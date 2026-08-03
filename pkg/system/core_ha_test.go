package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
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
