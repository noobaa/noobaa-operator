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

func TestGetBgConsumersReplicas(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		spec *nbv1.BgConsumersSpec
		want int32
	}{
		{name: "nil defaults to 1", spec: nil, want: 1},
		{name: "empty defaults to 1", spec: &nbv1.BgConsumersSpec{}, want: 1},
		{name: "zero defaults to 1", spec: &nbv1.BgConsumersSpec{Replicas: 0}, want: 1},
		{name: "negative defaults to 1", spec: &nbv1.BgConsumersSpec{Replicas: -2}, want: 1},
		{name: "explicit replicas", spec: &nbv1.BgConsumersSpec{Replicas: 3}, want: 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nb := &nbv1.NooBaa{Spec: nbv1.NooBaaSpec{BgConsumers: tt.spec}}
			if got := getBgConsumersReplicas(nb); got != tt.want {
				t.Fatalf("getBgConsumersReplicas() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestBgConsumersCommandSleepsWhenBgInitMissing(t *testing.T) {
	t.Parallel()

	// Repo-layout parity: old images without bg_init.js must sleep instead of CrashLoop.
	deploy := util.KubeObject(bundle.File_deploy_internal_deployment_bg_consumers_yaml).(*appsv1.Deployment)
	joined := strings.Join(deploy.Spec.Template.Spec.Containers[0].Command, "\n")
	if !strings.Contains(joined, "bg_init.js") || !strings.Contains(joined, "sleep infinity") {
		t.Fatalf("bg-consumers command must sleep when bg_init.js is missing, got: %v",
			deploy.Spec.Template.Spec.Containers[0].Command)
	}
	if deploy.Spec.Template.Spec.Containers[0].Name != "bg-consumers" {
		t.Fatalf("expected container name bg-consumers, got %q", deploy.Spec.Template.Spec.Containers[0].Name)
	}
	role := ""
	for _, e := range deploy.Spec.Template.Spec.Containers[0].Env {
		if e.Name == "NOOBAA_BG_ROLE" {
			role = e.Value
			break
		}
	}
	if role != "consumer" {
		t.Fatalf("bundle NOOBAA_BG_ROLE = %q, want consumer", role)
	}
}

func TestBgConsumersAlreadyRunningOnActualImage(t *testing.T) {
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
							Name:  "bg-consumers",
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
			got := bgConsumersAlreadyRunningOnActualImage(tt.deploy, tt.actualImage)
			if got != tt.want {
				t.Fatalf("bgConsumersAlreadyRunningOnActualImage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetDesiredBgConsumersEnv(t *testing.T) {
	t.Parallel()

	r := &Reconciler{
		Request: types.NamespacedName{Name: "noobaa", Namespace: "nbns"},
		ServiceMgmt: util.KubeObject(bundle.File_deploy_internal_service_mgmt_yaml).(*corev1.Service),
		ServiceS3:   util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service),
		ServiceBgWorkers: util.KubeObject(bundle.File_deploy_internal_service_bg_workers_yaml).(*corev1.Service),
	}
	r.ServiceMgmt.Name = "noobaa-mgmt"
	r.ServiceS3.Name = "s3"
	r.ServiceBgWorkers.Name = "noobaa-bg-workers"

	c := &corev1.Container{
		Env: []corev1.EnvVar{
			{Name: "NOOBAA_BG_ROLE"},
			{Name: "MGMT_ADDR"},
			{Name: "HOSTED_AGENTS_ADDR"},
			{Name: "MD_ADDR"},
			{Name: "BG_ADDR"},
		},
	}
	r.setDesiredBgConsumersEnv(c)

	got := map[string]string{}
	for _, e := range c.Env {
		got[e.Name] = e.Value
	}

	if got["NOOBAA_BG_ROLE"] != "consumer" {
		t.Fatalf("NOOBAA_BG_ROLE = %q, want consumer", got["NOOBAA_BG_ROLE"])
	}
	if !strings.HasPrefix(got["MGMT_ADDR"], "wss://noobaa-mgmt.nbns.svc:") {
		t.Fatalf("MGMT_ADDR = %q, want wss://noobaa-mgmt.nbns.svc:…", got["MGMT_ADDR"])
	}
	if !strings.HasPrefix(got["HOSTED_AGENTS_ADDR"], "wss://noobaa-mgmt.nbns.svc:") {
		t.Fatalf("HOSTED_AGENTS_ADDR = %q, want wss://noobaa-mgmt.nbns.svc:…", got["HOSTED_AGENTS_ADDR"])
	}
	if !strings.HasPrefix(got["MD_ADDR"], "wss://s3.nbns.svc:") {
		t.Fatalf("MD_ADDR = %q, want wss://s3.nbns.svc:…", got["MD_ADDR"])
	}
	wantBG := "wss://noobaa-bg-workers.nbns.svc:8445"
	if got["BG_ADDR"] != wantBG {
		t.Fatalf("BG_ADDR = %q, want %q (scanner Service, not localhost)", got["BG_ADDR"], wantBG)
	}
	if strings.Contains(got["BG_ADDR"], "localhost") {
		t.Fatal("BG_ADDR must not be localhost for consumers")
	}
}
