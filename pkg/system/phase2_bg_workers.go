package system

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetDesiredServiceBgWorkers updates the BG workers Service selectors and labels.
func (r *Reconciler) SetDesiredServiceBgWorkers() error {
	r.ServiceBgWorkers.Spec.Selector["noobaa-bg-workers"] = r.Request.Name
	r.ServiceBgWorkers.Labels["noobaa-bg-workers-svc"] = "true"
	return nil
}

// SetDesiredNetworkPolicyBgWorkers sets the BG workers network policy spec from the
// bundle so bundle changes reach existing policies, and updates the pod selector
// to match the actual NooBaa system name.
//
// The cross-namespace rule allows TCP 7002 (metrics) from any namespace because
// Prometheus/monitoring scrapes this port and the scraper namespace varies by platform.
// RPC on 8445 stays same-namespace only.
func (r *Reconciler) SetDesiredNetworkPolicyBgWorkers() error {
	desired := util.KubeObject(bundle.File_deploy_internal_networkpolicy_bg_workers_yaml).(*networkingv1.NetworkPolicy)
	r.NetworkPolicyBgWorkers.Spec = desired.Spec
	r.NetworkPolicyBgWorkers.Spec.PodSelector.MatchLabels["noobaa-bg-workers"] = r.Request.Name
	return nil
}

