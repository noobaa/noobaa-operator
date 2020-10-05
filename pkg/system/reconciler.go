package system

import (
	"context"
	"fmt"
	goruntime "runtime"
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
	routev1 "github.com/openshift/api/route/v1"
	cloudcredsv1 "github.com/openshift/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var (
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
	OAuthEndpoints  *util.OAuth2Endpoints

	NooBaa              *nbv1.NooBaa
	ServiceAccount      *corev1.ServiceAccount
	CoreApp             *appsv1.StatefulSet
	NooBaaMongoDB       *appsv1.StatefulSet
	NooBaaPostgresDB    *appsv1.StatefulSet
	ServiceMgmt         *corev1.Service
	ServiceS3           *corev1.Service
	ServiceDb           *corev1.Service
	SecretServer        *corev1.Secret
	SecretOp            *corev1.Secret
	SecretAdmin         *corev1.Secret
	SecretEndpoints     *corev1.Secret
	AWSCloudCreds       *cloudcredsv1.CredentialsRequest
	AzureCloudCreds     *cloudcredsv1.CredentialsRequest
	AzureContainerCreds *corev1.Secret
	GCPBucketCreds      *corev1.Secret
	GCPCloudCreds       *cloudcredsv1.CredentialsRequest
	DefaultBackingStore *nbv1.BackingStore
	DefaultBucketClass  *nbv1.BucketClass
	OBCStorageClass     *storagev1.StorageClass
	PrometheusRule      *monitoringv1.PrometheusRule
	ServiceMonitorMgmt  *monitoringv1.ServiceMonitor
	ServiceMonitorS3    *monitoringv1.ServiceMonitor
	SystemInfo          *nb.SystemInfo
	CephObjectstoreUser *cephv1.CephObjectStoreUser
	RouteMgmt           *routev1.Route
	RouteS3             *routev1.Route
	DeploymentEndpoint  *appsv1.Deployment
	HPAEndpoint         *autoscalingv1.HorizontalPodAutoscaler
	JoinSecret          *corev1.Secret
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
		NooBaa:              util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		ServiceAccount:      util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount),
		CoreApp:             util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet),
		NooBaaMongoDB:       util.KubeObject(bundle.File_deploy_internal_statefulset_db_yaml).(*appsv1.StatefulSet),
		NooBaaPostgresDB:    util.KubeObject(bundle.File_deploy_internal_statefulset_postgres_db_yaml).(*appsv1.StatefulSet),
		ServiceDb:           util.KubeObject(bundle.File_deploy_internal_service_db_yaml).(*corev1.Service),
		ServiceMgmt:         util.KubeObject(bundle.File_deploy_internal_service_mgmt_yaml).(*corev1.Service),
		ServiceS3:           util.KubeObject(bundle.File_deploy_internal_service_s3_yaml).(*corev1.Service),
		SecretServer:        util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretOp:            util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretAdmin:         util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		SecretEndpoints:     util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		AzureContainerCreds: util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		GCPBucketCreds:      util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		AWSCloudCreds:       util.KubeObject(bundle.File_deploy_internal_cloud_creds_aws_cr_yaml).(*cloudcredsv1.CredentialsRequest),
		AzureCloudCreds:     util.KubeObject(bundle.File_deploy_internal_cloud_creds_azure_cr_yaml).(*cloudcredsv1.CredentialsRequest),
		GCPCloudCreds:       util.KubeObject(bundle.File_deploy_internal_cloud_creds_gcp_cr_yaml).(*cloudcredsv1.CredentialsRequest),
		DefaultBackingStore: util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore),
		DefaultBucketClass:  util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml).(*nbv1.BucketClass),
		OBCStorageClass:     util.KubeObject(bundle.File_deploy_obc_storage_class_yaml).(*storagev1.StorageClass),
		PrometheusRule:      util.KubeObject(bundle.File_deploy_internal_prometheus_rules_yaml).(*monitoringv1.PrometheusRule),
		ServiceMonitorMgmt:  util.KubeObject(bundle.File_deploy_internal_servicemonitor_mgmt_yaml).(*monitoringv1.ServiceMonitor),
		ServiceMonitorS3:    util.KubeObject(bundle.File_deploy_internal_servicemonitor_s3_yaml).(*monitoringv1.ServiceMonitor),
		CephObjectstoreUser: util.KubeObject(bundle.File_deploy_internal_ceph_objectstore_user_yaml).(*cephv1.CephObjectStoreUser),
		RouteMgmt:           util.KubeObject(bundle.File_deploy_internal_route_mgmt_yaml).(*routev1.Route),
		RouteS3:             util.KubeObject(bundle.File_deploy_internal_route_s3_yaml).(*routev1.Route),
		DeploymentEndpoint:  util.KubeObject(bundle.File_deploy_internal_deployment_endpoint_yaml).(*appsv1.Deployment),
		HPAEndpoint:         util.KubeObject(bundle.File_deploy_internal_hpa_endpoint_yaml).(*autoscalingv1.HorizontalPodAutoscaler),
	}

	// Set Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.ServiceAccount.Namespace = r.Request.Namespace
	r.CoreApp.Namespace = r.Request.Namespace
	r.NooBaaMongoDB.Namespace = r.Request.Namespace
	r.NooBaaPostgresDB.Namespace = r.Request.Namespace
	r.ServiceMgmt.Namespace = r.Request.Namespace
	r.ServiceS3.Namespace = r.Request.Namespace
	r.ServiceDb.Namespace = r.Request.Namespace
	r.SecretServer.Namespace = r.Request.Namespace
	r.SecretOp.Namespace = r.Request.Namespace
	r.SecretAdmin.Namespace = r.Request.Namespace
	r.SecretEndpoints.Namespace = r.Request.Namespace
	r.AzureContainerCreds.Namespace = r.Request.Namespace
	r.GCPBucketCreds.Namespace = r.Request.Namespace
	r.AWSCloudCreds.Namespace = r.Request.Namespace
	r.AWSCloudCreds.Spec.SecretRef.Namespace = r.Request.Namespace
	r.AzureCloudCreds.Namespace = r.Request.Namespace
	r.AzureCloudCreds.Spec.SecretRef.Namespace = r.Request.Namespace
	r.GCPCloudCreds.Namespace = r.Request.Namespace
	r.GCPCloudCreds.Spec.SecretRef.Namespace = r.Request.Namespace
	r.DefaultBackingStore.Namespace = r.Request.Namespace
	r.DefaultBucketClass.Namespace = r.Request.Namespace
	r.PrometheusRule.Namespace = r.Request.Namespace
	r.ServiceMonitorMgmt.Namespace = r.Request.Namespace
	r.ServiceMonitorS3.Namespace = r.Request.Namespace
	r.CephObjectstoreUser.Namespace = r.Request.Namespace
	r.RouteMgmt.Namespace = r.Request.Namespace
	r.RouteS3.Namespace = r.Request.Namespace
	r.DeploymentEndpoint.Namespace = r.Request.Namespace
	r.HPAEndpoint.Namespace = r.Request.Namespace

	// Set Names
	r.NooBaa.Name = r.Request.Name
	r.ServiceAccount.Name = r.Request.Name
	r.CoreApp.Name = r.Request.Name + "-core"
	r.NooBaaMongoDB.Name = r.Request.Name + "-db"
	r.NooBaaPostgresDB.Name = r.Request.Name + "-db"
	r.ServiceMgmt.Name = r.Request.Name + "-mgmt"
	r.ServiceS3.Name = "s3"
	r.ServiceDb.Name = r.Request.Name + "-db"
	r.SecretServer.Name = r.Request.Name + "-server"
	r.SecretOp.Name = r.Request.Name + "-operator"
	r.SecretAdmin.Name = r.Request.Name + "-admin"
	r.SecretEndpoints.Name = r.Request.Name + "-endpoints"
	r.AWSCloudCreds.Name = r.Request.Name + "-aws-cloud-creds"
	r.AWSCloudCreds.Spec.SecretRef.Name = r.Request.Name + "-aws-cloud-creds-secret"
	r.AzureContainerCreds.Name = r.Request.Name + "-azure-container-creds"
	r.AzureCloudCreds.Name = r.Request.Name + "-azure-cloud-creds"
	r.AzureCloudCreds.Spec.SecretRef.Name = r.Request.Name + "-azure-cloud-creds-secret"
	r.GCPBucketCreds.Name = r.Request.Name + "-gcp-bucket-creds"
	r.GCPCloudCreds.Name = r.Request.Name + "-gcp-cloud-creds"
	r.GCPCloudCreds.Spec.SecretRef.Name = r.Request.Name + "-gcp-cloud-creds-secret"
	r.CephObjectstoreUser.Name = r.Request.Name + "-ceph-objectstore-user"
	r.DefaultBackingStore.Name = r.Request.Name + "-default-backing-store"
	r.DefaultBucketClass.Name = r.Request.Name + "-default-bucket-class"
	r.PrometheusRule.Name = r.Request.Name + "-prometheus-rules"
	r.ServiceMonitorMgmt.Name = r.ServiceMgmt.Name + "-service-monitor"
	r.ServiceMonitorS3.Name = r.ServiceS3.Name + "-service-monitor"
	r.RouteMgmt.Name = r.ServiceMgmt.Name
	r.RouteS3.Name = r.ServiceS3.Name
	r.DeploymentEndpoint.Name = r.Request.Name + "-endpoint"
	r.HPAEndpoint.Name = r.Request.Name + "-endpoint"

	// Set the target service for routes.
	r.RouteMgmt.Spec.To.Name = r.ServiceMgmt.Name
	r.RouteS3.Spec.To.Name = r.ServiceS3.Name

	// Set the target deployment for the horizontal auto scaler
	r.HPAEndpoint.Spec.ScaleTargetRef.Name = r.DeploymentEndpoint.Name

	// Since StorageClass is global we set the name and provisioner to have unique global name
	r.OBCStorageClass.Name = options.SubDomainNS()
	r.OBCStorageClass.Provisioner = options.ObjectBucketProvisionerName()

	r.SecretServer.StringData["jwt"] = util.RandomBase64(16)
	r.SecretServer.StringData["server_secret"] = util.RandomHex(4)

	return r
}

