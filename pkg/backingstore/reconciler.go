package backingstore

import (
	"context"
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
	"github.com/noobaa/noobaa-operator/pkg/options"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// Reconciler is the context for loading or reconciling a noobaa system
type Reconciler struct {
	Request  types.NamespacedName
	Client   client.Client
	Scheme   *runtime.Scheme
	Ctx      context.Context
	Logger   *logrus.Entry
	Recorder record.EventRecorder
	NBClient nb.Client

	BackingStore *nbv1.BackingStore
	NooBaa       *nbv1.NooBaa
	Secret       *corev1.Secret

	SystemInfo             *nb.SystemInfo
	ExternalConnectionInfo *nb.ExternalConnectionInfo
	PoolInfo               *nb.PoolInfo

	AddExternalConnectionParams *nb.AddExternalConnectionParams
	CreateCloudPoolParams       *nb.CreateCloudPoolParams
	CreateHostsPoolParams       *nb.CreateHostsPoolParams
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a backing store
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:      req,
		Client:       client,
		Scheme:       scheme,
		Recorder:     recorder,
		Ctx:          context.TODO(),
		Logger:       logrus.WithField("backingstore", req.Namespace+"/"+req.Name),
		BackingStore: util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore),
		NooBaa:       util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		Secret:       util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
	}

	// Set Namespace
	r.BackingStore.Namespace = r.Request.Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.Secret.Namespace = r.Request.Namespace

	// Set Names
	r.BackingStore.Name = r.Request.Name
	r.NooBaa.Name = options.SystemName
	r.Secret.Name = "backing-store-secret-" + r.Request.Name

	return r
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {

	res := reconcile.Result{}
	log := r.Logger
	log.Infof("Start ...")

	util.KubeCheck(r.BackingStore)

	if r.BackingStore.UID == "" {
		log.Infof("BackingStore %q not found or deleted. Skip reconcile.", r.BackingStore.Name)
		return reconcile.Result{}, nil
	}

	r.Secret.Name = r.BackingStore.Spec.Secret.Name
	r.Secret.Namespace = r.BackingStore.Spec.Secret.Namespace

	util.KubeCheck(r.NooBaa)
	util.KubeCheck(r.Secret)

	var err error
	if r.BackingStore.DeletionTimestamp != nil {
		err = r.ReconcileDeletion()
	} else {
		err = r.ReconcilePhases()
	}
	if err != nil {
		if perr, isPERR := err.(*util.PersistentError); isPERR {
			r.SetPhase(nbv1.BackingStorePhaseRejected, perr.Reason, perr.Message)
			log.Errorf("❌ Persistent Error: %s", err)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.BackingStore, corev1.EventTypeWarning, perr.Reason, perr.Message)
			}
		} else {
			res.RequeueAfter = 3 * time.Second
			// leave current phase as is
			r.SetPhase("", "TemporaryError", err.Error())
			log.Warnf("⏳ Temporary Error: %s", err)
		}
	} else {
		r.SetPhase(
			nbv1.BackingStorePhaseReady,
			"BackingStorePhaseReady",
			"noobaa operator completed reconcile - backing store is ready",
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

	r.SetPhase(
		nbv1.BackingStorePhaseConnecting,
		"BackingStorePhaseConnecting",
		"noobaa operator started phase 2/3 - \"Connecting\"",
	)

	if err := r.ReadSystemInfo(); err != nil {
		return err
	}

	r.SetPhase(
		nbv1.BackingStorePhaseCreating,
		"BackingStorePhaseCreating",
		"noobaa operator started phase 3/3 - \"Creating\"",
	)

	if err := r.ReconcileExternalConnection(); err != nil {
		return err
	}
	if err := r.ReconcilePool(); err != nil {
		return err
	}

	return nil
}

// SetPhase updates the status phase and conditions
func (r *Reconciler) SetPhase(phase nbv1.BackingStorePhase, reason string, message string) {

	c := &r.BackingStore.Status.Conditions

	if phase == "" {
		r.Logger.Infof("SetPhase: temporary error during phase %q", r.BackingStore.Status.Phase)
		util.SetProgressingCondition(c, reason, message)
		return
	}

	r.Logger.Infof("SetPhase: %s", phase)
	r.BackingStore.Status.Phase = phase
	switch phase {
	case nbv1.BackingStorePhaseReady:
		util.SetAvailableCondition(c, reason, message)
	case nbv1.BackingStorePhaseRejected:
		util.SetErrorCondition(c, reason, message)
	default:
		util.SetProgressingCondition(c, reason, message)
	}
}

// UpdateStatus updates the backing store status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() {
	err := r.Client.Status().Update(r.Ctx, r.BackingStore)
	if err != nil {
		r.Logger.Errorf("UpdateStatus: %s", err)
	} else {
		r.Logger.Infof("UpdateStatus: Done")
	}
}

