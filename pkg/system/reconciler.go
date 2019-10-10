package system

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
	"github.com/noobaa/noobaa-operator/v2/version"

	monitoringv1 "github.com/coreos/prometheus-operator/pkg/apis/monitoring/v1"
	semver "github.com/hashicorp/go-version"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
	// ContainerImageConstraint is the instantiated semver contraints used for image verification
	ContainerImageConstraint, _ = semver.NewConstraint(options.ContainerImageConstraintSemver)

	// ReadmeReady is a template for system.status.readme
	ReadmeReady = template.Must(template.New("system_status_readme_ready").
			Parse(bundle.File_deploy_internal_text_system_status_readme_ready_tmpl))

	// ReadmeProgress is a template for system.status.readme
	ReadmeProgress = template.Must(template.New("system_status_readme_progress").
			Parse(bundle.File_deploy_internal_text_system_status_readme_progress_tmpl))

	// ReadmeRejected is a template for system.status.readme
	ReadmeRejected = template.Must(template.New("system_status_readme_rejected").
			Parse(bundle.File_deploy_internal_text_system_status_readme_rejected_tmpl))
)

// Reconciler is the context for loading or reconciling a noobaa system
type Reconciler struct {
	Request         types.NamespacedName
	Client          client.Client
	Scheme          *runtime.Scheme
	Ctx             context.Context
	Logger          *logrus.Entry
	Recorder        record.EventRecorder
	NBClient        nb.Client
	CoreVersion     string
	OperatorVersion string

	NooBaa              *nbv1.NooBaa
	CoreApp             *appsv1.StatefulSet
	ServiceMgmt         *corev1.Service
	ServiceS3           *corev1.Service
	SecretServer        *corev1.Secret
	SecretOp            *corev1.Secret
	SecretAdmin         *corev1.Secret
	CloudCreds          *cloudcredsv1.CredentialsRequest
	DefaultBackingStore *nbv1.BackingStore
	DefaultBucketClass  *nbv1.BucketClass
	OBCStorageClass     *storagev1.StorageClass
	PrometheusRule      *monitoringv1.PrometheusRule
	ServiceMonitor      *monitoringv1.ServiceMonitor
	SystemInfo          *nb.SystemInfo
	CephObjectstoreUser *cephv1.CephObjectStoreUser
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a noobaa system
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:             req,
		Client:              client,
		Scheme:              scheme,
		Recorder:            recorder,
		OperatorVersion:     version.Version,
		CoreVersion:         options.ContainerImageTag,
		Ctx:                 context.TODO(),
		Logger:              logrus.WithField("sys", req.Namespace+"/"+req.Name),
		NooBaa:              util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		CoreApp:             util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet),
		ServiceMgmt:         util.KubeObject(bundle.File_deploy_internal_service_mgmt_yaml).(*corev1.Service),
		ServiceS3:           util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service),
		SecretServer:        util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretOp:            util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretAdmin:         util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		CloudCreds:          util.KubeObject(bundle.File_deploy_internal_cloud_creds_aws_cr_yaml).(*cloudcredsv1.CredentialsRequest),
		DefaultBackingStore: util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore),
		DefaultBucketClass:  util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_bucketclass_cr_yaml).(*nbv1.BucketClass),
		OBCStorageClass:     util.KubeObject(bundle.File_deploy_obc_storage_class_yaml).(*storagev1.StorageClass),
		PrometheusRule:      util.KubeObject(bundle.File_deploy_internal_prometheus_rules_yaml).(*monitoringv1.PrometheusRule),
		ServiceMonitor:      util.KubeObject(bundle.File_deploy_internal_service_monitor_yaml).(*monitoringv1.ServiceMonitor),
		CephObjectstoreUser: util.KubeObject(bundle.File_deploy_internal_ceph_objectstore_user_yaml).(*cephv1.CephObjectStoreUser),
	}

	// Set Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.CoreApp.Namespace = r.Request.Namespace
	r.ServiceMgmt.Namespace = r.Request.Namespace
	r.ServiceS3.Namespace = r.Request.Namespace
	r.SecretServer.Namespace = r.Request.Namespace
	r.SecretOp.Namespace = r.Request.Namespace
	r.SecretAdmin.Namespace = r.Request.Namespace
	r.CloudCreds.Namespace = r.Request.Namespace
	r.CloudCreds.Spec.SecretRef.Namespace = r.Request.Namespace
	r.DefaultBackingStore.Namespace = r.Request.Namespace
	r.DefaultBucketClass.Namespace = r.Request.Namespace
	r.PrometheusRule.Namespace = r.Request.Namespace
	r.ServiceMonitor.Namespace = r.Request.Namespace
	r.CephObjectstoreUser.Namespace = r.Request.Namespace

	// Set Names
	r.NooBaa.Name = r.Request.Name
	r.CoreApp.Name = r.Request.Name + "-core"
	r.ServiceMgmt.Name = r.Request.Name + "-mgmt"
	r.ServiceS3.Name = "s3"
	r.SecretServer.Name = r.Request.Name + "-server"
	r.SecretOp.Name = r.Request.Name + "-operator"
	r.SecretAdmin.Name = r.Request.Name + "-admin"
	r.CloudCreds.Name = r.Request.Name + "-cloud-creds"
	r.CloudCreds.Spec.SecretRef.Name = r.Request.Name + "-cloud-creds-secret"
	r.CephObjectstoreUser.Name = r.Request.Name + "-ceph-objectstore-user"
	r.DefaultBackingStore.Name = r.Request.Name + "-default-backing-store"
	r.DefaultBucketClass.Name = r.Request.Name + "-default-bucket-class"
	r.PrometheusRule.Name = r.Request.Name + "-prometheus-rules"
	r.ServiceMonitor.Name = r.Request.Name + "-service-monitor"

	// Since StorageClass is global we set the name and provisioner to have unique global name
	r.OBCStorageClass.Name = options.SubDomainNS()
	r.OBCStorageClass.Provisioner = options.ObjectBucketProvisionerName()

	return r
}

