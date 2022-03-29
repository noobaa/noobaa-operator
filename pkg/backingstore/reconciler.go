package backingstore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"regexp"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"

	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ModeInfo holds local information for a backing store mode.
type ModeInfo struct {
	Phase    nbv1.BackingStorePhase
	Severity string
}

var bsModeInfoMap map[string]ModeInfo

func init() {
	bsModeInfoMap = modeInfoMap()
}

func modeInfoMap() map[string]ModeInfo {
	return map[string]ModeInfo{
		"INITIALIZING":        {nbv1.BackingStorePhaseCreating, corev1.EventTypeNormal},
		"DELETING":            {nbv1.BackingStorePhaseReady, corev1.EventTypeNormal},
		"SCALING":             {nbv1.BackingStorePhaseReady, corev1.EventTypeNormal},
		"MOST_NODES_ISSUES":   {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"MANY_NODES_ISSUES":   {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"MOST_STORAGE_ISSUES": {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"MANY_STORAGE_ISSUES": {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"MANY_NODES_OFFLINE":  {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"LOW_CAPACITY":        {nbv1.BackingStorePhaseReady, corev1.EventTypeWarning},
		"OPTIMAL":             {nbv1.BackingStorePhaseReady, corev1.EventTypeNormal},
		"HAS_NO_NODES":        {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
		"ALL_NODES_OFFLINE":   {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
		"NO_CAPACITY":         {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
		"IO_ERRORS":           {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
		"STORAGE_NOT_EXIST":   {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
		"AUTH_FAILED":         {nbv1.BackingStorePhaseRejected, corev1.EventTypeWarning},
	}
}

// Reconciler is the context for loading or reconciling a noobaa system
type Reconciler struct {
	Request  types.NamespacedName
	Client   client.Client
	Scheme   *runtime.Scheme
	Ctx      context.Context
	Logger   *logrus.Entry
	Recorder record.EventRecorder
	NBClient nb.Client

	BackingStore     *nbv1.BackingStore
	NooBaa           *nbv1.NooBaa
	Secret           *corev1.Secret
	PodAgentTemplate *corev1.Pod
	PvcAgentTemplate *corev1.PersistentVolumeClaim
	ServiceAccount   *corev1.ServiceAccount

	SystemInfo             *nb.SystemInfo
	ExternalConnectionInfo *nb.ExternalConnectionInfo
	PoolInfo               *nb.PoolInfo
	HostsInfo              *[]nb.HostInfo

	AddExternalConnectionParams    *nb.AddExternalConnectionParams
	CreateCloudPoolParams          *nb.CreateCloudPoolParams
	CreateHostsPoolParams          *nb.CreateHostsPoolParams
	UpdateHostsPoolParams          *nb.UpdateHostsPoolParams
	UpdateExternalConnectionParams *nb.UpdateExternalConnectionParams
}

// Own sets the object owner references to the backingstore
func (r *Reconciler) Own(obj metav1.Object) {
	util.Panic(controllerutil.SetControllerReference(r.BackingStore, obj, r.Scheme))
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a backing store
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:          req,
		Client:           client,
		Scheme:           scheme,
		Recorder:         recorder,
		Ctx:              context.TODO(),
		Logger:           logrus.WithField("backingstore", req.Namespace+"/"+req.Name),
		BackingStore:     util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore),
		NooBaa:           util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		Secret:           util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
		ServiceAccount:   util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount),
		PodAgentTemplate: util.KubeObject(bundle.File_deploy_internal_pod_agent_yaml).(*corev1.Pod),
		PvcAgentTemplate: util.KubeObject(bundle.File_deploy_internal_pvc_agent_yaml).(*corev1.PersistentVolumeClaim),
	}

	// Set Namespace
	r.BackingStore.Namespace = r.Request.Namespace
	r.NooBaa.Namespace = r.Request.Namespace
	r.ServiceAccount.Namespace = r.Request.Namespace

	// Set Names
	r.BackingStore.Name = r.Request.Name
	r.NooBaa.Name = options.SystemName
	r.ServiceAccount.Name = options.SystemName

	// Set secret names to empty
	r.Secret.Namespace = ""
	r.Secret.Name = ""

	return r
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {
	log := r.Logger
	log.Infof("Start BackingStore Reconcile ...")

	systemFound := system.CheckSystem(r.NooBaa)

	if !util.KubeCheck(r.BackingStore) {
		log.Infof("❌ BackingStore %q not found.", r.BackingStore.Name)
		return reconcile.Result{}, nil // final state
	}

	if err := r.LoadBackingStoreSecret(); err != nil {
		return r.completeReconcile(err)
	}

	if ts := r.BackingStore.DeletionTimestamp; ts != nil {
		log.Infof("BackingStore %q was deleted on %v.", r.BackingStore.Name, ts)
		err := r.ReconcileDeletion(systemFound)
		return r.completeReconcile(err)
	}

	if !systemFound {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return r.completeReconcile(nil)
	}

	if util.EnsureCommonMetaFields(r.BackingStore, nbv1.Finalizer) {
		if !util.KubeUpdate(r.BackingStore) {
			err := fmt.Errorf("❌ BackingStore %q failed to add mandatory meta fields", r.BackingStore.Name)
			return r.completeReconcile(err)
		}
	}

	oldStatefulSet := &appsv1.StatefulSet{}
	oldStatefulSet.Name = fmt.Sprintf("%s-%s-noobaa", r.BackingStore.Name, options.SystemName)
	oldStatefulSet.Namespace = r.Request.Namespace

	if util.KubeCheck(oldStatefulSet) {
		if err := r.upgradeBackingStore(oldStatefulSet); err != nil {
			return r.completeReconcile(err)
		}
	}

	err := r.ReconcilePhases()
	return r.completeReconcile(err)
}

func (r *Reconciler) completeReconcile(err error) (reconcile.Result, error) {
	log := r.Logger
	res := reconcile.Result{}

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
		mode := r.BackingStore.Status.Mode.ModeCode
		phaseInfo, exist := bsModeInfoMap[mode]

		if exist && phaseInfo.Phase != r.BackingStore.Status.Phase {
			phaseName := fmt.Sprintf("BackingStorePhase%s", phaseInfo.Phase)
			desc := fmt.Sprintf("Backing store mode: %s", mode)
			r.SetPhase(phaseInfo.Phase, desc, phaseName)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.BackingStore, phaseInfo.Severity, phaseName, desc)
			}
		} else {
			r.SetPhase(
				nbv1.BackingStorePhaseReady,
				"BackingStorePhaseReady",
				"noobaa operator completed reconcile - backing store is ready",
			)
			log.Infof("✅ Done")
		}
	}

	err = r.UpdateStatus()
	// if updateStatus will fail to update the CR for any reason we will continue to requeue the reconcile
	// until the spec status will reflect the actual status of the backingstore
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
	if err := r.ReconcilePhaseCreating(); err != nil {
		return err
	}

	return nil
}

// LoadBackingStoreSecret loads the secret to the reconciler struct
func (r *Reconciler) LoadBackingStoreSecret() error {
	if !util.IsSTSClusterBS(r.BackingStore) {
		secretRef, err := util.GetBackingStoreSecret(r.BackingStore)
		if err != nil {
			return err
		}

		if secretRef != nil {
			r.Secret.Name = secretRef.Name
			r.Secret.Namespace = secretRef.Namespace

			if r.Secret.Namespace == "" {
				r.Secret.Namespace = r.BackingStore.Namespace
			}

			if r.Secret.Name == "" {
				if r.BackingStore.Spec.Type != nbv1.StoreTypePVPool {
					return util.NewPersistentError("EmptySecretName",
						"BackingStore Secret reference has an empty name")
				}
				r.Secret.Name = fmt.Sprintf("backing-store-%s-%s", nbv1.StoreTypePVPool, r.BackingStore.Name)
				r.Secret.Namespace = r.BackingStore.Namespace
				r.Secret.StringData = map[string]string{}
				r.Secret.Data = nil

				if !util.KubeCheck(r.Secret) {
					r.Own(r.Secret)
					if !util.KubeCreateFailExisting(r.Secret) {
						return util.NewPersistentError("EmptySecretName",
							fmt.Sprintf("Could not create Secret %q in Namespace %q (conflict)", r.Secret.Name, r.Secret.Namespace))
					}
				}
			} else {
				// check the existence of another secret in the system that contains the same credentials,
				// if found, point this BS secret reference to it.
				// so if the user will update the credentials, it will trigger updateExternalConnection in all the Backingstores
				secret, err := util.GetSecretFromSecretReference(secretRef)
				if err != nil {
					return nil
				}
				if secret != nil {
					suggestedSecret := util.CheckForIdenticalSecretsCreds(secret, util.MapStorTypeToMandatoryProperties[r.BackingStore.Spec.Type])
					if suggestedSecret != nil {
						secretRef.Name = suggestedSecret.Name
						secretRef.Namespace = suggestedSecret.Namespace
						err := util.SetBackingStoreSecretRef(r.BackingStore, secretRef)
						if err != nil {
							return err
						}
						if !util.KubeUpdate(r.BackingStore) {
							return fmt.Errorf("failed to update Backingstore: %q secret reference", r.BackingStore.Name)
						}
						secret = suggestedSecret
					}
					err = util.SetOwnerReference(r.BackingStore, secret, r.Scheme)
					if _, isAlreadyOwnedErr := err.(*controllerutil.AlreadyOwnedError); !isAlreadyOwnedErr {
						if err == nil {
							if !util.KubeUpdate(secret) {
								return fmt.Errorf("failed to update secret: %q owner reference", r.BackingStore.Name)
							}
						} else {
							return err
						}
					}
				}
			}
		}

		util.KubeCheck(r.Secret)
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
func (r *Reconciler) UpdateStatus() error {
	err := r.Client.Status().Update(r.Ctx, r.BackingStore)
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
		nbv1.BackingStorePhaseVerifying,
		"BackingStorePhaseVerifying",
		"noobaa operator started phase 1/3 - \"Verifying\"",
	)

	err := validations.ValidateBackingStore(*r.BackingStore)
	if err != nil {
		return util.NewPersistentError("BackingStoreValidationError", err.Error())
	}

	if r.NooBaa.UID == "" {
		return util.NewPersistentError("MissingSystem",
			fmt.Sprintf("NooBaa system %q not found or deleted", r.NooBaa.Name))
	}

	if r.Secret.Name != "" && r.Secret.UID == "" {
		if time.Since(r.BackingStore.CreationTimestamp.Time) < 5*time.Minute {
			return fmt.Errorf("BackingStore Secret %q not found, but not rejecting the young as it might be in process", r.Secret.Name)
		}
		return util.NewPersistentError("MissingSecret",
			fmt.Sprintf("BackingStore Secret %q not found or deleted", r.Secret.Name))
	}

	return nil
}

// ReconcilePhaseConnecting checks that we have the system and secret needed to reconcile
func (r *Reconciler) ReconcilePhaseConnecting() error {

	r.SetPhase(
		nbv1.BackingStorePhaseConnecting,
		"BackingStorePhaseConnecting",
		"noobaa operator started phase 2/3 - \"Connecting\"",
	)

	if err := r.ReadSystemInfo(); err != nil {
		return err
	}

	return nil
}

// ReconcilePhaseCreating checks that we have the system and secret needed to reconcile
func (r *Reconciler) ReconcilePhaseCreating() error {

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

// finalizeCore runs when the backing store is being deleted
// Handles NooBaa core side of the store deletion
func (r *Reconciler) finalizeCore() error {

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
			if account.DefaultResource == r.PoolInfo.Name {
				allowedBuckets := account.AllowedBuckets
				if allowedBuckets.PermissionList == nil {
					allowedBuckets.PermissionList = []string{}
				}
				err := r.NBClient.UpdateAccountS3Access(nb.UpdateAccountS3AccessParams{
					Email:           account.Email,
					S3Access:        account.HasS3Access,
					DefaultResource: &internalPoolName,
					AllowBuckets:    &allowedBuckets,
				})
				if err != nil {
					return err
				}
			}
		}
		err := r.NBClient.DeletePoolAPI(nb.DeletePoolParams{Name: r.PoolInfo.Name})
		if err != nil {
			if rpcErr, isRPCErr := err.(*nb.RPCError); isRPCErr {
				if rpcErr.RPCCode == "DEFAULT_RESOURCE" {
					return util.NewPersistentError("DefaultResource",
						fmt.Sprintf("DeletePoolAPI cannot complete because pool %q is an account default resource", r.PoolInfo.Name))
				}
				if rpcErr.RPCCode == "IN_USE" {
					return fmt.Errorf("DeletePoolAPI cannot complete because pool %q has buckets attached", r.PoolInfo.Name)
				}
			}
			return err
		}
	}

	if r.ExternalConnectionInfo != nil {
		// TODO we cannot assume we are the only one using this connection...
		err := r.NBClient.DeleteExternalConnectionAPI(nb.DeleteExternalConnectionParams{Name: r.ExternalConnectionInfo.Name})
		if err != nil {
			if rpcErr, isRPCErr := err.(*nb.RPCError); isRPCErr {
				if rpcErr.RPCCode != "IN_USE" {
					return err
				}
				r.Logger.Warnf("DeleteExternalConnection cannot complete because it is IN_USE %q", r.ExternalConnectionInfo.Name)
			} else {
				return err
			}
		}
	}

	// success
	return nil
}

// ReconcileDeletion handles the deletion of a backing-store using the noobaa api
func (r *Reconciler) ReconcileDeletion(systemFound bool) error {

	// Set the phase to let users know the operator has noticed the deletion request
	if r.BackingStore.Status.Phase != nbv1.BackingStorePhaseDeleting {
		r.SetPhase(
			nbv1.BackingStorePhaseDeleting,
			"BackingStorePhaseDeleting",
			"noobaa operator started deletion",
		)
		err := r.UpdateStatus()
		if err != nil {
			return err
		}
	}

	// Notify the NooBaa core if the system is running
	if systemFound {
		if err := r.finalizeCore(); err != nil {
			return err
		}
	}

	// Release the k8s volumes used
	if r.BackingStore.Spec.Type == nbv1.StoreTypePVPool {
		err := r.deletePvPool()
		if err != nil {
			return err
		}
	}

	r.Logger.Infof("BackingStore %q remove finalizer", r.BackingStore.Name)
	return r.FinalizeDeletion()
}

// FinalizeDeletion removed the finalizer and updates in order to let the backing-store get reclaimed by kubernetes
func (r *Reconciler) FinalizeDeletion() error {
	util.RemoveFinalizer(r.BackingStore, nbv1.Finalizer)
	if !util.KubeUpdate(r.BackingStore) {
		return fmt.Errorf("BackingStore %q failed to remove finalizer %q", r.BackingStore.Name, nbv1.Finalizer)
	}
	return nil
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

	// Check if pool exists
	for i := range r.SystemInfo.Pools {
		p := &r.SystemInfo.Pools[i]
		if p.Name == r.BackingStore.Name {
			r.PoolInfo = p
			break
		}
	}

	pool := r.PoolInfo
	if r.BackingStore.Spec.Type == nbv1.StoreTypePVPool {
		if pool != nil && pool.ResourceType != "HOSTS" {
			return util.NewPersistentError("InvalidBackingStore", fmt.Sprintf(
				"BackingStore %q w/existing pool %+v has unexpected resource type %+v",
				r.BackingStore.Name, pool, pool.ResourceType,
			))
		}

		const defaultVolumeSize = int64(20 * 1024 * 1024 * 1024) // 20Gi=20*1024^3
		var volumeSize int64
		pvPool := r.BackingStore.Spec.PVPool
		if pvPool.VolumeResources != nil {
			qty := pvPool.VolumeResources.Requests[corev1.ResourceName(corev1.ResourceStorage)]
			volumeSize = qty.Value()
		} else {
			volumeSize = int64(defaultVolumeSize)
		}

		if pool == nil {
			r.CreateHostsPoolParams = &nb.CreateHostsPoolParams{
				Name:       r.BackingStore.Name,
				IsManaged:  true,
				HostCount:  int(pvPool.NumVolumes),
				HostConfig: nb.PoolHostsInfo{VolumeSize: volumeSize},
				Backingstore: &nb.BackingStoreInfo{
					Name:      r.BackingStore.Name,
					Namespace: r.NooBaa.Namespace,
				},
			}
			r.HostsInfo = &[]nb.HostInfo{}
		} else {
			hostsInfo, err := r.NBClient.ListHostsAPI(nb.ListHostsParams{Query: nb.ListHostsQuery{Pools: []string{r.BackingStore.Name}}})
			if err != nil {
				return err
			}

			if len(hostsInfo.Hosts) > pvPool.NumVolumes { // scaling down - not supported
				return util.NewPersistentError("InvalidBackingStore",
					"Scaling down the number of nodes is not currently supported")
			}
			if pvPool.NumVolumes != int(pool.Hosts.ConfiguredCount) {
				r.UpdateHostsPoolParams = &nb.UpdateHostsPoolParams{ // update core
					Name: r.BackingStore.Name,
				}
			}
			r.HostsInfo = &hostsInfo.Hosts
		}

		return nil
	}

	if pool != nil && pool.ResourceType != "CLOUD" {
		return util.NewPersistentError("InvalidBackingStore", fmt.Sprintf(
			"BackingStore %q w/existing pool %+v has unexpected resource type %+v",
			r.BackingStore.Name, pool, pool.ResourceType,
		))
	}

	conn, err := r.MakeExternalConnectionParams()
	if err != nil {
		return err
	}

	// Check that noobaa-core uses the same connection as the pool
	// Due to noobaa/noobaa-core#5750 the identity (access-key) is not returned in the api call so just warn for now
	// TODO Improve handling of this condition
	if pool != nil {
		if pool.CloudInfo == nil ||
			pool.CloudInfo.EndpointType != conn.EndpointType ||
			pool.CloudInfo.Endpoint != conn.Endpoint ||
			pool.CloudInfo.Identity != conn.Identity {
			r.Logger.Warnf("using existing pool but connection mismatch %+v pool %+v %+v", conn, pool, pool.CloudInfo)
			r.UpdateExternalConnectionParams = &nb.UpdateExternalConnectionParams{
				Name:     conn.Name,
				Identity: conn.Identity,
				Secret:   conn.Secret,
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

	targetBucket, err := util.GetBackingStoreTargetBucket(r.BackingStore)
	if err != nil {
		return err
	}

	r.CreateCloudPoolParams = &nb.CreateCloudPoolParams{
		Name:         r.BackingStore.Name,
		Connection:   conn.Name,
		TargetBucket: targetBucket,
		Backingstore: &nb.BackingStoreInfo{
			Name:      r.BackingStore.Name,
			Namespace: r.NooBaa.Namespace,
		},
	}

	return nil
}

// MakeExternalConnectionParams translates the backing store spec and secret,
// to noobaa api structures to be used for creating/updating external connetion and pool
func (r *Reconciler) MakeExternalConnectionParams() (*nb.AddExternalConnectionParams, error) {

	conn := &nb.AddExternalConnectionParams{
		Name: r.BackingStore.Name,
	}

	r.fixAlternateKeysNames()

	switch r.BackingStore.Spec.Type {

	case nbv1.StoreTypeAWSS3:
		if util.IsSTSClusterBS(r.BackingStore) {
			conn.EndpointType = nb.EndpointTypeAwsSTS
			conn.AWSSTSARN = *r.BackingStore.Spec.AWSS3.AWSSTSRoleARN
		} else {
			conn.EndpointType = nb.EndpointTypeAws
			conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
			conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]
		}
		awsS3 := r.BackingStore.Spec.AWSS3
		u := url.URL{
			Scheme: "https",
			Host:   "s3.amazonaws.com",
		}
		if awsS3.SSLDisabled {
			u.Scheme = "http"
		}
		if awsS3.Region != "" {
			u.Host = fmt.Sprintf("s3.%s.amazonaws.com", awsS3.Region)
			conn.Region = awsS3.Region
		}
		conn.Endpoint = u.String()

	case nbv1.StoreTypeS3Compatible:
		conn.EndpointType = nb.EndpointTypeS3Compat
		conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]
		s3Compatible := r.BackingStore.Spec.S3Compatible
		if s3Compatible.SignatureVersion == nbv1.S3SignatureVersionV4 {
			conn.AuthMethod = "AWS_V4"
		} else if s3Compatible.SignatureVersion == nbv1.S3SignatureVersionV2 {
			conn.AuthMethod = "AWS_V2"
		}
		if s3Compatible.Endpoint == "" {
			u := url.URL{
				Scheme: "https",
				Host:   "127.0.0.1:6443",
			}
			// if s3Compatible.SSLDisabled {
			// 	u.Scheme = "http"
			// 	u.Host = fmt.Sprintf("127.0.0.1:6001")
			// }
			conn.Endpoint = u.String()
		} else {
			match, err := regexp.MatchString(`^\w+://`, s3Compatible.Endpoint)
			if err != nil {
				return nil, util.NewPersistentError("InvalidEndpoint",
					fmt.Sprintf("Invalid endpoint url %q: %v", s3Compatible.Endpoint, err))
			}
			if !match {
				s3Compatible.Endpoint = "https://" + s3Compatible.Endpoint
				// if s3Options.SSLDisabled {
				// 	u.Scheme = "http"
				// }
			}
			u, err := url.Parse(s3Compatible.Endpoint)
			if err != nil {
				return nil, util.NewPersistentError("InvalidEndpoint",
					fmt.Sprintf("Invalid endpoint url %q: %v", s3Compatible.Endpoint, err))
			}
			if u.Scheme == "" {
				u.Scheme = "https"
				// if s3Options.SSLDisabled {
				// 	u.Scheme = "http"
				// }
			}
			conn.Endpoint = u.String()
		}

	case nbv1.StoreTypeIBMCos:
		conn.EndpointType = nb.EndpointTypeIBMCos
		conn.Identity = r.Secret.StringData["IBM_COS_ACCESS_KEY_ID"]
		conn.Secret = r.Secret.StringData["IBM_COS_SECRET_ACCESS_KEY"]
		IBMCos := r.BackingStore.Spec.IBMCos
		if IBMCos.SignatureVersion == nbv1.S3SignatureVersionV4 {
			conn.AuthMethod = "AWS_V4"
		} else if IBMCos.SignatureVersion == nbv1.S3SignatureVersionV2 {
			conn.AuthMethod = "AWS_V2"
		}
		if IBMCos.Endpoint == "" {
			u := url.URL{
				Scheme: "https",
				Host:   "127.0.0.1:6443",
			}
			// if IBMCos.SSLDisabled {
			// 	u.Scheme = "http"
			// 	u.Host = fmt.Sprintf("127.0.0.1:6001")
			// }
			conn.Endpoint = u.String()
		} else {
			match, err := regexp.MatchString(`^\w+://`, IBMCos.Endpoint)
			if err != nil {
				return nil, util.NewPersistentError("InvalidEndpoint",
					fmt.Sprintf("Invalid endpoint url %q: %v", IBMCos.Endpoint, err))
			}
			if !match {
				IBMCos.Endpoint = "https://" + IBMCos.Endpoint
				// if s3Options.SSLDisabled {
				// 	u.Scheme = "http"
				// }
			}
			u, err := url.Parse(IBMCos.Endpoint)
			if err != nil {
				return nil, util.NewPersistentError("InvalidEndpoint",
					fmt.Sprintf("Invalid endpoint url %q: %v", IBMCos.Endpoint, err))
			}
			if u.Scheme == "" {
				u.Scheme = "https"
				// if s3Options.SSLDisabled {
				// 	u.Scheme = "http"
				// }
			}
			conn.Endpoint = u.String()
		}

	case nbv1.StoreTypeAzureBlob:
		conn.EndpointType = nb.EndpointTypeAzure
		conn.Endpoint = "https://blob.core.windows.net"
		conn.Identity = r.Secret.StringData["AccountName"]
		conn.Secret = r.Secret.StringData["AccountKey"]

	case nbv1.StoreTypeGoogleCloudStorage:
		conn.EndpointType = nb.EndpointTypeGoogle
		conn.Endpoint = "https://www.googleapis.com"
		privateKeyJSON := r.Secret.StringData["GoogleServiceAccountPrivateKeyJson"]
		privateKey := &struct {
			ID string `json:"private_key_id"`
		}{}
		err := json.Unmarshal([]byte(privateKeyJSON), privateKey)
		if err != nil {
			return nil, util.NewPersistentError("InvalidGoogleSecret", fmt.Sprintf(
				"Invalid secret for google type %q expected JSON in data.GoogleServiceAccountPrivateKeyJson",
				r.Secret.Name,
			))
		}
		conn.Identity = privateKey.ID
		conn.Secret = privateKeyJSON

	case nbv1.StoreTypePVPool:
		return nil, util.NewPersistentError("InvalidType",
			fmt.Sprintf("%q type does not have external connection params", r.BackingStore.Spec.Type))

	default:
		return nil, util.NewPersistentError("InvalidType",
			fmt.Sprintf("Invalid backing store type %q", r.BackingStore.Spec.Type))
	}
	if !util.IsSTSClusterBS(r.BackingStore) {
		if !util.IsStringGraphicOrSpacesCharsOnly(conn.Identity) || !util.IsStringGraphicOrSpacesCharsOnly(conn.Secret) {
			return nil, util.NewPersistentError("InvalidSecret",
				fmt.Sprintf("Invalid secret containing non graphic characters (perhaps not base64 encoded?) %q", r.Secret.Name))
		}
	}

	return conn, nil
}

func (r *Reconciler) fixAlternateKeysNames() {

	alternateAccessKeyNames := []string{"aws_access_key_id", "AccessKey"}
	alternateSecretKeyNames := []string{"aws_secret_access_key", "SecretKey"}

	if r.Secret.StringData["AWS_ACCESS_KEY_ID"] == "" {
		for _, key := range alternateAccessKeyNames {
			if r.Secret.StringData[key] != "" {
				r.Secret.StringData["AWS_ACCESS_KEY_ID"] = r.Secret.StringData[key]
				break
			}
		}
	}

	if r.Secret.StringData["AWS_SECRET_ACCESS_KEY"] == "" {
		for _, key := range alternateSecretKeyNames {
			if r.Secret.StringData[key] != "" {
				r.Secret.StringData["AWS_SECRET_ACCESS_KEY"] = r.Secret.StringData[key]
				break
			}
		}
	}
}

// ReconcileExternalConnection handles the external connection using noobaa api
func (r *Reconciler) ReconcileExternalConnection() error {

	if r.ExternalConnectionInfo != nil {
		return nil
	}

	if r.AddExternalConnectionParams == nil {
		return nil
	}

	checkConnectionParams := &nb.CheckExternalConnectionParams{
		Name:         r.AddExternalConnectionParams.Name,
		EndpointType: r.AddExternalConnectionParams.EndpointType,
		Endpoint:     r.AddExternalConnectionParams.Endpoint,
		Identity:     r.AddExternalConnectionParams.Identity,
		Secret:       r.AddExternalConnectionParams.Secret,
		AuthMethod:   r.AddExternalConnectionParams.AuthMethod,
	}

	if r.UpdateExternalConnectionParams != nil {
		checkConnectionParams.IgnoreNameAlreadyExist = true
		err := r.CheckExternalConnection(checkConnectionParams)
		if err != nil {
			return err
		}

		err = r.NBClient.UpdateExternalConnectionAPI(*r.UpdateExternalConnectionParams)
		if err != nil {
			return err
		}
		r.UpdateExternalConnectionParams = nil
		return nil
	}

	checkConnectionParams.IgnoreNameAlreadyExist = false
	err := r.CheckExternalConnection(checkConnectionParams)
	if err != nil {
		return err
	}

	err = r.NBClient.AddExternalConnectionAPI(*r.AddExternalConnectionParams)
	if err != nil {
		return err
	}

	return nil
}

// CheckExternalConnection checks an external connection using the noobaa api
func (r *Reconciler) CheckExternalConnection(connInfo *nb.CheckExternalConnectionParams) error {
	res, err := r.NBClient.CheckExternalConnectionAPI(*connInfo)
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
		if time.Since(r.BackingStore.CreationTimestamp.Time) < 5*time.Minute {
			r.Logger.Infof("got invalid endpoint. requeuing for 5 minutes to make sure it is not a temporary connection issue")
			return fmt.Errorf("got invalid endpoint. requeue again")
		}
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

	return nil
}

// ReconcilePool handles the pool using noobaa api
func (r *Reconciler) ReconcilePool() error {

	// TODO we only support creation here, but not updates - just for pvpool
	if r.PoolInfo != nil {
		if r.BackingStore.Spec.Type == nbv1.StoreTypePVPool {
			if r.UpdateHostsPoolParams != nil {
				err := r.NBClient.UpdateHostsPoolAPI(*r.UpdateHostsPoolParams)
				if err != nil {
					return err
				}
			}
			return r.reconcilePvPool()
		}
		return nil
	}

	poolName := ""

	if r.CreateHostsPoolParams != nil {
		res, err := r.NBClient.CreateHostsPoolAPI(*r.CreateHostsPoolParams)
		if err != nil {
			if nbErr, ok := err.(*nb.RPCError); ok {
				if nbErr.RPCCode == "BAD_REQUEST" {
					msg := fmt.Sprintf("NooBaa BackingStore is in rejected phase due to %s", nbErr.Message)
					return util.NewPersistentError("SmallVolumeSize", msg)
				}
			}
			return err
		}
		if r.Secret.StringData["AGENT_CONFIG"] == "" {
			r.Secret.StringData["AGENT_CONFIG"] = res
			util.KubeUpdate(r.Secret)
		}
		err = r.NBClient.UpdateAllBucketsDefaultPool(nb.UpdateDefaultResourceParams{
			PoolName: r.CreateHostsPoolParams.Name,
		})
		if err != nil {
			return err
		}
		err = r.reconcilePvPool()
		if err != nil {
			return err
		}
		return nil
	}

	if r.CreateCloudPoolParams != nil {
		if r.BackingStore.ObjectMeta.Annotations != nil {
			if _, ok := r.BackingStore.ObjectMeta.Annotations["rgw"]; ok {
				cephCluster := &cephv1.CephCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "ocs-storagecluster",
						Namespace: options.Namespace,
					},
				}
				if util.KubeCheck(cephCluster) {
					availCapacity := nb.UInt64ToBigInt(cephCluster.Status.CephStatus.Capacity.AvailableBytes)
					r.CreateCloudPoolParams.AvailableCapacity = &availCapacity
				}
			}
		}

		err := r.NBClient.CreateCloudPoolAPI(*r.CreateCloudPoolParams)
		if err != nil {
			return err
		}
		poolName = r.CreateCloudPoolParams.Name
	}

	if poolName != "" {
		err := r.NBClient.UpdateAllBucketsDefaultPool(nb.UpdateDefaultResourceParams{
			PoolName: poolName,
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Reconciler) reconcilePvPool() error {
	if r.Secret.StringData == nil {
		return fmt.Errorf("reconcilePvPool: r.Secret.StringData is not initialized yet")
	}
	if r.Secret.StringData["AGENT_CONFIG"] == "" {
		res, err := r.NBClient.GetHostsPoolAgentConfigAPI(nb.GetHostsPoolAgentConfigParams{
			Name: r.BackingStore.Name,
		})
		if err != nil {
			return err
		}
		r.Secret.StringData["AGENT_CONFIG"] = res
		util.KubeUpdate(r.Secret)
	}
	podsList := &corev1.PodList{}
	pvcsList := &corev1.PersistentVolumeClaimList{}
	util.KubeList(podsList, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	util.KubeList(pvcsList, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	if len(pvcsList.Items) < r.BackingStore.Spec.PVPool.NumVolumes {
		err := r.reconcileMissingPvcs(pvcsList)
		if err != nil {
			return err
		}
		util.KubeList(pvcsList, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	}
	if len(podsList.Items) < len(pvcsList.Items) {
		err := r.reconcileMissingPods(podsList, pvcsList)
		if err != nil {
			return err
		}
	}
	return r.reconcileExistingPods(podsList)
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func (r *Reconciler) reconcileMissingPods(podsList *corev1.PodList, pvcsList *corev1.PersistentVolumeClaimList) error {
	claimNames := []string{}
	for _, pod := range podsList.Items {
		claimNames = append(claimNames, pod.Spec.Volumes[1].PersistentVolumeClaim.ClaimName)
	}
	if err := r.updatePodTemplate(); err != nil {
		return err
	}
	for _, pvc := range pvcsList.Items {
		if !contains(claimNames, pvc.Name) {
			i := strings.LastIndex(pvc.Name, "-")
			postfix := pvc.Name[i+1:]
			newPod := r.PodAgentTemplate.DeepCopy()
			newPod.Name = fmt.Sprintf("%s-%s-pod-%s", r.BackingStore.Name, options.SystemName, postfix)
			newPod.Namespace = options.Namespace
			newPod.Spec.Volumes[1].PersistentVolumeClaim.ClaimName = pvc.Name
			r.Own(newPod)
			util.KubeCreateSkipExisting(newPod)
		}
	}
	return nil
}

func (r *Reconciler) reconcileExistingPods(podsList *corev1.PodList) error {
	noneAttachingAgents := 0
	failedAttachingAgents := 0
	for _, pod := range podsList.Items {
		// check if pod need to be updated and deleted
		if r.needUpdate(&pod) {
			util.KubeDelete(&pod)
		} else if !r.isPodinNoobaa(&pod) {
			noneAttachingAgents++
			if time.Since(pod.CreationTimestamp.Time) > 10*time.Minute {
				failedAttachingAgents++
				r.Logger.Errorf("Pod %s didn't attach to noobaa system for more than 10 minutes", pod.Name)
			} else {
				r.Logger.Warnf("Pod %s didn't attach yet to noobaa system", pod.Name)
			}
		}
	}
	if len(podsList.Items) < r.BackingStore.Spec.PVPool.NumVolumes {
		return fmt.Errorf("BackingStore Still didn't start all the pods. %d from %d has started",
			len(podsList.Items), r.BackingStore.Spec.PVPool.NumVolumes)
	}
	attachedAgents := len(podsList.Items) - noneAttachingAgents
	if attachedAgents < r.BackingStore.Spec.PVPool.NumVolumes {
		if failedAttachingAgents == noneAttachingAgents {
			return util.NewPersistentError("Failed connecting all pods in backingstore for more than 10 minutes",
				fmt.Sprintf("Current failing: %d from requested: %d",
					failedAttachingAgents, r.BackingStore.Spec.PVPool.NumVolumes))
		}
		return fmt.Errorf("BackingStore Still didn't connect all requested pods. %d from %d are pending",
			noneAttachingAgents, r.BackingStore.Spec.PVPool.NumVolumes)
	}
	return nil
}

// return true if update is required
func compareResourceList(template, container *corev1.ResourceList) bool {
	for _, res := range []corev1.ResourceName{corev1.ResourceCPU, corev1.ResourceMemory} {
		if qty, ok := (*template)[res]; ok {
			if qty.Cmp((*container)[res]) != 0 {
				return true
			}
		}
	}
	return false
}

// return true if resources need to be updated
func (r *Reconciler) needUpdateResources(c *corev1.Container) bool {
	pvPool := r.BackingStore.Spec.PVPool
	if pvPool == nil || pvPool.VolumeResources == nil {
		return false
	}

	return compareResourceList(&pvPool.VolumeResources.Requests, &c.Resources.Requests) ||
		compareResourceList(&pvPool.VolumeResources.Limits, &c.Resources.Limits)
}

func (r *Reconciler) needUpdate(pod *corev1.Pod) bool {
	var c = &pod.Spec.Containers[0]
	for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY"} {
		envVar := util.GetEnvVariable(&c.Env, name)
		val, ok := os.LookupEnv(name)
		if (envVar == nil && ok) || (envVar != nil && (!ok || envVar.Value != val)) {
			r.Logger.Warnf("Change in Env variables detected: os(%s) container(%v)", val, envVar)
			return true
		}
	}
	if c.Image != r.NooBaa.Status.ActualImage {
		r.Logger.Warnf("Change in Image detected: current image(%v) noobaa image(%v)", c.Image, r.NooBaa.Status.ActualImage)
		return true
	}

	if r.needUpdateResources(c) {
		r.Logger.Warnf("Change in backing store agent resources detected")
		return true
	}

	podSecrets := pod.Spec.ImagePullSecrets
	noobaaSecret := r.NooBaa.Spec.ImagePullSecret
	if noobaaSecret == nil {
		sa := util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount)
		sa.Name = pod.Spec.ServiceAccountName
		sa.Namespace = options.Namespace
		if util.KubeCheck(sa) && !reflect.DeepEqual(sa.ImagePullSecrets, podSecrets) {
			r.Logger.Warnf("Change in Image Pull Secrets detected: SA(%v) Spec(%v)", sa.ImagePullSecrets, podSecrets)
			return true
		}
	} else if len(podSecrets) == 0 || !reflect.DeepEqual(noobaaSecret, podSecrets[0]) {
		r.Logger.Warnf("Change in Image Pull Secrets detected: NoobaaSecret(%v) Spec(%v)", noobaaSecret, podSecrets)
		return true
	}
	return false
}

func (r *Reconciler) reconcileMissingPvcs(pvcsList *corev1.PersistentVolumeClaimList) error {
	r.updatePvcTemplate()
	for i := len(pvcsList.Items); i < r.BackingStore.Spec.PVPool.NumVolumes; i++ {
		postfix := util.RandomHex(4)
		pvcName := fmt.Sprintf("%s-%s-pvc-%s", r.BackingStore.Name, options.SystemName, postfix)
		newPvc := r.PvcAgentTemplate.DeepCopy()
		newPvc.Name = pvcName
		newPvc.Namespace = options.Namespace
		r.Own(newPvc)
		util.KubeCreateSkipExisting(newPvc)
	}
	return nil
}

func (r *Reconciler) isPodinNoobaa(pod *corev1.Pod) bool {
	for _, host := range *r.HostsInfo {
		if strings.HasPrefix(host.Name, pod.Name) {
			return true
		}
	}
	return false
}

func (r *Reconciler) updatePodTemplate() error {
	c := &r.PodAgentTemplate.Spec.Containers[0]
	for j := range c.Env {
		switch c.Env[j].Name {
		case "AGENT_CONFIG":
			c.Env[j].ValueFrom = &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.Secret.Name,
					},
					Key: "AGENT_CONFIG",
				},
			}
		}
	}
	util.ReflectEnvVariable(&c.Env, "HTTP_PROXY")
	util.ReflectEnvVariable(&c.Env, "HTTPS_PROXY")
	util.ReflectEnvVariable(&c.Env, "NO_PROXY")

	c.Image = r.NooBaa.Status.ActualImage
	if r.NooBaa.Spec.ImagePullSecret == nil {
		r.PodAgentTemplate.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{}
	} else {
		r.PodAgentTemplate.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
	}
	r.PodAgentTemplate.Labels = map[string]string{
		"app":  "noobaa",
		"pool": r.BackingStore.Name,
	}
	if r.NooBaa.Spec.Tolerations != nil {
		r.PodAgentTemplate.Spec.Tolerations = r.NooBaa.Spec.Tolerations
	}
	if r.NooBaa.Spec.Affinity != nil {
		r.PodAgentTemplate.Spec.Affinity = r.NooBaa.Spec.Affinity
	}

	return r.updatePodResourcesTemplate(c)
}

func (r *Reconciler) updatePodResourcesTemplate(c *corev1.Container) error {
	minimalCPU := resource.MustParse(minCPUString)
	minimalMemory := resource.MustParse(minMemoryString)
	var src, dst *corev1.ResourceList
	pvPool := r.BackingStore.Spec.PVPool

	// Request
	dst = &c.Resources.Requests
	if pvPool != nil && pvPool.VolumeResources != nil {
		src = &pvPool.VolumeResources.Requests
	}
	if err := r.reconcileResources(src, dst, minimalCPU, minimalMemory); err != nil {
		return err
	}

	// Limits
	src = nil
	if pvPool != nil && pvPool.VolumeResources != nil {
		src = &pvPool.VolumeResources.Limits
	}
	dst = &c.Resources.Limits
	if err := r.reconcileResources(src, dst, minimalCPU, minimalMemory); err != nil {
		return err
	}

	return nil
}

func (r *Reconciler) updatePvcTemplate() {
	if r.BackingStore.Spec.PVPool.StorageClass != "" {
		r.PvcAgentTemplate.Spec.StorageClassName = &r.BackingStore.Spec.PVPool.StorageClass
	} else if r.NooBaa.Spec.PVPoolDefaultStorageClass != nil {
		r.PvcAgentTemplate.Spec.StorageClassName = r.NooBaa.Spec.PVPoolDefaultStorageClass
	}
	r.PvcAgentTemplate.Spec.Resources = *r.BackingStore.Spec.PVPool.VolumeResources
	r.PvcAgentTemplate.Labels = map[string]string{
		"app":  "noobaa",
		"pool": r.BackingStore.Name,
	}
}

func (r *Reconciler) deletePvPool() error {
	podsList := &corev1.PodList{}
	util.KubeList(podsList, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	util.KubeDeleteAllOf(&corev1.Pod{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	util.KubeDeleteAllOf(&corev1.PersistentVolumeClaim{}, client.InNamespace(options.Namespace), client.MatchingLabels{"pool": r.BackingStore.Name})
	return nil
}

func (r *Reconciler) upgradeBackingStore(sts *appsv1.StatefulSet) error {
	r.Logger.Infof("Deleting old statefulset: %s", sts.Name)
	envVar := util.GetEnvVariable(&sts.Spec.Template.Spec.Containers[0].Env, "AGENT_CONFIG")
	if envVar == nil {
		return util.NewPersistentError("NoAgentConfig", "Old BackingStore stateful set not having agent config")
	}
	agentConfig := envVar.Value
	replicas := sts.Spec.Replicas
	stsName := sts.Name
	util.KubeDelete(sts, client.PropagationPolicy("Orphan")) // delete STS leave pods behind
	o := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml)
	secret := o.(*corev1.Secret)
	secret.Name = fmt.Sprintf("backing-store-%s-%s", nbv1.StoreTypePVPool, r.BackingStore.Name)
	secret.Namespace = options.Namespace
	util.KubeCheck(secret)
	secret.StringData["AGENT_CONFIG"] = agentConfig // update secret for future pods
	util.KubeUpdate(secret)
	for i := 0; i < int(*replicas); i++ {
		pod := &corev1.Pod{}
		pod.Name = fmt.Sprintf("%s-%d", stsName, i)
		pod.Namespace = r.BackingStore.Namespace
		if util.KubeCheck(pod) {
			pod.Spec.Containers[0].Image = r.NooBaa.Status.ActualImage
			if r.NooBaa.Spec.ImagePullSecret == nil {
				pod.Spec.ImagePullSecrets =
					[]corev1.LocalObjectReference{}
			} else {
				pod.Spec.ImagePullSecrets =
					[]corev1.LocalObjectReference{*r.NooBaa.Spec.ImagePullSecret}
			}
			pod.Labels = map[string]string{"pool": r.BackingStore.Name}
			r.Own(pod)
			util.KubeUpdate(pod)
			pvc := &corev1.PersistentVolumeClaim{}
			pvc.Name = fmt.Sprintf("noobaastorage-%s", pod.Name)
			pvc.Namespace = pod.Namespace
			util.KubeCheck(pvc)
			pvc.Labels = map[string]string{"pool": r.BackingStore.Name}
			r.Own(pvc)
			util.KubeUpdate(pvc)
		}
	}
	return nil
}

func (r *Reconciler) reconcileResources(src, dst *corev1.ResourceList, minCPU, minMem resource.Quantity) error {
	cpu := minCPU
	mem := minMem

	if src != nil {
		if qty, ok := (*src)[corev1.ResourceCPU]; ok {
			if qty.Cmp(minCPU) < 0 {
				return util.NewPersistentError("MinRequestCpu",
					fmt.Sprintf("NooBaa BackingStore %v is in rejected phase due to small cpu request %v, min is %v", r.BackingStore.Name, qty.String(), minCPU.String()))
			}
			cpu = qty
		}
		if qty, ok := (*src)[corev1.ResourceMemory]; ok {
			if qty.Cmp(minMem) < 0 {
				return util.NewPersistentError("MinRequestCpu",
					fmt.Sprintf("NooBaa BackingStore %v is in rejected phase due to small memory request %v, min is %v", r.BackingStore.Name, qty.String(), minMem.String()))
			}
			mem = qty
		}
	}

	(*dst)[corev1.ResourceCPU] = cpu
	(*dst)[corev1.ResourceMemory] = mem
	return nil
}
