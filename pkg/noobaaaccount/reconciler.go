package noobaaaccount

import (
	"context"
	"fmt"
	"reflect"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

	NBClient          nb.Client
	SystemInfo        *nb.SystemInfo
	NooBaaAccountInfo *nb.AccountInfo

	NooBaaAccount *nbv1.NooBaaAccount
	NooBaa        *nbv1.NooBaa

	Secret *corev1.Secret
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a noobaa account
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:       req,
		Client:        client,
		Scheme:        scheme,
		Recorder:      recorder,
		Ctx:           context.TODO(),
		Logger:        logrus.WithField("noobaaaccount", req.Namespace+"/"+req.Name),
		NooBaaAccount: util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml).(*nbv1.NooBaaAccount),
		NooBaa:        util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		Secret:        util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
	}

	// Set Namespace
	r.NooBaaAccount.Namespace = r.Request.Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.Secret.Namespace = r.Request.Namespace

	// Set Names
	r.NooBaaAccount.Name = r.Request.Name
	r.NooBaa.Name = options.SystemName
	r.Secret.Name = fmt.Sprintf("noobaa-account-%s", r.Request.Name)

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

	util.KubeCheck(r.NooBaaAccount)

	if r.NooBaaAccount.UID == "" {
		log.Infof("NooBaaAccount %q not found or deleted. Skip reconcile.", r.NooBaaAccount.Name)
		return reconcile.Result{}, nil
	}

	if util.EnsureCommonMetaFields(r.NooBaaAccount, nbv1.Finalizer) {
		if !util.KubeUpdate(r.NooBaaAccount) {
			log.Errorf("❌ NooBaaAccount %q failed to add mandatory meta fields", r.NooBaaAccount.Name)

			res.RequeueAfter = 3 * time.Second
			return res, nil
		}
	}

	system.CheckSystem(r.NooBaa)

	var err error
	if r.NooBaaAccount.DeletionTimestamp != nil {
		err = r.ReconcileDeletion()
	} else {
		err = r.ReconcilePhases()
	}
	if err != nil {
		if perr, isPERR := err.(*util.PersistentError); isPERR {
			r.SetPhase(nbv1.NooBaaAccountPhaseRejected, perr.Reason, perr.Message)
			log.Errorf("❌ Persistent Error: %s", err)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.NooBaaAccount, corev1.EventTypeWarning, perr.Reason, perr.Message)
			}
		} else {
			res.RequeueAfter = 3 * time.Second
			// leave current phase as is
			r.SetPhase("", "TemporaryError", err.Error())
			log.Warnf("⏳ Temporary Error: %s", err)
		}
	} else {
		r.SetPhase(
			nbv1.NooBaaAccountPhaseReady,
			"NooBaaAccountPhaseReady",
			"noobaa operator completed reconcile - noobaa account is ready",
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

// ReconcilePhases runs the reconcile flow and populates System.Status.
func (r *Reconciler) ReconcilePhases() error {

	if err := r.ReconcilePhaseVerifying(); err != nil {
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
func (r *Reconciler) SetPhase(phase nbv1.NooBaaAccountPhase, reason string, message string) {

	c := &r.NooBaaAccount.Status.Conditions

	if phase == "" {
		r.Logger.Infof("SetPhase: temporary error during phase %q", r.NooBaaAccount.Status.Phase)
		util.SetProgressingCondition(c, reason, message)
		return
	}

	r.Logger.Infof("SetPhase: %s", phase)
	r.NooBaaAccount.Status.Phase = phase
	switch phase {
	case nbv1.NooBaaAccountPhaseReady:
		util.SetAvailableCondition(c, reason, message)
	case nbv1.NooBaaAccountPhaseRejected:
		util.SetErrorCondition(c, reason, message)
	default:
		util.SetProgressingCondition(c, reason, message)
	}
}

// UpdateStatus updates the noobaa account status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() error {
	err := r.Client.Status().Update(r.Ctx, r.NooBaaAccount)
	if err != nil {
		r.Logger.Errorf("UpdateStatus: %s", err)
		return err
	}
	r.Logger.Infof("UpdateStatus: Done")
	return nil
}

// ReconcilePhaseVerifying checks that we have the system and secret needed to reconcile
func (r *Reconciler) ReconcilePhaseVerifying() error {

	r.SetPhase(
		nbv1.NooBaaAccountPhaseVerifying,
		"NooBaaAccountPhaseVerifying",
		"noobaa operator started phase 1/2 - \"Verifying\"",
	)

	if r.NooBaa.UID == "" {
		return util.NewPersistentError("MissingSystem",
			fmt.Sprintf("NooBaa system %q not found or deleted", r.NooBaa.Name))
	}

	if r.NooBaaAccount.Spec.AllowBucketCreate && r.NooBaaAccount.Spec.DefaultResource == "" {
		return util.NewPersistentError("MissingDefaultResource",
			fmt.Sprintf("Account %q is allowed to create buckets, but no resource is provided", r.NooBaaAccount.Name))
	}

	if r.NooBaaAccount.Spec.DefaultResource != "" {
		isResourceBackingStore := checkResourceBackingStore(r.NooBaaAccount.Spec.DefaultResource)
		isResourceNamespaceStore := checkResourceNamespaceStore(r.NooBaaAccount.Spec.DefaultResource)
		if !isResourceBackingStore && !isResourceNamespaceStore {
			return util.NewPersistentError("MissingDefaultResource",
				fmt.Sprintf("Account %q is allowed to create buckets, but resource %q was not found",
					r.NooBaaAccount.Name, r.NooBaaAccount.Spec.DefaultResource))
		} else if isResourceBackingStore && isResourceNamespaceStore {
			return util.NewPersistentError("MissingDefaultResource",
				fmt.Sprintf("BackingStore and NamespaceStore should not have the same name: %q, ", r.NooBaaAccount.Spec.DefaultResource))
		}
	}

	return nil
}

// ReconcilePhaseConnecting checks that we have the system and secret needed to reconcile
func (r *Reconciler) ReconcilePhaseConnecting() error {

	r.SetPhase(
		nbv1.NooBaaAccountPhaseConnecting,
		"NooBaaAccountPhaseConnecting",
		"noobaa operator started phase 2/3 - \"Connecting\"",
	)

	if err := r.ReadSystemInfo(); err != nil {
		return err
	}

	return nil
}

// ReconcilePhaseConfiguring create orupdates existing accounts to match the changes in noobaaaccount
func (r *Reconciler) ReconcilePhaseConfiguring() error {

	r.SetPhase(
		nbv1.NooBaaAccountPhaseConfiguring,
		"NooBaaAccountPhaseConfiguring",
		"noobaa operator started phase 2/2 - \"Configuring\"",
	)

	sysClient, err := system.Connect(false)
	if err != nil {
		return err
	}
	r.NBClient = sysClient.NBClient

	if r.NooBaaAccountInfo != nil {
		if err := r.UpdateNooBaaAccount(); err != nil {
			return err
		}
	} else {
		if err := r.CreateNooBaaAccount(); err != nil {
			return err
		}
	}

	return nil
}

// ReconcileDeletion handles the deletion of a noobaa account using the noobaa api
func (r *Reconciler) ReconcileDeletion() error {

	// Set the phase to let users know the operator has noticed the deletion request
	if r.NooBaaAccount.Status.Phase != nbv1.NooBaaAccountPhaseDeleting {
		r.SetPhase(
			nbv1.NooBaaAccountPhaseDeleting,
			"BucketClassPhaseDeleting",
			"noobaa operator started deletion",
		)
		err := r.UpdateStatus()
		if err != nil {
			return err
		}
	}

	if r.NooBaa.UID == "" {
		r.Logger.Infof("NooBaaAccount %q remove finalizer because NooBaa system is already deleted", r.NooBaaAccount.Name)
		return r.FinalizeDeletion()
	}

	sysClient, err := system.Connect(false)
	if err != nil {
		return err
	}
	r.NBClient = sysClient.NBClient
	err = r.NBClient.DeleteAccountAPI(nb.DeleteAccountParams{Email: r.NooBaaAccount.Name})
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_ACCOUNT" {
			r.Logger.Warnf("Account to delete was not found %q", r.NooBaaAccount.Name)
		} else {
			return fmt.Errorf("failed to delete account %q. got error: %v", r.NooBaaAccount.Name, err)
		}
	} else {
		r.Logger.Infof("✅ Successfully deleted account %q", r.NooBaaAccount.Name)
	}
	return r.FinalizeDeletion()
}

// FinalizeDeletion removed the finalizer and updates in order to let the bucket class get reclaimed by kubernetes
func (r *Reconciler) FinalizeDeletion() error {
	util.RemoveFinalizer(r.NooBaaAccount, nbv1.Finalizer)
	if !util.KubeUpdate(r.NooBaaAccount) {
		return fmt.Errorf("BucketClass %q failed to remove finalizer %q", r.NooBaaAccount.Name, nbv1.Finalizer)
	}
	return nil
}

// CreateNooBaaAccount creates a new noobaa account
func (r *Reconciler) CreateNooBaaAccount() error {
	log := r.Logger

	if r.NooBaaAccount == nil {
		return fmt.Errorf("NooBaaAccount not loaded %#v", r)
	}

	createAccountParams := nb.CreateAccountParams{
		Name:              r.NooBaaAccount.Name,
		Email:             r.NooBaaAccount.Name,
		DefaultResource:   r.NooBaaAccount.Spec.DefaultResource,
		HasLogin:          false,
		S3Access:          true,
		AllowBucketCreate: r.NooBaaAccount.Spec.AllowBucketCreate,
	}

	if r.NooBaaAccount.Spec.NsfsAccountConfig != nil {
		createAccountParams.NsfsAccountConfig = &nbv1.AccountNsfsConfig{
			UID:            r.NooBaaAccount.Spec.NsfsAccountConfig.UID,
			GID:            r.NooBaaAccount.Spec.NsfsAccountConfig.GID,
			NewBucketsPath: r.NooBaaAccount.Spec.NsfsAccountConfig.NewBucketsPath,
			NsfsOnly:       r.NooBaaAccount.Spec.NsfsAccountConfig.NsfsOnly,
		}
	}

	accountInfo, err := r.NBClient.CreateAccountAPI(createAccountParams)
	if err != nil {
		return err
	}

	var accessKeys nb.S3AccessKeys
	// if we didn't get the access keys in the create_account reply we might be talking to an older noobaa version (prior to 5.1)
	// in that case try to get it using read account
	if len(accountInfo.AccessKeys) == 0 {
		log.Info("CreateAccountAPI did not return access keys. calling ReadAccountAPI to get keys..")
		readAccountReply, err := r.NBClient.ReadAccountAPI(nb.ReadAccountParams{Email: r.NooBaaAccount.Name})
		if err != nil {
			return err
		}
		accessKeys = readAccountReply.AccessKeys[0]
	} else {
		accessKeys = accountInfo.AccessKeys[0]
	}
	r.Secret.StringData = map[string]string{}
	r.Secret.StringData["AWS_ACCESS_KEY_ID"] = accessKeys.AccessKey
	r.Secret.StringData["AWS_SECRET_ACCESS_KEY"] = accessKeys.SecretKey
	r.Own(r.Secret)
	err = r.Client.Create(r.Ctx, r.Secret)
	if err != nil {
		r.Logger.Errorf("got error on NooBaaAccount creation. error: %v", err)
		return err
	}

	log.Infof("✅ Successfully created account %q", r.NooBaaAccount.Name)
	return nil
}

// UpdateNooBaaAccount update an existing noobaa account
func (r *Reconciler) UpdateNooBaaAccount() error {
	log := r.Logger

	if r.NooBaaAccount == nil {
		return fmt.Errorf("NooBaaAccount not loaded %#v", r)
	}

	if r.needUpdate() {

		updateAccountS3AccessParams := nb.UpdateAccountS3AccessParams{
			Email:               r.NooBaaAccount.Name,
			DefaultResource:     &r.NooBaaAccount.Spec.DefaultResource,
			S3Access:            true,
			AllowBucketCreation: &r.NooBaaAccount.Spec.AllowBucketCreate,
		}

		if r.NooBaaAccount.Spec.NsfsAccountConfig != nil {
			updateAccountS3AccessParams.NsfsAccountConfig = &nbv1.AccountNsfsConfig{
				UID:            r.NooBaaAccount.Spec.NsfsAccountConfig.UID,
				GID:            r.NooBaaAccount.Spec.NsfsAccountConfig.GID,
				NewBucketsPath: r.NooBaaAccount.Spec.NsfsAccountConfig.NewBucketsPath,
				NsfsOnly:       r.NooBaaAccount.Spec.NsfsAccountConfig.NsfsOnly,
			}
		}

		err := r.NBClient.UpdateAccountS3Access(updateAccountS3AccessParams)
		if err != nil {
			return err
		}
		log.Infof("✅ Successfully updated account %q", r.NooBaaAccount.Name)
	}

	return nil
}

func (r *Reconciler) needUpdate() bool {
	return r.NooBaaAccount.Spec.AllowBucketCreate != r.NooBaaAccountInfo.CanCreateBuckets ||
		r.NooBaaAccount.Spec.DefaultResource != r.NooBaaAccountInfo.DefaultResource ||
		!reflect.DeepEqual(r.NooBaaAccount.Spec.NsfsAccountConfig, r.NooBaaAccountInfo.NsfsAccountConfig) ||
		r.NooBaaAccount.Spec.NsfsAccountConfig != nil && r.NooBaaAccountInfo.NsfsAccountConfig == nil
}

// ReadSystemInfo loads the information from the noobaa system api,
// and prepares the structures to reconcile
func (r *Reconciler) ReadSystemInfo() error {

	sysClient, err := system.Connect(false)
	if err != nil {
		return err
	}
	r.NBClient = sysClient.NBClient

	systemInfo, err := r.NBClient.ReadSystemAPI()
	if err != nil {
		return err
	}
	r.SystemInfo = &systemInfo

	// Check if account exists
	for i := range r.SystemInfo.Accounts {
		a := &r.SystemInfo.Accounts[i]
		if a.Name == r.NooBaaAccount.Name {
			r.NooBaaAccountInfo = a
			break
		}
	}

	return nil
}

// Own sets the object owner references to the noobaaAccount
func (r *Reconciler) Own(obj metav1.Object) {
	util.Panic(controllerutil.SetControllerReference(r.NooBaaAccount, obj, r.Scheme))
}

// IsBucketInBucketsArray returns true if a bucket array contains a bucket with a spesific name.
func IsBucketInBucketsArray(buckets []nb.BucketInfo, bucketName string) bool {
	for _, b := range buckets {
		if b.Name == bucketName {
			return true
		}
	}
	return false
}
