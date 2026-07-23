package system

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const upgradeJobBackoffLimit = int32(4)

func (r *Reconciler) getDeployedCoreImage() string {
	for i := range r.CoreApp.Spec.Template.Spec.Containers {
		if r.CoreApp.Spec.Template.Spec.Containers[i].Name == "core" {
			return r.CoreApp.Spec.Template.Spec.Containers[i].Image
		}
	}
	return ""
}

func (r *Reconciler) isCoreImageUpgradeNeeded() bool {
	if !util.KubeCheckQuiet(r.CoreApp) {
		return false
	}
	return r.NooBaa.Status.ActualImage != r.getDeployedCoreImage()
}

func (r *Reconciler) isCoreUpgradeInProgress() bool {
	switch r.NooBaa.Status.UpgradePhase {
	case nbv1.UpgradePhaseUpgrade, nbv1.UpgradePhaseFailed:
		return true
	default:
		return false
	}
}

func (r *Reconciler) shouldBlockWorkloadRestore() bool {
	return r.isCoreUpgradeInProgress()
}

func (r *Reconciler) isCoreUpgradeFinishedForActualImage() bool {
	return r.NooBaa.Status.UpgradePhase == nbv1.UpgradePhaseFinished &&
		r.NooBaa.Status.LastSuccessfulUpgradeImage != "" &&
		r.NooBaa.Status.LastSuccessfulUpgradeImage == r.NooBaa.Status.ActualImage
}

func (r *Reconciler) reconcileCoreImageUpgrade() error {
	if r.NooBaa.Status.UpgradePhase == nbv1.UpgradePhaseFailed {
		if util.KubeCheckQuiet(r.UpgradeJob) {
			return fmt.Errorf("core upgrade job failed; delete job %q to retry", r.UpgradeJob.Name)
		}
		r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseUpgrade
		if err := r.UpdateStatus(); err != nil {
			return err
		}
	}

	if !r.isCoreImageUpgradeNeeded() {
		if r.NooBaa.Status.UpgradePhase != nbv1.UpgradePhaseNone && r.NooBaa.Status.UpgradePhase != "" {
			if r.getDeployedCoreImage() == r.NooBaa.Status.ActualImage {
				r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseNone
				if err := r.cleanupUpgradeJob(); err != nil {
					return err
				}
				return r.UpdateStatus()
			}
		}
		return nil
	}

	if r.isCoreUpgradeFinishedForActualImage() {
		return nil
	}

	numRunningPods, err := r.stopNoobaaPodsAndGetNumRunningPods()
	if err != nil {
		return fmt.Errorf("got error stopping noobaa-core and noobaa-endpoint pods. error: %v", err)
	}
	if numRunningPods != 0 {
		return fmt.Errorf("waiting for noobaa-core and noobaa-endpoint pods to be terminated before upgrade. %d pods are still running", numRunningPods)
	}

	if r.NooBaa.Status.UpgradePhase != nbv1.UpgradePhaseUpgrade {
		r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseUpgrade
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, nil, corev1.EventTypeNormal,
				"CoreUpgradeStarted", "CoreUpgradeStarted",
				"Starting core image upgrade to %q", r.NooBaa.Status.ActualImage)
		}
		if err := r.UpdateStatus(); err != nil {
			return err
		}
	}

	deleted, err := r.ensureUpgradeJobRecreate()
	if err != nil {
		return err
	}
	if deleted {
		return fmt.Errorf("waiting for old core upgrade job to be deleted before recreating it")
	}

	if err := r.ReconcileObject(r.UpgradeJob, r.setDesiredUpgradeJob); err != nil {
		return err
	}

	if err := r.Client.Get(r.Ctx, client.ObjectKeyFromObject(r.UpgradeJob), r.UpgradeJob); err != nil {
		return err
	}

	if r.UpgradeJob.Status.Succeeded > 0 {
		r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseFinished
		r.NooBaa.Status.LastSuccessfulUpgradeImage = r.NooBaa.Status.ActualImage
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, nil, corev1.EventTypeNormal,
				"CoreUpgradeSucceeded", "CoreUpgradeSucceeded",
				"Core image upgrade to %q completed successfully", r.NooBaa.Status.ActualImage)
		}
		if err := r.UpdateStatus(); err != nil {
			return err
		}
		return r.cleanupUpgradeJob()
	}

	if r.isUpgradeJobFailed() {
		r.NooBaa.Status.UpgradePhase = nbv1.UpgradePhaseFailed
		if r.Recorder != nil {
			r.Recorder.Eventf(r.NooBaa, nil, corev1.EventTypeWarning,
				"CoreUpgradeFailed", "CoreUpgradeFailed",
				"Core image upgrade to %q failed; delete job %q to retry",
				r.NooBaa.Status.ActualImage, r.UpgradeJob.Name)
		}
		if err := r.UpdateStatus(); err != nil {
			return err
		}
		return fmt.Errorf("core upgrade job failed; delete job %q to retry", r.UpgradeJob.Name)
	}

	return fmt.Errorf("waiting for core upgrade job to complete")
}