// CheckAll checks the state of all the objects controlled by the system
func (r *Reconciler) CheckAll() {

	CheckSystem(r.NooBaa)
	util.KubeCheck(r.CoreApp)
	util.KubeCheck(r.ServiceMgmt)
	util.KubeCheck(r.ServiceS3)
	if r.NooBaa.Spec.DBType == "postgres" {
		util.KubeCheck(r.NooBaaPostgresDB)
	} else {
		util.KubeCheck(r.NooBaaMongoDB)
	}
	util.KubeCheck(r.ServiceDb)
	util.KubeCheck(r.SecretServer)
	util.KubeCheck(r.SecretOp)
	util.KubeCheck(r.SecretEndpoints)
	util.KubeCheck(r.SecretAdmin)
	util.KubeCheck(r.OBCStorageClass)
	util.KubeCheck(r.DefaultBucketClass)
	util.KubeCheck(r.DeploymentEndpoint)
	util.KubeCheck(r.HPAEndpoint)
	util.KubeCheckOptional(r.DefaultBackingStore)
	util.KubeCheckOptional(r.AWSCloudCreds)
	util.KubeCheckOptional(r.AzureCloudCreds)
	util.KubeCheckOptional(r.AzureContainerCreds)
	util.KubeCheckOptional(r.GCPBucketCreds)
	util.KubeCheckOptional(r.GCPCloudCreds)
	util.KubeCheckOptional(r.PrometheusRule)
	util.KubeCheckOptional(r.ServiceMonitorMgmt)
	util.KubeCheckOptional(r.ServiceMonitorS3)
	util.KubeCheckOptional(r.RouteMgmt)
	util.KubeCheckOptional(r.RouteS3)
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {

	res := reconcile.Result{}
	log := r.Logger
	log.Infof("Start ...")

	if !CheckSystem(r.NooBaa) {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return res, nil
	}

	if util.EnsureCommonMetaFields(r.NooBaa, nbv1.GracefulFinalizer) {
		if !util.KubeUpdate(r.NooBaa) {
			log.Errorf("❌ NooBaa %q failed to add mandatory meta fields", r.NooBaa.Name)

			res.RequeueAfter = 3 * time.Second
			return res, nil
		}
	}

	if err := r.VerifyObjectBucketCleanup(); err != nil {
		r.SetPhase("", "TemporaryError", err.Error())
		log.Warnf("⏳ Temporary Error: %s", err)
	}

	if r.NooBaa.Spec.JoinSecret != nil {
		r.JoinSecret = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: r.NooBaa.Spec.JoinSecret.Namespace,
				Name:      r.NooBaa.Spec.JoinSecret.Name,
			},
		}

		if !util.KubeCheck(r.JoinSecret) {
			log.Infof("Join secret not found or already deleted. Skip reconcile.")
			return res, nil
		}
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

	err = r.UpdateStatus()
	// if updateStatus will fail to update the CR for any reason we will continue to requeue the reconcile
	// until the spec status will reflect the actual status of the bucketclass
	if err != nil {
		res.RequeueAfter = 3 * time.Second
		log.Warnf("⏳ Temporary Error: %s", err)
	}
	return res, nil
}

