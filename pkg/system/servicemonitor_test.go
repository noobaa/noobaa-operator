package system

import (
	"strings"
	"testing"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
)

func TestMgmtServiceMonitorKeepsReadyPodsOnly(t *testing.T) {
	t.Parallel()

	sm := util.KubeObject(bundle.File_deploy_internal_servicemonitor_mgmt_yaml).(*monitoringv1.ServiceMonitor)
	if len(sm.Spec.Endpoints) == 0 {
		t.Fatal("expected mgmt ServiceMonitor endpoints")
	}
	for i, ep := range sm.Spec.Endpoints {
		if ep.Path == "/metrics/bg_workers" {
			t.Fatal("mgmt ServiceMonitor must not scrape /metrics/bg_workers; that moved to the bg-workers monitor")
		}
		if !endpointHasKeepReadyPodRelabel(ep.RelabelConfigs) {
			t.Fatalf("endpoint %d path %q missing keep-ready relabel", i, ep.Path)
		}
	}
}

func TestBgWorkersServiceMonitorKeepsReadyPodsOnly(t *testing.T) {
	t.Parallel()

	sm := util.KubeObject(bundle.File_deploy_internal_servicemonitor_bg_workers_yaml).(*monitoringv1.ServiceMonitor)
	if len(sm.Spec.Endpoints) != 1 {
		t.Fatalf("expected 1 bg-workers ServiceMonitor endpoint, got %d", len(sm.Spec.Endpoints))
	}
	ep := sm.Spec.Endpoints[0]
	if ep.Port != "metrics" || ep.Path != "/metrics" {
		t.Fatalf("unexpected bg-workers endpoint port %q path %q", ep.Port, ep.Path)
	}
	if !endpointHasKeepReadyPodRelabel(ep.RelabelConfigs) {
		t.Fatal("bg-workers ServiceMonitor missing keep-ready relabel")
	}
}

func TestSetServiceMonitorKeepReadyPodRelabel(t *testing.T) {
	t.Parallel()

	r := &Reconciler{}
	managedBy := monitoringv1.RelabelConfig{
		TargetLabel: "managedBy",
		Replacement: strPtr("odf"),
	}
	keepReady := monitoringv1.RelabelConfig{
		SourceLabels: []monitoringv1.LabelName{podReadyMetaLabel},
		Action:       keepReadyRelabelAction,
		Regex:        keepReadyRelabelRegex,
	}
	endpoints := []monitoringv1.Endpoint{
		{},
		{RelabelConfigs: []monitoringv1.RelabelConfig{managedBy}},
		{RelabelConfigs: []monitoringv1.RelabelConfig{keepReady, managedBy}},
	}

	r.setServiceMonitorKeepReadyPodRelabel(endpoints)
	r.setServiceMonitorKeepReadyPodRelabel(endpoints) // idempotent

	for i, ep := range endpoints {
		if !endpointHasKeepReadyPodRelabel(ep.RelabelConfigs) {
			t.Fatalf("endpoint %d missing keep-ready relabel", i)
		}
		if i == 0 && len(ep.RelabelConfigs) != 1 {
			t.Fatalf("endpoint 0: got %d relabels, want 1", len(ep.RelabelConfigs))
		}
		if i > 0 {
			if len(ep.RelabelConfigs) != 2 {
				t.Fatalf("endpoint %d: got %d relabels, want 2 (keep-ready + existing)", i, len(ep.RelabelConfigs))
			}
			last := ep.RelabelConfigs[len(ep.RelabelConfigs)-1]
			if last.TargetLabel != "managedBy" {
				t.Fatalf("endpoint %d: existing relabel not preserved: %+v", i, last)
			}
		}
	}
}

func endpointHasKeepReadyPodRelabel(configs []monitoringv1.RelabelConfig) bool {
	for _, c := range configs {
		if !strings.EqualFold(c.Action, keepReadyRelabelAction) || c.Regex != keepReadyRelabelRegex {
			continue
		}
		if len(c.SourceLabels) == 1 && string(c.SourceLabels[0]) == podReadyMetaLabel {
			return true
		}
	}
	return false
}

func strPtr(s string) *string { return &s }
