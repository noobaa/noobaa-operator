package system

import (
	"testing"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/cnpg"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// TestRequiredSCCAnnotationForAllNooBaaDeployments verifies openshift.io/required-scc values
// documented in doc/scc-annotation.md under "For All NooBaa Deployments".
//
// Gaps (not covered by unit tests):
//   - noobaa-db-pg-cluster: openshift.io/required-scc is set on Cluster.Spec.InheritedMetadata
//     in reconcileDBCluster() during new-cluster creation (pkg/system/db_reconciler.go).
//   - noobaa-db-pg-cluster-init Job: inherits the same annotation from InheritedMetadata via CNPG.
//
// These require integration/envtest or a fake client to exercise reconcileDBCluster() without
// duplicating annotation logic in tests.
func TestRequiredSCCAnnotationForAllNooBaaDeployments(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		check    func(t *testing.T) map[string]string
	}{
		{
			name:     "noobaa-core StatefulSet",
			expected: "noobaa-core",
			check: func(*testing.T) map[string]string {
				coreStatefulSet := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
				return coreStatefulSet.Spec.Template.Annotations
			},
		},
		{
			name:     "noobaa-endpoint Deployment",
			expected: "noobaa-endpoint",
			check: func(*testing.T) map[string]string {
				endpointDeployment := util.KubeObject(bundle.File_deploy_internal_deployment_endpoint_yaml).(*appsv1.Deployment)
				return endpointDeployment.Spec.Template.Annotations
			},
		},
		{
			name:     "noobaa-agent Pod",
			expected: "noobaa-agent",
			check: func(*testing.T) map[string]string {
				agentPod := util.KubeObject(bundle.File_deploy_internal_pod_agent_yaml).(*corev1.Pod)
				return agentPod.Annotations
			},
		},
		{
			name:     "noobaa-operator Deployment",
			expected: "restricted-v2",
			check: func(*testing.T) map[string]string {
				operatorDeployment := util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)
				return operatorDeployment.Spec.Template.Annotations
			},
		},
		{
			name:     "cnpg-controller-manager Deployment",
			expected: "restricted-v2",
			check: func(t *testing.T) map[string]string {
				cnpgResources, err := cnpg.LoadCnpgResources()
				if err != nil {
					t.Fatalf("LoadCnpgResources() error: %v", err)
				}
				return cnpgResources.CnpgOperatorDeployment.Spec.Template.Annotations
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			assertRequiredSCCAnnotation(t, testCase.check(t), testCase.expected)
		})
	}
}

func assertRequiredSCCAnnotation(t *testing.T, annotations map[string]string, expected string) {
	t.Helper()
	got, ok := annotations[secv1.RequiredSCCAnnotation]
	if !ok {
		t.Fatalf("missing annotation %q", secv1.RequiredSCCAnnotation)
	}
	if got != expected {
		t.Fatalf("annotation %q = %q, want %q", secv1.RequiredSCCAnnotation, got, expected)
	}
}