// ReconcilePhaseVerifying checks that we have the system and secret needed to reconcile
func (r *Reconciler) ReconcilePhaseVerifying() error {

	r.SetPhase(
		nbv1.BackingStorePhaseVerifying,
		"BackingStorePhaseVerifying",
		"noobaa operator started phase 1/3 - \"Verifying\"",
	)

	if r.NooBaa.UID == "" {
		return util.NewPersistentError("MissingSystem",
			fmt.Sprintf("NooBaa system %q not found or deleted", r.NooBaa.Name))
	}

	if r.Secret.UID == "" {
		return util.NewPersistentError("MissingSecret",
			fmt.Sprintf("BackingStore Secret %q not found or deleted", r.Secret.Name))
	}

	return nil
}

// ReadSystemInfo loads the information from the noobaa system api,
// and prepares the structures to reconcile
func (r *Reconciler) ReadSystemInfo() error {

	sysClient, err := system.Connect()
	if err != nil {
		return nil
	}
	r.NBClient = sysClient.NBClient

	systemInfo, err := r.NBClient.ReadSystemAPI()
	if err != nil {
		return nil
	}
	r.SystemInfo = &systemInfo

	conn, err := r.MakeExternalConnectionParams()
	if err != nil {
		return err
	}

	// Check if pool exists
	for i := range r.SystemInfo.Pools {
		p := &r.SystemInfo.Pools[i]
		if p.Name == r.BackingStore.Name {
			if p.CloudInfo != nil &&
				p.CloudInfo.EndpointType == conn.EndpointType &&
				p.CloudInfo.Endpoint == conn.Endpoint &&
				p.CloudInfo.Identity == conn.Identity {
				// pool exists and connection match
				r.PoolInfo = p
			} else {
				// TODO pool exists but connection mismatch
				r.Logger.Errorf("pool exists but connection mismatch %+v pool %+v %+v", conn, p, p.CloudInfo)
				r.PoolInfo = p
			}
		}
	}

	// Reuse an existing connection if match is found
	for i := range r.SystemInfo.Accounts {
		account := &r.SystemInfo.Accounts[i]
		for j := range account.ExternalConnections.Connections {
			c := &account.ExternalConnections.Connections[j]
			if c.EndpointType == conn.EndpointType &&
				c.Endpoint == conn.Endpoint &&
				c.Identity == conn.Identity {
				r.ExternalConnectionInfo = c
				conn.Name = c.Name
			}
		}
	}

	r.AddExternalConnectionParams = conn

	r.CreateCloudPoolParams = &nb.CreateCloudPoolParams{
		Name:         r.BackingStore.Name,
		Connection:   conn.Name,
		TargetBucket: r.BackingStore.Spec.BucketName,
	}

	return nil
}

