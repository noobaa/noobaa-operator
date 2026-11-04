package system

import (
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetBgExecutorsReplicas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		spec *nbv1.BgExecutorsSpec
		want int32
	}{
		{name: "nil defaults to 1", spec: nil, want: 1},
		{name: "empty defaults to 1", spec: &nbv1.BgExecutorsSpec{}, want: 1},
		{name: "zero defaults to 1", spec: &nbv1.BgExecutorsSpec{Replicas: 0}, want: 1},
		{name: "negative defaults to 1", spec: &nbv1.BgExecutorsSpec{Replicas: -2}, want: 1},
		{name: "explicit replicas", spec: &nbv1.BgExecutorsSpec{Replicas: 3}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{BgExecutors: tt.spec}}
			if got := getBgExecutorsReplicas(nb); got != tt.want {
				t.Fatalf("getBgExecutorsReplicas() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBgExecutorsCommandRunsBgWorkers(t *testing.T) {
	t.Parallel()

	// Repo-layout parity: executors run bg_workers.js with BG_WORKERS_MODE=executor.
	deploy := util.KubeObject(bundle.File_deploy_internal_deployment_bg_executors_yaml).(*appsv1.Deployment)
	const wantScript = "/root/node_modules/noobaa-core/src/server/bg_workers.js"
	cmd := deploy.Spec.Template.Spec.Containers[0].Command
	if len(cmd) < 3 || cmd[2] != wantScript {
		t.Fatalf("bg-executors command script = %v, want … %s", cmd, wantScript)
	}
	if deploy.Spec.Template.Spec.Containers[0].Name != "bg-executors" {
		t.Fatalf("expected container name bg-executors, got %q", deploy.Spec.Template.Spec.Containers[0].Name)
	}
	mode := ""
	for _, e := range deploy.Spec.Template.Spec.Containers[0].Env {
		if e.Name == "BG_WORKERS_MODE" {
			mode = e.Value
		}
	}
	if mode != "executor" {
		t.Fatalf("bundle BG_WORKERS_MODE = %q, want executor", mode)
	}
	if deploy.Spec.Template.Spec.ServiceAccountName != "noobaa-bg-executors" {
		t.Fatalf("serviceAccountName = %q, want noobaa-bg-executors", deploy.Spec.Template.Spec.ServiceAccountName)
	}
	if got := deploy.Spec.Template.Annotations["openshift.io/required-scc"]; got != "noobaa-bg-executors" {
		t.Fatalf("required-scc = %q, want noobaa-bg-executors", got)
	}
	podSC := deploy.Spec.Template.Spec.SecurityContext
	if podSC == nil || podSC.RunAsUser == nil || *podSC.RunAsUser != 0 {
		t.Fatalf("pod securityContext.runAsUser = %v, want 0", podSC)
	}
	cSC := deploy.Spec.Template.Spec.Containers[0].SecurityContext
	if cSC == nil || cSC.Capabilities == nil {
		t.Fatalf("container securityContext capabilities missing: %+v", cSC)
	}
	hasSETUID, hasSETGID := false, false
	for _, cap := range cSC.Capabilities.Add {
		switch string(cap) {
		case "SETUID":
			hasSETUID = true
		case "SETGID":
			hasSETGID = true
		}
	}
	if !hasSETUID || !hasSETGID {
		t.Fatalf("capabilities.Add = %v, want SETUID and SETGID", cSC.Capabilities.Add)
	}
	probe := deploy.Spec.Template.Spec.Containers[0].ReadinessProbe
	if probe == nil || probe.Exec == nil || len(probe.Exec.Command) == 0 || probe.Exec.Command[0] != "true" {
		t.Fatalf("expected exec readinessProbe [true], got %+v", probe)
	}
}

func TestBgExecutorsAlreadyRunningOnActualImage(t *testing.T) {
	t.Parallel()

	const newImage = "noobaa/noobaa-core:new"
	const oldImage = "noobaa/noobaa-core:old"

	bgDeploy := func(replicas int32, image string) *appsv1.Deployment {
		return &appsv1.Deployment{
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
				Template: corev1.PodTemplateSpec{
					Spec: corev1.PodSpec{
						Containers: []corev1.Container{{
							Name:  "bg-executors",
							Image: image,
						}},
					},
				},
			},
		}
	}

	tests := []struct {
		name        string
		deploy      *appsv1.Deployment
		actualImage string
		want        bool
	}{
		{name: "nil deploy", deploy: nil, actualImage: newImage, want: false},
		{name: "empty actual image", deploy: bgDeploy(1, newImage), actualImage: "", want: false},
		{name: "zero replicas", deploy: bgDeploy(0, newImage), actualImage: newImage, want: false},
		{name: "old image still running", deploy: bgDeploy(1, oldImage), actualImage: newImage, want: false},
		{name: "already on actual image", deploy: bgDeploy(1, newImage), actualImage: newImage, want: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := bgExecutorsAlreadyRunningOnActualImage(tt.deploy, tt.actualImage)
			if got != tt.want {
				t.Fatalf("bgExecutorsAlreadyRunningOnActualImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDesiredBgExecutorsEnv(t *testing.T) {
	t.Parallel()

	r := &Reconciler{
		Request:          types.NamespacedName{Name: "noobaa", Namespace: "nbns"},
		ServiceMgmt:      util.KubeObject(bundle.File_deploy_internal_service_mgmt_yaml).(*corev1.Service),
		ServiceS3:        util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service),
		ServiceBgWorkers: util.KubeObject(bundle.File_deploy_internal_service_bg_workers_yaml).(*corev1.Service),
	}
	r.ServiceMgmt.Name = "noobaa-mgmt"
	r.ServiceS3.Name = "s3"
	r.ServiceBgWorkers.Name = "noobaa-bg-workers"

	c := &corev1.Container{
		Env: []corev1.EnvVar{
			{Name: "BG_WORKERS_MODE"},
			{Name: "MGMT_ADDR"},
			{Name: "HOSTED_AGENTS_ADDR"},
			{Name: "MD_ADDR"},
			{Name: "BG_ADDR"},
		},
	}
	r.setDesiredBgExecutorsEnv(c)

	got := map[string]string{}
	for _, e := range c.Env {
		got[e.Name] = e.Value
	}

	if got["BG_WORKERS_MODE"] != "executor" {
		t.Fatalf("BG_WORKERS_MODE = %q, want executor", got["BG_WORKERS_MODE"])
	}
	want := map[string]string{
		"MGMT_ADDR":          "wss://noobaa-mgmt.nbns.svc:443",
		"HOSTED_AGENTS_ADDR": "wss://noobaa-mgmt.nbns.svc:8446",
		"MD_ADDR":            "wss://s3.nbns.svc:8444",
		"BG_ADDR":            "wss://noobaa-bg-workers.nbns.svc:8445",
	}
	for name, wantAddr := range want {
		if got[name] != wantAddr {
			t.Fatalf("%s = %q, want %q", name, got[name], wantAddr)
		}
	}
	if strings.Contains(got["BG_ADDR"], "localhost") {
		t.Fatal("BG_ADDR must not be localhost for executors")
	}
}