// VerifyObjectBucketCleanup checks if the un-installation is in mode graceful and
// if OBs still exist in the system the operator will wait
// and the finalizer on noobaa CR won't be removed
func (r *Reconciler) VerifyObjectBucketCleanup() error {
	log := r.Logger

	if r.NooBaa.DeletionTimestamp == nil {
		return nil
	}

	if r.NooBaa.Spec.CleanupPolicy.Confirmation == nbv1.DeleteOBCConfirmation {
		util.RemoveFinalizer(r.NooBaa, nbv1.GracefulFinalizer)
		if !util.KubeUpdate(r.NooBaa) {
			log.Errorf("NooBaa %q failed to remove finalizer %q", r.NooBaa.Name, nbv1.GracefulFinalizer)
		}
		return nil
	}

	obcSelector, _ := labels.Parse("noobaa-domain=" + options.SubDomainNS())
	objectBuckets := &nbv1.ObjectBucketList{}
	util.KubeList(objectBuckets, &client.ListOptions{LabelSelector: obcSelector})

	if len(objectBuckets.Items) != 0 {
		var bucketNames []string
		for i := range objectBuckets.Items {
			ob := &objectBuckets.Items[i]
			bucketNames = append(bucketNames, ob.Name)
		}
		msg := fmt.Sprintf("Failed to delete NooBaa. object buckets in namespace %q are not cleaned up. remaining buckets: %+v",
			r.NooBaa.Namespace, bucketNames)
		log.Errorf(msg)
		return fmt.Errorf(msg)
	}

	log.Infof("All object buckets deleted in namespace %q", r.NooBaa.Namespace)
	util.RemoveFinalizer(r.NooBaa, nbv1.GracefulFinalizer)
	if !util.KubeUpdate(r.NooBaa) {
		log.Errorf("NooBaa %q failed to remove finalizer %q", r.NooBaa.Name, nbv1.GracefulFinalizer)
	}

	return nil
}