// MakeExternalConnectionParams translates the backing store spec and secret,
// to noobaa api structures to be used for creating/updating external connetion and pool
func (r *Reconciler) MakeExternalConnectionParams() (*nb.AddExternalConnectionParams, error) {

	conn := &nb.AddExternalConnectionParams{
		Name: r.BackingStore.Name,
	}

	// fix keys to handle the case that the secret holds lowercase keys
	if r.Secret.StringData["AWS_ACCESS_KEY_ID"] == "" && r.Secret.StringData["aws_access_key_id"] != "" {
		r.Secret.StringData["AWS_ACCESS_KEY_ID"] = r.Secret.StringData["aws_access_key_id"]
	}
	if r.Secret.StringData["AWS_SECRET_ACCESS_KEY"] == "" && r.Secret.StringData["aws_secret_access_key"] != "" {
		r.Secret.StringData["AWS_SECRET_ACCESS_KEY"] = r.Secret.StringData["aws_secret_access_key"]
	}

	switch r.BackingStore.Spec.Type {

	case nbv1.StoreTypeAWSS3:
		conn.EndpointType = nb.EndpointTypeAws
		conn.Endpoint = r.BackingStore.Spec.S3Options.Endpoint
		proto := "https"
		if r.BackingStore.Spec.S3Options.SSLDisabled {
			proto = "http"
		}
		if conn.Endpoint == "" && r.BackingStore.Spec.S3Options.Region != "" {
			conn.Endpoint = fmt.Sprintf("%s://s3.%s.amazonaws.com", proto, r.BackingStore.Spec.S3Options.Region)
		}
		if conn.Endpoint == "" {
			conn.Endpoint = fmt.Sprintf("%s://s3.amazonaws.com", proto)
		}
		conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]

	case nbv1.StoreTypeS3Compatible:
		conn.EndpointType = nb.EndpointTypeS3Compat
		conn.Endpoint = r.BackingStore.Spec.S3Options.Endpoint
		conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]
		// conn.AuthMethod =

	case nbv1.StoreTypeAzureBlob:
		return nil, util.NewPersistentError("NotYetImplemented",
			fmt.Sprintf("Not yet implemented backing store type %q", r.BackingStore.Spec.Type))
		// conn.EndpointType = nb.EndpointTypeAzure
		// conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		// conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]

	case nbv1.StoreTypeGoogleCloudStorage:
		return nil, util.NewPersistentError("NotYetImplemented",
			fmt.Sprintf("Not yet implemented backing store type %q", r.BackingStore.Spec.Type))
		// conn.EndpointType = nb.EndpointTypeGoogle
		// conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		// conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]

	case nbv1.StoreTypePV:
		return nil, util.NewPersistentError("NotYetImplemented",
			fmt.Sprintf("Not yet implemented backing store type %q", r.BackingStore.Spec.Type))

	default:
		return nil, util.NewPersistentError("InvalidType",
			fmt.Sprintf("Invalid backing store type %q", r.BackingStore.Spec.Type))
	}

	return conn, nil
}

// ReconcileExternalConnection handles the external connection using noobaa api
func (r *Reconciler) ReconcileExternalConnection() error {

	if r.ExternalConnectionInfo != nil {
		return nil
	}

	res, err := r.NBClient.CheckExternalConnectionAPI(*r.AddExternalConnectionParams)
	if err != nil {
		if rpcErr, isRPCErr := err.(*nb.RPCError); isRPCErr {
			if rpcErr.RPCCode == "INVALID_SCHEMA_PARAMS" {
				return util.NewPersistentError("InvalidConnectionParams", rpcErr.Message)
			}
		}
		return err
	}
	switch res.Status {

	case nb.ExternalConnectionSuccess:
		// good

	case nb.ExternalConnectionInvalidCredentials:
		if time.Since(r.BackingStore.CreationTimestamp.Time) < 5*time.Minute {
			r.Logger.Infof("got invalid credentials. sometimes access keys take time to propagate inside AWS. requeuing for 5 minutes")
			return fmt.Errorf("Got InvalidCredentials. requeue again")
		}
		fallthrough
	case nb.ExternalConnectionInvalidEndpoint:
		fallthrough
	case nb.ExternalConnectionTimeSkew:
		fallthrough
	case nb.ExternalConnectionNotSupported:
		return util.NewPersistentError(string(res.Status),
			fmt.Sprintf("BackingStore %q invalid external connection %q", r.BackingStore.Name, res.Status))
	case nb.ExternalConnectionTimeout:
		fallthrough
	case nb.ExternalConnectionUnknownFailure:
		fallthrough

	default:
		return fmt.Errorf("CheckExternalConnection Status=%s Error=%s Message=%s",
			res.Status, res.Error.Code, res.Error.Message)
	}

	err = r.NBClient.AddExternalConnectionAPI(*r.AddExternalConnectionParams)
	if err != nil {
		return err
	}

	return nil
}

