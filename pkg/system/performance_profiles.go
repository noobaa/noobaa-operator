package system

import (
	"runtime"

	"github.com/sirupsen/logrus"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

const (
	ibmZCpuArch         = "s390x"
	ibmZCpuAdjustFactor = 0.2
)

type performanceProfile struct {
	coreResources     corev1.ResourceRequirements
	logResources      corev1.ResourceRequirements
	dbResources       corev1.ResourceRequirements
	endpointResources corev1.ResourceRequirements
	pvPoolResources   corev1.ResourceRequirements
	endpointMinCount  int32
	endpointMaxCount  int32
	dbInstances       int
	pvPoolNumVolumes  int
}

func profileResources(cpuReq, cpuLim, memReq, memLim string) corev1.ResourceRequirements {
	return corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuReq),
			corev1.ResourceMemory: resource.MustParse(memReq),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse(cpuLim),
			corev1.ResourceMemory: resource.MustParse(memLim),
		},
	}
}

var performanceProfiles = map[nbv1.PerformanceProfileType]performanceProfile{
	nbv1.PerformanceProfileDefault: {
		coreResources:     profileResources("500m", "1", "1Gi", "4Gi"),
		logResources:      profileResources("200m", "200m", "500Mi", "500Mi"),
		dbResources:       profileResources("1", "1", "2Gi", "2Gi"),
		endpointResources: profileResources("500m", "2", "1Gi", "3Gi"),
		pvPoolResources:   profileResources("400m", "400m", "800Mi", "800Mi"),
		endpointMinCount:  1,
		endpointMaxCount:  2,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
	nbv1.PerformanceProfileMixedWorkload: {
		coreResources:     profileResources("1", "2", "2Gi", "4Gi"),
		logResources:      profileResources("200m", "200m", "500Mi", "500Mi"),
		dbResources:       profileResources("4", "4", "8Gi", "8Gi"),
		endpointResources: profileResources("2", "4", "2Gi", "4Gi"),
		pvPoolResources:   profileResources("1", "1", "2Gi", "2Gi"),
		endpointMinCount:  2,
		endpointMaxCount:  4,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
	nbv1.PerformanceProfileSmallObjects: {
		coreResources:     profileResources("1", "2", "2Gi", "6Gi"),
		logResources:      profileResources("200m", "200m", "500Mi", "500Mi"),
		dbResources:       profileResources("6", "6", "16Gi", "16Gi"),
		// Endpoint CPU request is 1 (not 2): a single-process endpoint serving small
		// objects is main-thread bound and tops out at ~1 core, so request 2 leaves the
		// HPA reading ~50% utilisation on a fully-saturated pod and it never scales.
		// Request 1 makes a saturated pod read ~100%, restoring the CPU-based trigger.
		// Larger profiles keep request 2 because their per-pod ceiling is ~2.5 cores.
		endpointResources: profileResources("1", "4", "2Gi", "4Gi"),
		pvPoolResources:   profileResources("1", "1", "2Gi", "2Gi"),
		endpointMinCount:  2,
		endpointMaxCount:  4,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
	nbv1.PerformanceProfileDevEnv: {
		coreResources:     profileResources("500m", "500m", "1Gi", "1Gi"),
		logResources:      profileResources("200m", "200m", "500Mi", "500Mi"),
		dbResources:       profileResources("1", "1", "2Gi", "2Gi"),
		endpointResources: profileResources("500m", "500m", "500Mi", "500Mi"),
		pvPoolResources:   profileResources("500m", "500m", "500Mi", "500Mi"),
		endpointMinCount:  1,
		endpointMaxCount:  1,
		dbInstances:       1,
		pvPoolNumVolumes:  1,
	},
	nbv1.PerformanceProfileMiniEnv: {
		coreResources:     profileResources("100m", "100m", "1Gi", "1Gi"),
		logResources:      profileResources("50m", "50m", "200Mi", "200Mi"),
		dbResources:       profileResources("100m", "100m", "500Mi", "500Mi"),
		endpointResources: profileResources("100m", "100m", "500Mi", "500Mi"),
		pvPoolResources:   profileResources("100m", "100m", "400Mi", "400Mi"),
		endpointMinCount:  1,
		endpointMaxCount:  1,
		dbInstances:       1,
		pvPoolNumVolumes:  1,
	},
}

// On IBM Z (s390x), halve CPU requests for all profile-managed components once at
// startup, matching ocs-operator's IbmZCpuAdjustFactor.
// See https://redhat.atlassian.net/browse/RHSTOR-9067
func init() {
	if runtime.GOARCH != ibmZCpuArch {
		return
	}
	for name, profile := range performanceProfiles {
		profile.coreResources = adjustCpuResourcesForIbmZ(profile.coreResources, ibmZCpuAdjustFactor)
		profile.logResources = adjustCpuResourcesForIbmZ(profile.logResources, ibmZCpuAdjustFactor)
		profile.dbResources = adjustCpuResourcesForIbmZ(profile.dbResources, ibmZCpuAdjustFactor)
		profile.endpointResources = adjustCpuResourcesForIbmZ(profile.endpointResources, ibmZCpuAdjustFactor)
		profile.pvPoolResources = adjustCpuResourcesForIbmZ(profile.pvPoolResources, ibmZCpuAdjustFactor)
		performanceProfiles[name] = profile
	}
}

func lookupProfile(nb *nbv1.NooBaa) performanceProfile {
	profileType := nb.Spec.PerformanceProfile
	if profile, ok := performanceProfiles[profileType]; ok {
		return profile
	}
	if profileType != "" {
		logrus.Warnf("Unknown performanceProfile %q, falling back to %q", profileType, nbv1.PerformanceProfileDefault)
	}
	return performanceProfiles[nbv1.PerformanceProfileDefault]
}

func getCoreResources(nb *nbv1.NooBaa) corev1.ResourceRequirements {
	if nb.Spec.CoreResources != nil {
		return *nb.Spec.CoreResources
	}
	return lookupProfile(nb).coreResources
}

func getLogResources(nb *nbv1.NooBaa) corev1.ResourceRequirements {
	if nb.Spec.LogResources != nil {
		return *nb.Spec.LogResources
	}
	return lookupProfile(nb).logResources
}

func getDBResources(nb *nbv1.NooBaa) corev1.ResourceRequirements {
	if nb.Spec.DBSpec != nil && nb.Spec.DBSpec.DBResources != nil {
		return *nb.Spec.DBSpec.DBResources
	}
	return lookupProfile(nb).dbResources
}

func getEndpointResources(nb *nbv1.NooBaa) corev1.ResourceRequirements {
	if nb.Spec.Endpoints != nil && nb.Spec.Endpoints.Resources != nil {
		return *nb.Spec.Endpoints.Resources
	}
	return lookupProfile(nb).endpointResources
}

// adjustCpuResourcesForIbmZ multiplies CPU requests by adjustFactor (limits are unchanged).
func adjustCpuResourcesForIbmZ(rr corev1.ResourceRequirements, adjustFactor float64) corev1.ResourceRequirements {
	rrCopy := rr.DeepCopy()
	if rrCopy.Requests != nil {
		if cpuRequest, exists := rrCopy.Requests[corev1.ResourceCPU]; exists {
			rrCopy.Requests[corev1.ResourceCPU] = adjustCpu(cpuRequest, adjustFactor)
		}
	}
	return *rrCopy
}

func adjustCpu(cpuQty resource.Quantity, adjustFactor float64) resource.Quantity {
	return *resource.NewMilliQuantity(int64(float64(cpuQty.MilliValue())*adjustFactor), resource.DecimalSI)
}

func getDBInstances(nb *nbv1.NooBaa) int {
	if nb.Spec.DBSpec != nil && nb.Spec.DBSpec.Instances != nil {
		return *nb.Spec.DBSpec.Instances
	}
	return lookupProfile(nb).dbInstances
}

// getPVPoolNumVolumes determines the NumVolumes for the default pv-pool backingstore.
// Rules:
// - New deployment (existingVolumes <= 0): use profile value
// - Existing + "default" profile (or unset): keep current (no migration)
// - Existing + non-default profile: max(current, profile value) — never decrease
func getPVPoolNumVolumes(nb *nbv1.NooBaa, existingVolumes int) int {
	profile := lookupProfile(nb)
	if existingVolumes <= 0 {
		return profile.pvPoolNumVolumes
	}
	profileType := nb.Spec.PerformanceProfile
	if profileType == "" || profileType == nbv1.PerformanceProfileDefault {
		return existingVolumes
	}
	if profile.pvPoolNumVolumes > existingVolumes {
		return profile.pvPoolNumVolumes
	}
	return existingVolumes
}

func getEndpointMinMax(nb *nbv1.NooBaa) (int32, int32) {
	profile := lookupProfile(nb)
	minCount := profile.endpointMinCount
	maxCount := profile.endpointMaxCount
	if nb.Spec.Endpoints != nil {
		if nb.Spec.Endpoints.MinCount > 0 {
			minCount = nb.Spec.Endpoints.MinCount
		}
		if nb.Spec.Endpoints.MaxCount > 0 {
			maxCount = nb.Spec.Endpoints.MaxCount
		}
	}
	return minCount, maxCount
}

// GetPVPoolResources returns the default CPU/memory resources for PV pool
// backingstore pods based on the performance profile.
// In test env, returns minimal resources regardless of the profile.
func GetPVPoolResources(nb *nbv1.NooBaa) corev1.ResourceRequirements {
	if util.IsTestEnv() {
		return profileResources("50m", "50m", "200Mi", "200Mi")
	}
	return lookupProfile(nb).pvPoolResources
}