// bToMb convert bytes to megabytes
func (r *Reconciler) bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// PrintMemUsage prints memory usage message.
func (r *Reconciler) PrintMemUsage(phase string) {
	var m goruntime.MemStats
	goruntime.ReadMemStats(&m)
	r.Logger.Infof("Memory Usage: Phase %q - Alloc = %v MiB  Sys = %v MiB  NumGC = %v", phase, r.bToMb(m.Alloc), r.bToMb(m.Sys), m.NumGC)
}

// ReconcilePhases runs the reconcile flow and populates System.Status.
func (r *Reconciler) ReconcilePhases() error {
	r.PrintMemUsage("Starting")
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
	r.PrintMemUsage("Finishing")
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
func (r *Reconciler) UpdateStatus() error {
	r.NooBaa.Status.ObservedGeneration = r.NooBaa.Generation
	err := r.Client.Status().Update(r.Ctx, r.NooBaa)
	if err != nil {
		r.Logger.Errorf("UpdateStatus: %s", err)
		return err
	}
	r.Logger.Infof("UpdateStatus: Done generation %d", r.NooBaa.Generation)
	return nil
}

// ReconcileObject is a generic call to reconcile a kubernetes object
// desiredFunc can be passed to modify the object before create/update.
// Currently we ignore enforcing a desired state, but it might be needed on upgrades.
func (r *Reconciler) ReconcileObject(obj runtime.Object, desiredFunc func() error) error {
	return r.reconcileObject(obj, desiredFunc, false)
}

// ReconcileObjectOptional is like ReconcileObject but also ignores if the CRD is missing
func (r *Reconciler) ReconcileObjectOptional(obj runtime.Object, desiredFunc func() error) error {
	return r.reconcileObject(obj, desiredFunc, true)
}

func (r *Reconciler) reconcileObject(obj runtime.Object, desiredFunc func() error, optionalCRD bool) error {

	gvk := obj.GetObjectKind().GroupVersionKind()
	objMeta, _ := meta.Accessor(obj)
	r.Own(objMeta)

	op, err := controllerutil.CreateOrUpdate(
		r.Ctx, r.Client, obj.(runtime.Object), func() error {
			if desiredFunc != nil {
				if err := desiredFunc(); err != nil {
					return err
				}
			}
			return nil
		},
	)
	if err != nil {
		if optionalCRD && (meta.IsNoMatchError(err) || runtime.IsNotRegisteredError(err)) {
			r.Logger.Printf("ReconcileObject: (Optional) CRD Unavailable: %s %s\n", gvk.Kind, objMeta.GetSelfLink())
			return nil
		}
		r.Logger.Errorf("ReconcileObject: Error %s %s %v", gvk.Kind, objMeta.GetSelfLink(), err)
		return err
	}

	r.Logger.Infof("ReconcileObject: Done - %s %s %s", op, gvk.Kind, objMeta.GetSelfLink())
	return nil
}

// Own sets the object owner references to the noobaa system
func (r *Reconciler) Own(obj metav1.Object) {
	util.Panic(controllerutil.SetControllerReference(r.NooBaa, obj, r.Scheme))
}
