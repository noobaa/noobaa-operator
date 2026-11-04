package system

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
)

// SetDesiredNetworkPolicyBgConsumers sets the BG consumers network policy spec from
// the bundle so bundle changes reach existing policies, and updates the pod
// selector to match the actual NooBaa system name.
//
// Consumers are outbound-only for now (no Service/metrics), so the policy denies
// all ingress. A metrics allow rule can be added when a ServiceMonitor lands.
func (r *Reconciler) SetDesiredNetworkPolicyBgConsumers() error {
	desired := util.KubeObject(bundle.File_deploy_internal_networkpolicy_bg_consumers_yaml).(*networkingv1.NetworkPolicy)
	r.NetworkPolicyBgConsumers.Spec = desired.Spec
	r.NetworkPolicyBgConsumers.Spec.PodSelector.MatchLabels["noobaa-bg-consumers"] = r.Request.Name
	return nil
}

// SetDesiredBgConsumersApp updates the BG consumers Deployment as desired for reconciling.
// It copies CR scheduling, Postgres SSL mounts, bucket-logging / notification
// volumes, and the noobaa-config hash so ConfigMap changes roll consumer pods.
//
// Replicas stay at 0 until a Ready core pod runs ActualImage (and no Ready core
// still runs an older image), unless consumers are already running on ActualImage.
// That avoids dual BG during upgrade without tearing consumers down on every
// core Ready flap. Same gate as the scanner StatefulSet.
func (r *Reconciler) SetDesiredBgConsumersApp() error {
	r.BgConsumersApp.Spec.Template.Labels["app"] = "noobaa"
	r.BgConsumersApp.Spec.Template.Labels["noobaa-bg-consumers"] = r.Request.Name
	r.BgConsumersApp.Spec.Template.Labels["noobaa-component"] = "bg-consumers"
	r.BgConsumersApp.Spec.Selector.MatchLabels["noobaa-bg-consumers"] = r.Request.Name

	coreReady, err := r.isCoreReadyOnActualImage()
	if err != nil {
		return err
	}
	// CreateOrUpdate passes the live object in; read image before we overwrite it.
	alreadyRunning := bgConsumersAlreadyRunningOnActualImage(r.BgConsumersApp, r.NooBaa.Status.ActualImage)
	// Static scale from Spec.BgConsumers.Replicas (default 1). No HPA yet —
	// the operator owns Deployment.spec.replicas every reconcile.
	replicas := int32(0)
	if coreReady || alreadyRunning {
		replicas = r.getBgConsumersReplicas()
	} else {
		r.Logger.Infof("holding bg-consumers at 0 replicas until a Ready core pod runs %s",
			r.NooBaa.Status.ActualImage)
	}
	r.BgConsumersApp.Spec.Replicas = &replicas

	podSpec := &r.BgConsumersApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-core"
	podSpec.Volumes = r.DefaultBgConsumersApp.Volumes

	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		if i < len(r.DefaultBgConsumersApp.Containers) {
			c.VolumeMounts = r.DefaultBgConsumersApp.Containers[i].VolumeMounts
			util.MergeEnvArrays(&c.Env, &r.DefaultBgConsumersApp.Containers[i].Env)
			c.Command = append([]string(nil), r.DefaultBgConsumersApp.Containers[i].Command...)
			c.Ports = append([]corev1.ContainerPort(nil), r.DefaultBgConsumersApp.Containers[i].Ports...)
			// Keep bundle defaults on existing Deployments (CreateOrUpdate starts from live object).
			c.Resources = r.DefaultBgConsumersApp.Containers[i].Resources
		}
		r.setDesiredCoreEnv(c)
		r.setDesiredBgConsumersEnv(c)

		if c.Name != "bg-consumers" {
			continue
		}

		c.Image = r.NooBaa.Status.ActualImage
		util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
		util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
		util.ReflectEnvVariable(&c.Env, "NO_PROXY")

		r.setDesiredRootMasterKeyMounts(podSpec, c)

		// No RPC readiness on :8445 — scanner owns that port. Consumers may sleep
		// in core until the consumer/MQ path exists; a TCP probe would never Ready.
		c.ReadinessProbe = nil
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
		if r.NooBaa.Spec.BucketLogging.LoggingType == nbv1.BucketLoggingTypeGuaranteed {
			bucketLogVolumeMounts := []corev1.VolumeMount{{
				Name:      r.BucketLoggingVolume,
				MountPath: r.BucketLoggingVolumeMount,
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &bucketLogVolumeMounts)
		}
		if r.NooBaa.Spec.BucketNotifications.Enabled {
			notificationVolumeMounts := []corev1.VolumeMount{{
				Name:      notificationsVolume,
				MountPath: "/var/logs/notifications",
			}}
			util.MergeVolumeMountList(&c.VolumeMounts, &notificationVolumeMounts)

			for _, notifSecret := range r.NooBaa.Spec.BucketNotifications.Connections {
				secretVolumeMounts := []corev1.VolumeMount{{
					Name:      notifSecret.Name,
					MountPath: "/etc/notif_connect/" + notifSecret.Name,
					ReadOnly:  true,
				}}
				util.MergeVolumeMountList(&c.VolumeMounts, &secretVolumeMounts)
			}
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
	if r.NooBaa.Spec.BucketLogging.LoggingType == nbv1.BucketLoggingTypeGuaranteed {
		bucketLogVolumes := []corev1.Volume{{
			Name: r.BucketLoggingVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: r.BucketLoggingPVC.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &bucketLogVolumes)
	}
	if r.NooBaa.Spec.BucketNotifications.Enabled {
		notificationVolumes := []corev1.Volume{{
			Name: notificationsVolume,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: r.BucketNotificationsPVC.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &notificationVolumes)

		for _, notifSecret := range r.NooBaa.Spec.BucketNotifications.Connections {
			secretVolumes := []corev1.Volume{{
				Name: notifSecret.Name,
				VolumeSource: corev1.VolumeSource{
					Secret: &corev1.SecretVolumeSource{
						SecretName: notifSecret.Name,
					},
				},
			}}
			util.MergeVolumeList(&podSpec.Volumes, &secretVolumes)
		}
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

	if r.BgConsumersApp.Spec.Template.Annotations == nil {
		r.BgConsumersApp.Spec.Template.Annotations = make(map[string]string)
	}
	r.BgConsumersApp.Spec.Template.Annotations[coreConfigMapHashAnnotation] = r.CoreAppConfig.Annotations[coreConfigMapHashAnnotation]
	r.BgConsumersApp.Spec.Template.Annotations[secv1.RequiredSCCAnnotation] = "noobaa-core"
	return nil
}

// setDesiredBgConsumersEnv fills RPC router and role env for the BG consumers pod.
func (r *Reconciler) setDesiredBgConsumersEnv(c *corev1.Container) {
	mgmtBaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceMgmt.Name, r.Request.Namespace)
	s3BaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceS3.Name, r.Request.Namespace)

	for j := range c.Env {
		switch c.Env[j].Name {
		case "NOOBAA_BG_ROLE":
			c.Env[j].Value = "consumer"
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
			// Consumers talk to the scanner Service, not localhost (scanner owns RPC).
			c.Env[j].Value = r.bgWorkersServiceAddr()
		}
	}
}

// getBgConsumersReplicas returns the static replica count from Spec.BgConsumers.
// Defaults to 1 when the field is unset or less than 1.
func getBgConsumersReplicas(nb *nbv1.NooBaa) int32 {
	if nb != nil && nb.Spec.BgConsumers != nil && nb.Spec.BgConsumers.Replicas > 0 {
		return nb.Spec.BgConsumers.Replicas
	}
	return 1
}

func (r *Reconciler) getBgConsumersReplicas() int32 {
	return getBgConsumersReplicas(r.NooBaa)
}

// bgConsumersAlreadyRunningOnActualImage is true when the live BG consumers
// Deployment already runs on actualImage. Used so a core Ready flap does not
// scale a healthy consumer Deployment to 0.
func bgConsumersAlreadyRunningOnActualImage(deploy *appsv1.Deployment, actualImage string) bool {
	if actualImage == "" || deploy == nil || deploy.Spec.Replicas == nil || *deploy.Spec.Replicas == 0 {
		return false
	}
	return bgConsumersContainerImage(deploy) == actualImage
}

func bgConsumersContainerImage(deploy *appsv1.Deployment) string {
	if deploy == nil {
		return ""
	}
	for i := range deploy.Spec.Template.Spec.Containers {
		c := &deploy.Spec.Template.Spec.Containers[i]
		if c.Name == "bg-consumers" {
			return c.Image
		}
	}
	return ""
}
