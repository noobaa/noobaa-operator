package system

import (
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestLookupProfile(t *testing.T) {
	tests := []struct {
		name     string
		profile  nbv1.PerformanceProfileType
		expected performanceProfile
	}{
		{
			name:     "default profile",
			profile:  nbv1.PerformanceProfileDefault,
			expected: performanceProfiles[nbv1.PerformanceProfileDefault],
		},
		{
			name:     "mixed-workload profile",
			profile:  nbv1.PerformanceProfileMixedWorkload,
			expected: performanceProfiles[nbv1.PerformanceProfileMixedWorkload],
		},
		{
			name:     "small-objects profile",
			profile:  nbv1.PerformanceProfileSmallObjects,
			expected: performanceProfiles[nbv1.PerformanceProfileSmallObjects],
		},
		{
			name:     "unset profile falls back to default",
			profile:  "",
			expected: performanceProfiles[nbv1.PerformanceProfileDefault],
		},
		{
			name:     "unknown profile falls back to default",
			profile:  nbv1.PerformanceProfileType("unknown"),
			expected: performanceProfiles[nbv1.PerformanceProfileDefault],
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: tt.profile,
				},
			}
			got := lookupProfile(nb)
			assertPerformanceProfile(t, got, tt.expected)
		})
	}
}

func TestGetCoreResources(t *testing.T) {
	explicit := profileResources("3", "3", "3Gi", "3Gi")

	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected corev1.ResourceRequirements
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileDefault].coreResources,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMixedWorkload].coreResources,
		},
		{
			name: "small-objects profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileSmallObjects].coreResources,
		},
		{
			name: "explicit core resources override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					CoreResources:      &explicit,
				},
			},
			expected: explicit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getCoreResources(tt.nb)
			assertResourceRequirements(t, got, tt.expected)
		})
	}
}

func TestGetLogResources(t *testing.T) {
	explicit := profileResources("300m", "300m", "300Mi", "300Mi")

	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected corev1.ResourceRequirements
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileDefault].logResources,
		},
		{
			name: "mini-env profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMiniEnv,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMiniEnv].logResources,
		},
		{
			name: "explicit log resources override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
					LogResources:       &explicit,
				},
			},
			expected: explicit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLogResources(tt.nb)
			assertResourceRequirements(t, got, tt.expected)
		})
	}
}

func TestGetDBResources(t *testing.T) {
	explicitCNPG := profileResources("5", "5", "5Gi", "5Gi")

	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected corev1.ResourceRequirements
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileDefault].dbResources,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMixedWorkload].dbResources,
		},
		{
			name: "small-objects profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileSmallObjects].dbResources,
		},
		{
			name: "dbSpec dbResources override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					DBSpec: &nbv1.NooBaaDBSpec{
						DBResources: &explicitCNPG,
					},
				},
			},
			expected: explicitCNPG,
		},
		{
			name: "legacy dbResources does not override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					DBResources:        &explicitCNPG,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileSmallObjects].dbResources,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDBResources(tt.nb)
			assertResourceRequirements(t, got, tt.expected)
		})
	}
}

func TestGetEndpointResources(t *testing.T) {
	explicit := profileResources("3", "3", "3Gi", "3Gi")

	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected corev1.ResourceRequirements
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileDefault].endpointResources,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMixedWorkload].endpointResources,
		},
		{
			name: "small-objects profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileSmallObjects].endpointResources,
		},
		{
			name: "explicit endpoint resources override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					Endpoints: &nbv1.EndpointsSpec{
						Resources: &explicit,
					},
				},
			},
			expected: explicit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEndpointResources(tt.nb)
			assertResourceRequirements(t, got, tt.expected)
		})
	}
}

func TestGetPVPoolResources(t *testing.T) {
	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected corev1.ResourceRequirements
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileDefault].pvPoolResources,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMixedWorkload].pvPoolResources,
		},
		{
			name: "mini profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMiniEnv,
				},
			},
			expected: performanceProfiles[nbv1.PerformanceProfileMiniEnv].pvPoolResources,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPVPoolResources(tt.nb)
			assertResourceRequirements(t, got, tt.expected)
		})
	}
}