// SetDesiredBgWorkersApp updates the BG workers Deployment as desired for reconciling.
// It copies CR scheduling, Postgres SSL mounts, and the noobaa-config hash so
// ConfigMap changes Recreate the BG workers.
func (r *Reconciler) SetDesiredBgWorkersApp() error {
	r.BgWorkersApp.Spec.Template.Labels["app"] = "noobaa"
	r.BgWorkersApp.Spec.Template.Labels["noobaa-bg-workers"] = r.Request.Name
	r.BgWorkersApp.Spec.Template.Labels["noobaa-component"] = "bg-workers"
	r.BgWorkersApp.Spec.Selector.MatchLabels["noobaa-bg-workers"] = r.Request.Name

	oneReplica := int32(1)
	r.BgWorkersApp.Spec.Replicas = &oneReplica

	podSpec := &r.BgWorkersApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-core"
	podSpec.Volumes = r.DefaultBgWorkersApp.Volumes

	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if i < len(r.DefaultBgWorkersApp.Containers) {
			c.VolumeMounts = r.DefaultBgWorkersApp.Containers[i].VolumeMounts
			util.MergeEnvArrays(&c.Env, &r.DefaultBgWorkersApp.Containers[i].Env)
			c.Command = append([]string(nil), r.DefaultBgWorkersApp.Containers[i].Command...)
			// Keep bundle defaults on existing Deployments (CreateOrUpdate starts from live object).
			c.Resources = r.DefaultBgWorkersApp.Containers[i].Resources
		}
		r.setDesiredCoreEnv(c)
		r.setDesiredBgWorkersEnv(c)

		if c.Name != "bg-workers" {
			continue
		}

		c.Image = r.NooBaa.Status.ActualImage
		util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
		util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
		util.ReflectEnvVariable(&c.Env, "NO_PROXY")

		r.setDesiredRootMasterKeyMounts(podSpec, c)

		if c.ReadinessProbe == nil {
			c.ReadinessProbe = &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					TCPSocket: &corev1.TCPSocketAction{
						Port: intstr.FromInt(8445),
					},
				},
				InitialDelaySeconds: 5,
				PeriodSeconds:       10,
				TimeoutSeconds:      2,
			}
		}
		// Upgrade: a startupProbe on :8445 would CrashLoop old images that
		// sleep because bg_init.js is missing. Readiness-only keeps 0/1 Running.
		c.StartupProbe = nil

		if r.shouldReconcileCNPGCluster() {
			dbSecretVolumeMounts := []corev1.VolumeMount{{
				Name:      r.CNPGCluster.Name,
				MountPath: postgresSecretMountPath,
				ReadOnly:  true,
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &dbSecretVolumeMounts)
		} else if r.NooBaa.Spec.ExternalPgSecret != nil {
			dbSecretVolumeMounts := []corev1.VolumeMount{{
				Name:      r.NooBaa.Spec.ExternalPgSecret.Name,
				MountPath: postgresSecretMountPath,
				ReadOnly:  true,
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &dbSecretVolumeMounts)
		}

		if util.KubeCheckQuiet(r.CaBundleConf) && len(r.CaBundleConf.Data) > 0 {
			configMapVolumeMounts := []corev1.VolumeMount{{
				Name:      r.CaBundleConf.Name,
				MountPath: "/etc/ocp-injected-ca-bundle",
				ReadOnly:  true,
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &configMapVolumeMounts)
		}
		if r.ExternalPgSSLSecret != nil && util.KubeCheckQuiet(r.ExternalPgSSLSecret) {
			secretVolumeMounts := []corev1.VolumeMount{{
				Name:      r.ExternalPgSSLSecret.Name,
				MountPath: "/etc/external-db-secret",
				ReadOnly:  true,
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &secretVolumeMounts)
		}
	}

	if r.shouldReconcileCNPGCluster() {
		dbSecretVolumes := []corev1.Volume{{
			Name: r.CNPGCluster.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.getClusterSecretName(),
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &dbSecretVolumes)
	} else if r.NooBaa.Spec.ExternalPgSecret != nil {
		dbSecretVolumes := []corev1.Volume{{
			Name: r.NooBaa.Spec.ExternalPgSecret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.NooBaa.Spec.ExternalPgSecret.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &dbSecretVolumes)
	}

	if util.KubeCheckQuiet(r.CaBundleConf) && len(r.CaBundleConf.Data) > 0 {
		configMapVolumes := []corev1.Volume{{
			Name: r.CaBundleConf.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.CaBundleConf.Name,
					},
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &configMapVolumes)
	}
	if r.ExternalPgSSLSecret != nil && util.KubeCheckQuiet(r.ExternalPgSSLSecret) {
		secretVolumes := []corev1.Volume{{
			Name: r.ExternalPgSSLSecret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.ExternalPgSSLSecret.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &secretVolumes)
	}

	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	podSpec.Tolerations = r.NooBaa.Spec.Tolerations
	podSpec.Affinity = r.GetAffinity()

	if r.BgWorkersApp.Spec.Template.Annotations == nil {
		r.BgWorkersApp.Spec.Template.Annotations = make(map[string]string)
	}
	r.BgWorkersApp.Spec.Template.Annotations[coreConfigMapHashAnnotation] = r.CoreAppConfig.Annotations[coreConfigMapHashAnnotation]
	r.BgWorkersApp.Spec.Template.Annotations[secv1.RequiredSCCAnnotation] = "noobaa-core"
	return nil
}

// setDesiredBgWorkersEnv fills RPC router and role env for the BG workers pod.
func (r *Reconciler) setDesiredBgWorkersEnv(c *corev1.Container) {
	mgmtBaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceMgmt.Name, r.Request.Namespace)
	s3BaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceS3.Name, r.Request.Namespace)

	for j := range c.Env {
		switch c.Env[j].Name {
		case "NOOBAA_BG_ROLE":
			c.Env[j].Value = "scanner"
		case "MGMT_ADDR":
			port := nb.FindPortByName(r.ServiceMgmt, "mgmt-https")
			c.Env[j].Value = fmt.Sprintf(`%s:%d`, mgmtBaseAddr, port.Port)
		case "HOSTED_AGENTS_ADDR":
			port := nb.FindPortByName(r.ServiceMgmt, "hosted-agents-https")
			c.Env[j].Value = fmt.Sprintf(`%s:%d`, mgmtBaseAddr, port.Port)
		case "MD_ADDR":
			port := nb.FindPortByName(r.ServiceS3, "md-https")
			c.Env[j].Value = fmt.Sprintf(`%s:%d`, s3BaseAddr, port.Port)
		case "BG_ADDR":
			// Self-RPC for archive/replication/scrubber within the BG workers process.
			c.Env[j].Value = "wss://localhost:8445"
		}
	}
}

// bgWorkersServiceAddr returns the cluster DNS WSS address for the BG workers RPC service.
func (r *Reconciler) bgWorkersServiceAddr() string {
	port := nb.FindPortByName(r.ServiceBgWorkers, "bg-https")
	return fmt.Sprintf(`wss://%s.%s.svc:%d`, r.ServiceBgWorkers.Name, r.Request.Namespace, port.Port)
}

// deleteLegacyBgScannerResources removes leftover *-bg-scanner objects after
// the rename to *-bg-workers.
func (r *Reconciler) deleteLegacyBgScannerResources() {
	legacyName := r.Request.Name + "-bg-scanner"
	ns := r.Request.Namespace
	objs := []client.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: ns}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: ns}},
		&networkingv1.NetworkPolicy{ObjectMeta: metav1.ObjectMeta{Name: legacyName, Namespace: ns}},
		&monitoringv1.ServiceMonitor{ObjectMeta: metav1.ObjectMeta{Name: legacyName + "-service-monitor", Namespace: ns}},
	}
	for _, obj := range objs {
		if err := r.Client.Delete(r.Ctx, obj); err != nil && !errors.IsNotFound(err) {
			r.Logger.Errorf("failed deleting leftover %T %s: %v", obj, obj.GetName(), err)
		}
	}
}