// ReconcilePool handles the pool using noobaa api
func (r *Reconciler) ReconcilePool() error {

	if r.PoolInfo != nil {
		return nil
	}

	err := r.NBClient.CreateCloudPoolAPI(*r.CreateCloudPoolParams)
	if err != nil {
		return err
	}

	return nil
}

// ReconcileDeletion handles the deletion of a backing-store using the noobaa api
func (r *Reconciler) ReconcileDeletion() error {

	// Set the phase to let users know the operator has noticed the deletion request
	if r.BackingStore.Status.Phase != nbv1.BackingStorePhaseDeleting {
		r.SetPhase(
			nbv1.BackingStorePhaseDeleting,
			"BackingStorePhaseDeleting",
			"noobaa operator started deletion",
		)
		r.UpdateStatus()
	}

	if r.NooBaa.UID == "" {
		r.Logger.Infof("BackingStore %q remove finalizer because NooBaa system is already deleted", r.BackingStore.Name)
		return r.FinalizeDeletion()
	}

	if err := r.ReadSystemInfo(); err != nil {
		return err
	}

	if r.PoolInfo != nil {
		internalPoolName := ""
		for i := range r.SystemInfo.Pools {
			pool := &r.SystemInfo.Pools[i]
			if pool.ResourceType == "INTERNAL" {
				internalPoolName = pool.Name
				break
			}
		}
		for i := range r.SystemInfo.Accounts {
			account := &r.SystemInfo.Accounts[i]
			if account.DefaultPool == r.PoolInfo.Name {
				allowedBuckets := account.AllowedBuckets
				if allowedBuckets.PermissionList == nil {
					allowedBuckets.PermissionList = []string{}
				}
				err := r.NBClient.UpdateAccountS3Access(nb.UpdateAccountS3AccessParams{
					Email:        account.Email,
					S3Access:     account.HasS3Access,
					DefaultPool:  &internalPoolName,
					AllowBuckets: &allowedBuckets,
				})
				if err != nil {
					return err
				}
			}
		}
		err := r.NBClient.DeletePoolAPI(nb.DeletePoolParams{Name: r.PoolInfo.Name})
		if err != nil {
			return err
		}
	}

	if r.ExternalConnectionInfo != nil {
		// TODO we cannot assume we are the only one using this connection...
		err := r.NBClient.DeleteExternalConnectionAPI(nb.DeleteExternalConnectionParams{Name: r.ExternalConnectionInfo.Name})
		if err != nil {
			return err
		}
	}

	return r.FinalizeDeletion()
}

// FinalizeDeletion removed the finalizer and updates in order to let the backing-store get reclaimed by kubernetes
func (r *Reconciler) FinalizeDeletion() error {
	util.RemoveFinalizer(r.BackingStore, nbv1.BackingStoreFinalizer)
	if !util.KubeUpdate(r.BackingStore) {
		return fmt.Errorf("BackingStore %q failed to remove finalizer %q", r.BackingStore.Name, nbv1.BackingStoreFinalizer)
	}
	return nil
}