func (r *Reconciler) isUpgradeJobFailed() bool {
	for _, c := range r.UpgradeJob.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return r.UpgradeJob.Status.Failed > upgradeJobBackoffLimit
}

func (r *Reconciler) ensureUpgradeJobRecreate() (bool, error) {
	existing := &batchv1.Job{}
	key := client.ObjectKeyFromObject(r.UpgradeJob)
	err := r.Client.Get(r.Ctx, key, existing)
	if k8serrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if len(existing.Spec.Template.Spec.Containers) == 0 {
		return false, nil
	}
	if existing.Spec.Template.Spec.Containers[0].Image == r.NooBaa.Status.ActualImage {
		return false, nil
	}
	return true, r.deleteUpgradeJob(existing)
}

func (r *Reconciler) setDesiredUpgradeJob() error {
	podSpec := &r.UpgradeJob.Spec.Template.Spec
	defaultPodSpec := r.DefaultUpgradeJob
	defaultContainer := defaultPodSpec.Containers[0]

	podSpec.ServiceAccountName = "noobaa-core"
	podSpec.RestartPolicy = corev1.RestartPolicyNever
	podSpec.SecurityContext = defaultPodSpec.SecurityContext.DeepCopy()
	backoffLimit := upgradeJobBackoffLimit
	r.UpgradeJob.Spec.BackoffLimit = &backoffLimit

	if r.NooBaa.Spec.ImagePullSecret == nil {
		podSpec.ImagePullSecrets = []corev1.LocalObjectReference{}
	} else {
		podSpec.ImagePullSecrets = []corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}

	podSpec.Volumes = defaultPodSpec.Volumes

	c := &podSpec.Containers[0]
	c.Command = append([]string(nil), defaultContainer.Command...)
	c.EnvFrom = append([]corev1.EnvFromSource(nil), defaultContainer.EnvFrom...)
	c.VolumeMounts = defaultContainer.VolumeMounts
	c.SecurityContext = defaultContainer.SecurityContext.DeepCopy()
	util.MergeEnvArrays(&c.Env, &defaultContainer.Env)
	c.Image = r.NooBaa.Status.ActualImage
	r.setDesiredCoreEnv(c)
	r.setDesiredRootMasterKeyMounts(podSpec, c)

	if r.shouldReconcileCNPGCluster() {
		dbSecretVolumeMounts := []corev1.VolumeMount{{
			Name:      r.CNPGCluster.Name,
			MountPath: postgresSecretMountPath,
			ReadOnly:  true,
		}}
		util.MergeVolumeMountList(&c.VolumeMounts, &dbSecretVolumeMounts)
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
		dbSecretVolumeMounts := []corev1.VolumeMount{{
			Name:      r.NooBaa.Spec.ExternalPgSecret.Name,
			MountPath: postgresSecretMountPath,
			ReadOnly:  true,
		}}
		util.MergeVolumeMountList(&c.VolumeMounts, &dbSecretVolumeMounts)
		externalPgVolumes := []corev1.Volume{{
			Name: r.NooBaa.Spec.ExternalPgSecret.Name,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: r.NooBaa.Spec.ExternalPgSecret.Name,
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &externalPgVolumes)
	}

	if util.KubeCheckQuiet(r.CaBundleConf) && len(r.CaBundleConf.Data) > 0 {
		configMapVolumes := []corev1.Volume{{
			Name: r.CaBundleConf.Name,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.CaBundleConf.Name,
					},
					Items: []corev1.KeyToPath{{
						Key:  "ca-bundle.crt",
						Path: "ca-bundle.crt",
					}},
				},
			},
		}}
		util.MergeVolumeList(&podSpec.Volumes, &configMapVolumes)
		configMapVolumeMounts := []corev1.VolumeMount{{
			Name:      r.CaBundleConf.Name,
			MountPath: "/etc/ocp-injected-ca-bundle",
			ReadOnly:  true,
		}}
		util.MergeVolumeMountList(&c.VolumeMounts, &configMapVolumeMounts)
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
		secretVolumeMounts := []corev1.VolumeMount{{
			Name:      r.ExternalPgSSLSecret.Name,
			MountPath: "/etc/external-db-secret",
			ReadOnly:  true,
		}}
		util.MergeVolumeMountList(&c.VolumeMounts, &secretVolumeMounts)
	}

	util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
	util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
	util.ReflectEnvVariable(&c.Env, "NO_PROXY")

	return nil
}

func (r *Reconciler) cleanupUpgradeJob() error {
	if r.UpgradeJob == nil || r.UpgradeJob.Name == "" {
		return nil
	}
	if !util.KubeCheckQuiet(r.UpgradeJob) {
		return nil
	}
	return r.deleteUpgradeJob(r.UpgradeJob)
}

func (r *Reconciler) deleteUpgradeJob(job *batchv1.Job) error {
	propagationPolicy := metav1.DeletePropagationForeground
	deleteOpts := client.DeleteOptions{PropagationPolicy: &propagationPolicy}
	if err := r.Client.Delete(r.Ctx, job, &deleteOpts); err != nil && !k8serrors.IsNotFound(err) {
		return err
	}
	return nil
}
