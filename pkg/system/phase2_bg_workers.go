package system

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// SetDesiredServiceBgWorkers updates the BG workers Service selectors, labels, and ports.
// Ports come from the bundle so an existing Service drops the old HTTP metrics port.
func (r *Reconciler) SetDesiredServiceBgWorkers() error {
	r.ServiceBgWorkers.Spec.Selector["noobaa-bg-workers"] = r.Request.Name
	r.ServiceBgWorkers.Labels["noobaa-bg-workers-svc"] = "true"
	desired := util.KubeObject(bundle.File_deploy_internal_service_bg_workers_yaml).(*corev1.Service)
	r.ServiceBgWorkers.Spec.Ports = desiredServicePorts(r.ServiceBgWorkers.Spec.Ports, desired.Spec.Ports)
	return nil
}

// SetDesiredNetworkPolicyBgWorkers sets the BG workers network policy spec from the
// bundle so bundle changes reach existing policies, and updates the pod selector
// to match the actual NooBaa system name.
//
// The cross-namespace rule allows TCP 9445 (metrics HTTPS) from any namespace because
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
//
// Replicas stay at 0 until a Ready core pod runs ActualImage (and no Ready core
// still runs an older image), unless BG is already running on ActualImage. That
// avoids dual BG during upgrade without tearing BG down on every core Ready flap.
func (r *Reconciler) SetDesiredBgWorkersApp() error {
	r.BgWorkersApp.Spec.Template.Labels["app"] = "noobaa"
	r.BgWorkersApp.Spec.Template.Labels["noobaa-bg-workers"] = r.Request.Name
	r.BgWorkersApp.Spec.Template.Labels["noobaa-component"] = "bg-workers"
	r.BgWorkersApp.Spec.Selector.MatchLabels["noobaa-bg-workers"] = r.Request.Name

	coreReady, err := r.isCoreReadyOnActualImage()
	if err != nil {
		return err
	}
	// CreateOrUpdate passes the live object in; read image before we overwrite it.
	alreadyRunning := bgWorkersAlreadyRunningOnActualImage(r.BgWorkersApp, r.NooBaa.Status.ActualImage)
	replicas := int32(0)
	if coreReady || alreadyRunning {
		replicas = 1
	} else {
		r.Logger.Infof("holding bg-workers at 0 replicas until a Ready core pod runs %s",
			r.NooBaa.Status.ActualImage)
	}
	r.BgWorkersApp.Spec.Replicas = &replicas

	podSpec := &r.BgWorkersApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-core"
	podSpec.Volumes = r.DefaultBgWorkersApp.Volumes

	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if i < len(r.DefaultBgWorkersApp.Containers) {
			c.VolumeMounts = r.DefaultBgWorkersApp.Containers[i].VolumeMounts
			util.MergeEnvArrays(&c.Env, &r.DefaultBgWorkersApp.Containers[i].Env)
			c.Command = append([]string(nil), r.DefaultBgWorkersApp.Containers[i].Command...)
			c.Ports = append([]corev1.ContainerPort(nil), r.DefaultBgWorkersApp.Containers[i].Ports...)
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

// isCoreReadyOnActualImage reports whether core has cut over to ActualImage.
func (r *Reconciler) isCoreReadyOnActualImage() (bool, error) {
	podList := &corev1.PodList{}
	if !util.KubeList(podList, client.InNamespace(r.Request.Namespace), client.MatchingLabels{"noobaa-core": r.Request.Name}) {
		return false, fmt.Errorf("failed listing noobaa-core pods")
	}
	return coreReadyOnActualImage(r.NooBaa.Status.ActualImage, r.CoreApp, podList.Items), nil
}

// bgWorkersAlreadyRunningOnActualImage is true when the live BG Deployment already
// runs on actualImage. Used so a core Ready flap does not scale a healthy BG to 0.
func bgWorkersAlreadyRunningOnActualImage(dep *appsv1.Deployment, actualImage string) bool {
	if actualImage == "" || dep == nil || dep.Spec.Replicas == nil || *dep.Spec.Replicas == 0 {
		return false
	}
	return bgWorkersContainerImage(dep) == actualImage
}

func bgWorkersContainerImage(dep *appsv1.Deployment) string {
	if dep == nil {
		return ""
	}
	for i := range dep.Spec.Template.Spec.Containers {
		c := &dep.Spec.Template.Spec.Containers[i]
		if c.Name == "bg-workers" {
			return c.Image
		}
	}
	return ""
}

// coreReadyOnActualImage is true when the core STS template is on actualImage and
// at least one Ready, non-deleting core pod runs that image, with no Ready
// non-deleting core pod still on another image.
func coreReadyOnActualImage(actualImage string, sts *appsv1.StatefulSet, pods []corev1.Pod) bool {
	if actualImage == "" || sts == nil {
		return false
	}
	if len(sts.Spec.Template.Spec.Containers) == 0 ||
		sts.Spec.Template.Spec.Containers[0].Image != actualImage {
		return false
	}

	readyOnActual := false
	for i := range pods {
		pod := &pods[i]
		if pod.DeletionTimestamp != nil || !isPodReady(pod) {
			continue
		}
		img := ""
		if len(pod.Spec.Containers) > 0 {
			img = pod.Spec.Containers[0].Image
		}
		if img != actualImage {
			return false
		}
		readyOnActual = true
	}
	return readyOnActual
}
