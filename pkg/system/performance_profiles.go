package system

import (
	"github.com/sirupsen/logrus"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type performanceProfile struct {
	coreResources     corev1.ResourceRequirements
	dbResources       corev1.ResourceRequirements
	endpointResources corev1.ResourceRequirements
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
		dbResources:       profileResources("1", "1", "2Gi", "2Gi"),
		endpointResources: profileResources("500m", "2", "1Gi", "3Gi"),
		endpointMinCount:  1,
		endpointMaxCount:  2,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
	nbv1.PerformanceProfileMixedWorkload: {
		coreResources:     profileResources("1", "2", "2Gi", "4Gi"),
		dbResources:       profileResources("4", "4", "8Gi", "8Gi"),
		endpointResources: profileResources("2", "4", "2Gi", "4Gi"),
		endpointMinCount:  2,
		endpointMaxCount:  4,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
	nbv1.PerformanceProfileSmallObjects: {
		coreResources:     profileResources("1", "2", "2Gi", "6Gi"),
		dbResources:       profileResources("6", "6", "16Gi", "16Gi"),
		endpointResources: profileResources("2", "4", "2Gi", "4Gi"),
		endpointMinCount:  2,
		endpointMaxCount:  4,
		dbInstances:       2,
		pvPoolNumVolumes:  3,
	},
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