// CheckAll checks the state of all the objects controlled by the system
func (r *Reconciler) CheckAll() {
	util.KubeCheck(r.NooBaa)
	util.KubeCheck(r.CoreApp)
	util.KubeCheck(r.ServiceMgmt)
	util.KubeCheck(r.ServiceS3)
	util.KubeCheck(r.SecretServer)
	util.KubeCheck(r.SecretOp)
	util.KubeCheck(r.SecretAdmin)
	util.KubeCheck(r.OBCStorageClass)
	util.KubeCheck(r.DefaultBucketClass)
	util.KubeCheckOptional(r.DefaultBackingStore)
	util.KubeCheckOptional(r.CloudCreds)
	util.KubeCheckOptional(r.PrometheusRule)
	util.KubeCheckOptional(r.ServiceMonitor)
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {

	res := reconcile.Result{}
	log := r.Logger
	log.Infof("Start ...")

	util.KubeCheck(r.NooBaa)

	if r.NooBaa.UID == "" {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return res, nil
	}

	err := r.ReconcilePhases()

	if err != nil {
		if perr, isPERR := err.(*util.PersistentError); isPERR {
			r.SetPhase(nbv1.SystemPhaseRejected, perr.Reason, perr.Message)
			log.Errorf("❌ Persistent Error: %s", err)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.NooBaa, corev1.EventTypeWarning, perr.Reason, perr.Message)
			}
		} else {
			res.RequeueAfter = 3 * time.Second
			// leave current phase as is
			r.SetPhase("", "TemporaryError", err.Error())
			log.Warnf("⏳ Temporary Error: %s", err)
		}
	} else {
		r.SetPhase(
			nbv1.SystemPhaseReady,
			"SystemPhaseReady",
			"noobaa operator completed reconcile - system is ready",
		)
		log.Infof("✅ Done")
	}

	r.UpdateStatus()
	return res, nil
}

// ReconcilePhases runs the reconcile flow and populates System.Status.
func (r *Reconciler) ReconcilePhases() error {
	if err := r.ReconcilePhaseVerifying(); err != nil {
		return err
	}
	if err := r.ReconcilePhaseCreating(); err != nil {
		return err
	}
	if err := r.ReconcilePhaseConnecting(); err != nil {
		return err
	}
	if err := r.ReconcilePhaseConfiguring(); err != nil {
		return err
	}
	return nil
}

// SetPhase updates the status phase and conditions
func (r *Reconciler) SetPhase(phase nbv1.SystemPhase, reason string, message string) {

	c := &r.NooBaa.Status.Conditions

	if phase == "" {
		r.Logger.Infof("SetPhase: temporary error during phase %q", r.NooBaa.Status.Phase)
		r.SetReadme(ReadmeProgress)
		util.SetProgressingCondition(c, reason, message)
		return
	}

	r.Logger.Infof("SetPhase: %q", phase)
	switch phase {
	case nbv1.SystemPhaseReady:
		r.SetReadme(ReadmeReady)
		util.SetAvailableCondition(c, reason, message)
	case nbv1.SystemPhaseRejected:
		r.SetReadme(ReadmeRejected)
		util.SetErrorCondition(c, reason, message)
	default:
		r.SetReadme(ReadmeProgress)
		util.SetProgressingCondition(c, reason, message)
	}

	r.NooBaa.Status.Phase = phase
}

// SetReadme runs the template and sets the readme
func (r *Reconciler) SetReadme(t *template.Template) {
	var writer strings.Builder
	err := t.Execute(&writer, r)
	if err != nil {
		r.Logger.Errorf("SetReadme: Error in readme template %s: %v", t.Name(), err)
		readme := `
	
	ERROR: readme template %q failed: %v

	This is a problem in the noobaa operator code -
	Please report it to https://github.com/noobaa/noobaa-operator/issues
	
`
		r.NooBaa.Status.Readme = fmt.Sprintf(readme, t.Name(), err)
		return
	}
	r.NooBaa.Status.Readme = writer.String()
}

// UpdateStatus updates the system status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() {
	r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
	err := r.Client.Status().Update(r.Ctx, r.NooBaa)
	if err != nil {
		r.Logger.Errorf("UpdateStatus: %s", err)
	} else {
		r.Logger.Infof("UpdateStatus: Done generation %d", r.NooBaa.Generation)
	}
}

// ReconcileObject is a generic call to reconcile a kubernetes object
// desiredFunc can be passed to modify the object before create/update.
// Currently we ignore enforcing a desired state, but it might be needed on upgrades.
func (r *Reconciler) ReconcileObject(obj runtime.Object, desiredFunc func()) error {

	objMeta, _ := meta.Accessor(obj)
	r.Own(objMeta)

	op, err := controllerutil.CreateOrUpdate(
		r.Ctx, r.Client, obj.(runtime.Object),
		func(obj runtime.Object) error {
			if desiredFunc != nil {
				desiredFunc()
			}
			return nil
		},
	)
	if err != nil {
		r.Logger.Errorf("ReconcileObject: Error %v obj %s", err, objMeta.GetSelfLink())
		return err
	}

	r.Logger.Infof("ReconcileObject: Done - %s %s", op, objMeta.GetSelfLink())
	return nil
}

// Own sets the object owner references to the noobaa system
func (r *Reconciler) Own(obj metav1.Object) {
	util.Panic(controllerutil.SetControllerReference(r.NooBaa, obj, r.Scheme))
}
