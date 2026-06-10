package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func boolPtr(v bool) *bool {
	return &v
}

func TestValidateAzureSTSParams(t *testing.T) {
	tests := []struct {
		name            string
		clientID        string
		resourcegroupID string
		tenantID        string
		subscriptionID  string
		wantOK          bool
	}{
		{
			name:   "none set",
			wantOK: true,
		},
		{
			name:            "all four non-empty",
			clientID:        "c",
			resourcegroupID: "rg",
			tenantID:        "t",
			subscriptionID:  "s",
			wantOK:          true,
		},
		{
			name:     "only clientID",
			clientID: "c",
			wantOK:   false,
		},
		{
			name:            "three of four",
			clientID:        "c",
			resourcegroupID: "rg",
			tenantID:        "t",
			wantOK:          false,
		},
		{
			name:            "two of four",
			clientID:        "c",
			resourcegroupID: "rg",
			wantOK:          false,
		},
		{
			name:            "whitespace-only strings count as set — all four",
			clientID:        " ",
			resourcegroupID: " ",
			tenantID:        " ",
			subscriptionID:  " ",
			wantOK:          true,
		},
		{
			name:     "one non-empty and rest empty",
			clientID: "x",
			wantOK:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validateAzureSTSParams(tt.clientID, tt.resourcegroupID, tt.tenantID, tt.subscriptionID)
			if got != tt.wantOK {
				t.Errorf("validateAzureSTSParams() = %v, want %v", got, tt.wantOK)
			}
		})
	}
}

func TestIsCoreHAEnabled(t *testing.T) {
	tests := []struct {
		name   string
		coreHA *bool
		want   bool
	}{
		{name: "nil is disabled", coreHA: nil, want: false},
		{name: "false is disabled", coreHA: boolPtr(false), want: false},
		{name: "true is enabled", coreHA: boolPtr(true), want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{CoreHA: tt.coreHA}}
			if got := isCoreHAEnabled(nb); got != tt.want {
				t.Errorf("isCoreHAEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsCoreLeaseConfiguredOnDeployed(t *testing.T) {
	tests := []struct {
		name     string
		leaseEnv string
		want     bool
	}{
		{name: "empty lease env", leaseEnv: "", want: false},
		{name: "lease env set", leaseEnv: "noobaa-core-lease", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := reconcilerWithCoreLeaseEnv(tt.leaseEnv)
			if got := isCoreLeaseConfiguredOnDeployed(r); got != tt.want {
				t.Errorf("isCoreLeaseConfiguredOnDeployed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetDesiredCoreReplicas(t *testing.T) {
	tests := []struct {
		name   string
		coreHA *bool
		want   int32
	}{
		{name: "HA disabled", coreHA: boolPtr(false), want: 1},
		{name: "HA enabled", coreHA: boolPtr(true), want: options.CoreHAReplicaCount},
		{name: "HA unset", coreHA: nil, want: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{CoreHA: tt.coreHA}}
			if got := getDesiredCoreReplicas(nb); got != tt.want {
				t.Errorf("getDesiredCoreReplicas() = %d, want %d", got, tt.want)
			}
		})
	}
}

func reconcilerWithCoreLeaseEnv(leaseEnv string) *Reconciler {
	return &Reconciler{
		NooBaa: &nbv1.NooBaa{
			Spec: nbv1.NooBaaSpec{},
		},
		CoreApp: &appsv1.StatefulSet{
			Spec: appsv1.StatefulSetSpec{
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{
							{Name: "db"},
							{
								Name: "core",
								Env: []corev1.EnvVar{
									{Name: "NOOBAA_CORE_LEASE_NAME", Value: leaseEnv},
								},
							},
						},
					},
				},
			},
		},
	}
}