func TestGetDBInstances(t *testing.T) {
	explicitInstances := 5

	tests := []struct {
		name     string
		nb       *nbv1.NooBaa
		expected int
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expected: 2,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expected: 2,
		},
		{
			name: "explicit dbSpec instances override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					DBSpec: &nbv1.NooBaaDBSpec{
						Instances: &explicitInstances,
					},
				},
			},
			expected: explicitInstances,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDBInstances(tt.nb)
			if got != tt.expected {
				t.Errorf("getDBInstances() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGetPVPoolNumVolumes(t *testing.T) {
	tests := []struct {
		name            string
		profile         nbv1.PerformanceProfileType
		existingVolumes int
		expected        int
	}{
		{
			name:            "new deployment with default profile",
			profile:         nbv1.PerformanceProfileDefault,
			existingVolumes: 0,
			expected:        3,
		},
		{
			name:            "new deployment with mixed-workload profile",
			profile:         nbv1.PerformanceProfileMixedWorkload,
			existingVolumes: 0,
			expected:        3,
		},
		{
			name:            "existing with default profile keeps current",
			profile:         nbv1.PerformanceProfileDefault,
			existingVolumes: 1,
			expected:        1,
		},
		{
			name:            "existing with unset profile keeps current",
			profile:         "",
			existingVolumes: 1,
			expected:        1,
		},
		{
			name:            "existing with mixed-workload profile increases",
			profile:         nbv1.PerformanceProfileMixedWorkload,
			existingVolumes: 1,
			expected:        3,
		},
		{
			name:            "existing with mixed-workload profile never decreases",
			profile:         nbv1.PerformanceProfileMixedWorkload,
			existingVolumes: 5,
			expected:        5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: tt.profile,
				},
			}
			got := getPVPoolNumVolumes(nb, tt.existingVolumes)
			if got != tt.expected {
				t.Errorf("getPVPoolNumVolumes() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGetEndpointMinMax(t *testing.T) {
	tests := []struct {
		name        string
		nb          *nbv1.NooBaa
		expectedMin int32
		expectedMax int32
	}{
		{
			name: "default profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileDefault,
				},
			},
			expectedMin: 1,
			expectedMax: 2,
		},
		{
			name: "mixed-workload profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
				},
			},
			expectedMin: 2,
			expectedMax: 4,
		},
		{
			name: "small-objects profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
				},
			},
			expectedMin: 2,
			expectedMax: 4,
		},
		{
			name: "explicit endpoint min/max override profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileSmallObjects,
					Endpoints: &nbv1.EndpointsSpec{
						MinCount: 5,
						MaxCount: 10,
					},
				},
			},
			expectedMin: 5,
			expectedMax: 10,
		},
		{
			name: "zero min count falls back to profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
					Endpoints: &nbv1.EndpointsSpec{
						MinCount: 0,
						MaxCount: 4,
					},
				},
			},
			expectedMin: 2,
			expectedMax: 4,
		},
		{
			name: "zero max count falls back to profile",
			nb: &nbv1.NooBaa{
				Spec: nbv1.NooBaaSpec{
					PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
					Endpoints: &nbv1.EndpointsSpec{
						MinCount: 2,
						MaxCount: 0,
					},
				},
			},
			expectedMin: 2,
			expectedMax: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotMin, gotMax := getEndpointMinMax(tt.nb)
			if gotMin != tt.expectedMin || gotMax != tt.expectedMax {
				t.Errorf("getEndpointMinMax() = (%d, %d), want (%d, %d)", gotMin, gotMax, tt.expectedMin, tt.expectedMax)
			}
		})
	}
}

func TestGetResourcesPrecedencePerComponent(t *testing.T) {
	explicitCore := profileResources("3", "3", "3Gi", "3Gi")
	explicitDB := profileResources("5", "5", "5Gi", "5Gi")
	explicitEndpoint := profileResources("7", "7", "7Gi", "7Gi")

	nb := &nbv1.NooBaa{
		Spec: nbv1.NooBaaSpec{
			PerformanceProfile: nbv1.PerformanceProfileMixedWorkload,
			CoreResources:      &explicitCore,
			DBSpec: &nbv1.NooBaaDBSpec{
				DBResources: &explicitDB,
			},
			Endpoints: &nbv1.EndpointsSpec{
				Resources: &explicitEndpoint,
			},
		},
	}

	assertResourceRequirements(t, getCoreResources(nb), explicitCore)
	assertResourceRequirements(t, getDBResources(nb), explicitDB)
	assertResourceRequirements(t, getEndpointResources(nb), explicitEndpoint)
}

