package system

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// SetDesiredBgExecutorsApp updates the BG executors Deployment as desired for reconciling.
// It copies CR scheduling, Postgres SSL mounts, bucket-logging / notification
// volumes, and the noobaa-config hash so ConfigMap changes roll executor pods.
//
// Replicas stay at 0 until a Ready core pod runs ActualImage (and no Ready core
// still runs an older image), unless executors are already running on ActualImage.
// That avoids dual BG during upgrade without tearing executors down on every
// core Ready flap. Same gate as the scanner StatefulSet.
func (r *Reconciler) SetDesiredBgExecutorsApp() error {
	r.BgExecutorsApp.Spec.Template.Labels["app"] = "noobaa"
	r.BgExecutorsApp.Spec.Template.Labels["noobaa-bg-executors"] = r.Request.Name
	r.BgExecutorsApp.Spec.Template.Labels["noobaa-component"] = "bg-executors"
	r.BgExecutorsApp.Spec.Selector.MatchLabels["noobaa-bg-executors"] = r.Request.Name

	coreReady, err := r.isCoreReadyOnActualImage()
	if err != nil {
		return err
	}
	// CreateOrUpdate passes the live object in; read image before we overwrite it.
	alreadyRunning := bgExecutorsAlreadyRunningOnActualImage(r.BgExecutorsApp, r.NooBaa.Status.ActualImage)
	// Static scale from Spec.BgExecutors.Replicas (default 1). No HPA yet —
	// the operator owns Deployment.spec.replicas every reconcile.
	replicas := int32(0)
	if coreReady || alreadyRunning {
		replicas = r.getBgExecutorsReplicas()
	} else {
		r.Logger.Infof("holding bg-executors at 0 replicas until a Ready core pod runs %s",
			r.NooBaa.Status.ActualImage)
	}
	r.BgExecutorsApp.Spec.Replicas = &replicas

	podSpec := &r.BgExecutorsApp.Spec.Template.Spec
	podSpec.ServiceAccountName = "noobaa-bg-executors"
	podSpec.Volumes = r.DefaultBgExecutorsApp.Volumes
	// Keep pod securityContext from the bundle (runAsUser/Group 0) on existing Deployments.
	if r.DefaultBgExecutorsApp.SecurityContext != nil {
		podSpec.SecurityContext = r.DefaultBgExecutorsApp.SecurityContext.DeepCopy()
	}

	for i := range podSpec.Containers {
		c := &podSpec.Containers[i]
		for j := range r.DefaultBgExecutorsApp.Containers {
			def := &r.DefaultBgExecutorsApp.Containers[j]
			if def.Name != c.Name {
				continue
			}
			c.VolumeMounts = def.VolumeMounts
			util.MergeEnvArrays(&c.Env, &def.Env)
			c.Command = append([]string(nil), def.Command...)
			c.Ports = append([]corev1.ContainerPort(nil), def.Ports...)
			// Keep bundle defaults on existing Deployments (CreateOrUpdate starts from live object).
			c.Resources = def.Resources
			c.ReadinessProbe = def.ReadinessProbe.DeepCopy()
			if def.SecurityContext != nil {
				c.SecurityContext = def.SecurityContext.DeepCopy()
			}
			break
		}
		r.setDesiredCoreEnv(c)
		r.setDesiredBgExecutorsEnv(c)

		if c.Name != "bg-executors" {
			continue
		}

		c.Image = r.NooBaa.Status.ActualImage
		util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
		util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
		util.ReflectEnvVariable(&c.Env, "NO_PROXY")

		r.setDesiredRootMasterKeyMounts(podSpec, c)

		// No startup probe — temporary exec readinessProbe ("true") stands in
		// until MQ/health exists. Do not use a TCP probe on :8445 (scanner owns it).
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

	if r.BgExecutorsApp.Spec.Template.Annotations == nil {
		r.BgExecutorsApp.Spec.Template.Annotations = make(map[string]string)
	}
	r.BgExecutorsApp.Spec.Template.Annotations[coreConfigMapHashAnnotation] = r.CoreAppConfig.Annotations[coreConfigMapHashAnnotation]
	r.BgExecutorsApp.Spec.Template.Annotations[secv1.RequiredSCCAnnotation] = "noobaa-bg-executors"
	return nil
}

// setDesiredBgExecutorsEnv fills RPC router and mode env for the BG executors pod.
func (r *Reconciler) setDesiredBgExecutorsEnv(c *corev1.Container) {
	mgmtBaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceMgmt.Name, r.Request.Namespace)
	s3BaseAddr := fmt.Sprintf(`wss://%s.%s.svc`, r.ServiceS3.Name, r.Request.Namespace)

	for j := range c.Env {
		switch c.Env[j].Name {
		case "BG_WORKERS_MODE":
			// Executors only drain queues (no scanners / producers).
			c.Env[j].Value = "executor"
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
			// Executors talk to the scanner Service, not localhost (scanner owns RPC).
			c.Env[j].Value = r.bgWorkersServiceAddr()
		}
	}
}

// getBgExecutorsReplicas returns the static replica count from Spec.BgExecutors.
// Defaults to 1 when the field is unset or less than 1.
func getBgExecutorsReplicas(nb *nbv1.NooBaa) int32 {
	if nb != nil && nb.Spec.BgExecutors != nil && nb.Spec.BgExecutors.Replicas > 0 {
		return nb.Spec.BgExecutors.Replicas
	}
	return 1
}

func (r *Reconciler) getBgExecutorsReplicas() int32 {
	return getBgExecutorsReplicas(r.NooBaa)
}

// bgExecutorsAlreadyRunningOnActualImage is true when the live BG executors
// Deployment already runs on actualImage. Used so a core Ready flap does not
// scale a healthy executor Deployment to 0.
func bgExecutorsAlreadyRunningOnActualImage(deploy *appsv1.Deployment, actualImage string) bool {
	if actualImage == "" || deploy == nil || deploy.Spec.Replicas == nil || *deploy.Spec.Replicas == 0 {
		return false
	}
	return bgExecutorsContainerImage(deploy) == actualImage
}

func bgExecutorsContainerImage(deploy *appsv1.Deployment) string {
	if deploy == nil {
		return ""
	}
	for i := range deploy.Spec.Template.Spec.Containers {
		c := &deploy.Spec.Template.Spec.Containers[i]
		if c.Name == "bg-executors" {
			return c.Image
		}
	}
	return ""
}
