package system

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	networkingv1 "k8s.io/api/networking/v1"
)

// ReconcileNetworkPolicies reconciles network policies for NooBaa operands.
// Operator and CNPG controller policies are Phase 2 scope (ODF 5.1).
func (r *Reconciler) ReconcileNetworkPolicies() error {

	if err := r.ReconcileObject(r.NetworkPolicyCore, r.SetDesiredNetworkPolicyCore); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.NetworkPolicyEndpoint, r.SetDesiredNetworkPolicyEndpoint); err != nil {
		return err
	}

	if err := r.ReconcileObject(r.NetworkPolicyPVPool, r.SetDesiredNetworkPolicyPVPool); err != nil {
		return err
	}

	if r.shouldReconcileCNPGCluster() {
		if err := r.ReconcileObject(r.NetworkPolicyDb, r.SetDesiredNetworkPolicyDb); err != nil {
			return err
		}
	}

	return nil
}

// SetDesiredNetworkPolicyCore updates the core network policy pod selector
// to match the actual NooBaa system name.
// The cross-namespace rule allows TCP 8443 (mgmt-https) from any namespace because:
// - Prometheus/monitoring scrapes metrics on this port (namespace varies by platform)
// - Ingress controllers route external RouteMgmt traffic to this port
func (r *Reconciler) SetDesiredNetworkPolicyCore() error {
	r.NetworkPolicyCore.Spec.PodSelector.MatchLabels["noobaa-core"] = r.Request.Name
	return nil
}

// SetDesiredNetworkPolicyEndpoint sets the endpoint network policy spec from the bundle,
// so bundle changes also reach existing policies, and updates the pod selector
// to match the actual NooBaa system name
func (r *Reconciler) SetDesiredNetworkPolicyEndpoint() error {
	desired := util.KubeObject(bundle.File_deploy_internal_networkpolicy_endpoint_yaml).(*networkingv1.NetworkPolicy)
	r.NetworkPolicyEndpoint.Spec = desired.Spec
	r.NetworkPolicyEndpoint.Spec.PodSelector.MatchLabels["noobaa-s3"] = r.Request.Name
	return nil
}

// SetDesiredNetworkPolicyPVPool sets the pv-pool network policy spec from the bundle,
// so bundle changes also reach existing policies
func (r *Reconciler) SetDesiredNetworkPolicyPVPool() error {
	desired := util.KubeObject(bundle.File_deploy_internal_networkpolicy_pvpool_yaml).(*networkingv1.NetworkPolicy)
	r.NetworkPolicyPVPool.Spec = desired.Spec
	return nil
}

// SetDesiredNetworkPolicyDb updates the DB network policy pod selector
// to match the actual CNPG cluster name
func (r *Reconciler) SetDesiredNetworkPolicyDb() error {
	r.NetworkPolicyDb.Spec.PodSelector.MatchLabels["cnpg.io/cluster"] = r.CNPGCluster.Name
	return nil
}