func assertPerformanceProfile(t *testing.T, got, want performanceProfile) {
	t.Helper()
	assertResourceRequirements(t, got.coreResources, want.coreResources)
	assertResourceRequirements(t, got.logResources, want.logResources)
	assertResourceRequirements(t, got.dbResources, want.dbResources)
	assertResourceRequirements(t, got.endpointResources, want.endpointResources)
	assertResourceRequirements(t, got.pvPoolResources, want.pvPoolResources)
	if got.endpointMinCount != want.endpointMinCount {
		t.Errorf("endpointMinCount = %d, want %d", got.endpointMinCount, want.endpointMinCount)
	}
	if got.endpointMaxCount != want.endpointMaxCount {
		t.Errorf("endpointMaxCount = %d, want %d", got.endpointMaxCount, want.endpointMaxCount)
	}
	if got.dbInstances != want.dbInstances {
		t.Errorf("dbInstances = %d, want %d", got.dbInstances, want.dbInstances)
	}
	if got.pvPoolNumVolumes != want.pvPoolNumVolumes {
		t.Errorf("pvPoolNumVolumes = %d, want %d", got.pvPoolNumVolumes, want.pvPoolNumVolumes)
	}
}

func assertResourceRequirements(t *testing.T, got, want corev1.ResourceRequirements) {
	t.Helper()
	assertResourceList(t, got.Requests, want.Requests, "requests")
	assertResourceList(t, got.Limits, want.Limits, "limits")
}

func assertResourceList(t *testing.T, got, want corev1.ResourceList, listName string) {
	t.Helper()
	for name, wantQty := range want {
		gotQty, ok := got[name]
		if !ok {
			t.Errorf("%s missing resource %q", listName, name)
			continue
		}
		if gotQty.Cmp(wantQty) != 0 {
			t.Errorf("%s[%q] = %s, want %s", listName, name, gotQty.String(), wantQty.String())
		}
	}
	for name := range got {
		if _, ok := want[name]; !ok {
			qty := got[name]
			t.Errorf("%s has unexpected resource %q = %s", listName, name, qty.String())
		}
	}
}

func TestProfileResourcesValues(t *testing.T) {
	tests := []struct {
		name     string
		profile  nbv1.PerformanceProfileType
		core     [4]string
		log      [4]string
		db       [4]string
		endpoint [4]string
		pvPool   [4]string
	}{
		{
			name:     "default",
			profile:  nbv1.PerformanceProfileDefault,
			core:     [4]string{"500m", "1", "1Gi", "4Gi"},
			log:      [4]string{"200m", "200m", "500Mi", "500Mi"},
			db:       [4]string{"1", "1", "2Gi", "2Gi"},
			endpoint: [4]string{"500m", "2", "1Gi", "3Gi"},
			pvPool:   [4]string{"400m", "400m", "800Mi", "800Mi"},
		},
		{
			name:     "mixed-workload",
			profile:  nbv1.PerformanceProfileMixedWorkload,
			core:     [4]string{"1", "2", "2Gi", "4Gi"},
			log:      [4]string{"200m", "200m", "500Mi", "500Mi"},
			db:       [4]string{"4", "4", "8Gi", "8Gi"},
			endpoint: [4]string{"2", "4", "2Gi", "4Gi"},
			pvPool:   [4]string{"1", "1", "2Gi", "2Gi"},
		},
		{
			name:     "small-objects",
			profile:  nbv1.PerformanceProfileSmallObjects,
			core:     [4]string{"1", "2", "2Gi", "6Gi"},
			log:      [4]string{"200m", "200m", "500Mi", "500Mi"},
			db:       [4]string{"6", "6", "16Gi", "16Gi"},
			endpoint: [4]string{"1", "4", "2Gi", "4Gi"},
			pvPool:   [4]string{"1", "1", "2Gi", "2Gi"},
		},
		{
			name:     "mini-env",
			profile:  nbv1.PerformanceProfileMiniEnv,
			core:     [4]string{"100m", "100m", "1Gi", "1Gi"},
			log:      [4]string{"50m", "50m", "200Mi", "200Mi"},
			db:       [4]string{"100m", "100m", "500Mi", "500Mi"},
			endpoint: [4]string{"100m", "100m", "500Mi", "500Mi"},
			pvPool:   [4]string{"100m", "100m", "400Mi", "400Mi"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			profile := performanceProfiles[tt.profile]
			assertResourceQuantity(t, profile.coreResources, tt.core)
			assertResourceQuantity(t, profile.logResources, tt.log)
			assertResourceQuantity(t, profile.dbResources, tt.db)
			assertResourceQuantity(t, profile.endpointResources, tt.endpoint)
			assertResourceQuantity(t, profile.pvPoolResources, tt.pvPool)
		})
	}
}

func assertResourceQuantity(t *testing.T, resources corev1.ResourceRequirements, values [4]string) {
	t.Helper()
	cpuReq := resource.MustParse(values[0])
	cpuLim := resource.MustParse(values[1])
	memReq := resource.MustParse(values[2])
	memLim := resource.MustParse(values[3])

	if resources.Requests.Cpu().Cmp(cpuReq) != 0 {
		t.Errorf("cpu request = %s, want %s", resources.Requests.Cpu().String(), cpuReq.String())
	}
	if resources.Limits.Cpu().Cmp(cpuLim) != 0 {
		t.Errorf("cpu limit = %s, want %s", resources.Limits.Cpu().String(), cpuLim.String())
	}
	if resources.Requests.Memory().Cmp(memReq) != 0 {
		t.Errorf("memory request = %s, want %s", resources.Requests.Memory().String(), memReq.String())
	}
	if resources.Limits.Memory().Cmp(memLim) != 0 {
		t.Errorf("memory limit = %s, want %s", resources.Limits.Memory().String(), memLim.String())
	}
}

func TestAdjustCpuResourcesForIbmZ(t *testing.T) {
	tests := []struct {
		name       string
		cpuReq     string
		cpuLim     string
		factor     float64
		wantCPUReq string
		wantCPULim string
	}{
		{
			name:       "whole number cpu",
			cpuReq:     "2",
			cpuLim:     "4",
			factor:     0.5,
			wantCPUReq: "1",
			wantCPULim: "4",
		},
		{
			name:       "fractional cpu",
			cpuReq:     "1.5",
			cpuLim:     "2",
			factor:     0.5,
			wantCPUReq: "750m",
			wantCPULim: "2",
		},
		{
			name:       "even milli cpu",
			cpuReq:     "500m",
			cpuLim:     "1",
			factor:     0.5,
			wantCPUReq: "250m",
			wantCPULim: "1",
		},
		{
			name:       "odd milli cpu",
			cpuReq:     "501m",
			cpuLim:     "1",
			factor:     0.5,
			wantCPUReq: "250m", // int64 truncates 250.5
			wantCPULim: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := profileResources(tt.cpuReq, tt.cpuLim, "1Gi", "4Gi")
			got := adjustCpuResourcesForIbmZ(input, tt.factor)

			wantCPUReq := resource.MustParse(tt.wantCPUReq)
			wantCPULim := resource.MustParse(tt.wantCPULim)
			if got.Requests.Cpu().Cmp(wantCPUReq) != 0 {
				t.Errorf("cpu request = %s, want %s", got.Requests.Cpu().String(), wantCPUReq.String())
			}
			if got.Limits.Cpu().Cmp(wantCPULim) != 0 {
				t.Errorf("cpu limit = %s, want %s", got.Limits.Cpu().String(), wantCPULim.String())
			}
			if got.Requests.Memory().Cmp(*input.Requests.Memory()) != 0 {
				t.Errorf("memory request changed unexpectedly")
			}
			if got.Limits.Memory().Cmp(*input.Limits.Memory()) != 0 {
				t.Errorf("memory limit changed unexpectedly")
			}
		})
	}
}
