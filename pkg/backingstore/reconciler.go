package backingstore

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var bsModeToPhaseMap map[string]nbv1.BackingStorePhaseInfo

func init() {
	bsModeToPhaseMap = modeToPhaseMap()
}

func modeToPhaseMap() map[string]nbv1.BackingStorePhaseInfo {
	return map[string]nbv1.BackingStorePhaseInfo{
		"INITIALIZING":        nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: INITALIZING"},
		"DELETING":            nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: DELETING"},
		"SCALING":             nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: SCALING"},
		"MOST_NODES_ISSUES":   nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: MOST_NODES_ISSUES"},
		"MANY_NODES_ISSUES":   nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: MANY_NODES_ISSUES"},
		"MOST_STORAGE_ISSUES": nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: MOST_STORAGE_ISSUES"},
		"MANY_STORAGE_ISSUES": nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: MANY_STORAGE_ISSUES"},
		"MANY_NODES_OFFLINE":  nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: MANY_NODES_OFFLINE"},
		"LOW_CAPACITY":        nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: LOW_CAPACITY"},
		"OPTIMAL":             nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseReady, "BackingStorePhaseReady", "Backing store mode: OPTIMAL"},
		"HAS_NO_NODES":        nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: HAS_NO_NODES"},
		"ALL_NODES_OFFLINE":   nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: ALL_NODES_OFFLINE"},
		"NO_CAPACITY":         nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: NO_CAPACITY"},
		"IO_ERRORS":           nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: IO_ERRORS"},
		"STORAGE_NOT_EXIST":   nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: STORAGE_NOT_EXIST"},
		"AUTH_FAILED":         nbv1.BackingStorePhaseInfo{nbv1.BackingStorePhaseRejected, "BackingStorePhaseRejected", "Backing store mode: AUTH_FAILED"},
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
		BackingStore: util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_backingstore_cr_yaml).(*nbv1.BackingStore),
		NooBaa:       util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
		Secret:       util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret),
	}

	// Set Namespace
	r.BackingStore.Namespace = r.Request.Namespace
	r.NooBaa.Namespace = r.Request.Namespace

	// Set Names
	r.BackingStore.Name = r.Request.Name
	r.NooBaa.Name = options.SystemName

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

	res := reconcile.Result{}
	log := r.Logger
	log.Infof("Start ...")

	util.KubeCheck(r.BackingStore)

	if r.BackingStore.UID == "" {
		log.Infof("BackingStore %q not found or deleted. Skip reconcile.", r.BackingStore.Name)
		return reconcile.Result{}, nil
	}

	system.CheckSystem(r.NooBaa)

	secretRef := GetBackingStoreSecret(r.BackingStore)
	if secretRef != nil {
		r.Secret.Name = secretRef.Name
		r.Secret.Namespace = secretRef.Namespace
		if r.Secret.Namespace == "" {
			r.Secret.Namespace = r.BackingStore.Namespace
		}
		util.KubeCheck(r.Secret)
	}

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
		bsPhase, exist := bsModeToPhaseMap[r.BackingStore.Status.Mode.ModeCode]

		if exist && bsPhase.Phase != r.BackingStore.Status.Phase {
			r.SetPhase(bsPhase.Phase, bsPhase.Message, bsPhase.Reason)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.BackingStore, corev1.EventTypeWarning, bsPhase.Reason, bsPhase.Message)
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

	r.UpdateStatus()
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
			if rpcErr, isRPCErr := err.(*nb.RPCError); isRPCErr {
				if rpcErr.RPCCode == "DEFAULT_RESOURCE" {
					return util.NewPersistentError("DefaultResource",
						fmt.Sprintf("DeletePoolAPI cannot complete because pool %q is an account default resource", r.PoolInfo.Name))
				}
				if rpcErr.RPCCode == "IN_USE" {
					return util.NewPersistentError("ResourceInUse",
						fmt.Sprintf("DeletePoolAPI cannot complete because pool %q has buckets attached", r.PoolInfo.Name))
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
		const defaultVolumeSize = int64(21474836480) // 20Gi=20*1024^3
		var volumeSize int64
		pvPool := r.BackingStore.Spec.PVPool
		if pvPool.VolumeResources != nil {
			qty := pvPool.VolumeResources.Requests[corev1.ResourceName(corev1.ResourceStorage)]
			volumeSize = qty.Value()
		}
		if volumeSize == 0 {
			volumeSize = int64(defaultVolumeSize)
		}
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
		TargetBucket: GetBackingStoreTargetBucket(r.BackingStore),
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
		conn.EndpointType = nb.EndpointTypeAws
		conn.Identity = r.Secret.StringData["AWS_ACCESS_KEY_ID"]
		conn.Secret = r.Secret.StringData["AWS_SECRET_ACCESS_KEY"]
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
		} else if s3Compatible.SignatureVersion != "" {
			return nil, util.NewPersistentError("InvalidSignatureVersion",
				fmt.Sprintf("Invalid s3 signature version %q for backing store %q",
					s3Compatible.SignatureVersion, r.BackingStore.Name))
		}
		if s3Compatible.Endpoint == "" {
			u := url.URL{
				Scheme: "https",
				Host:   fmt.Sprintf("127.0.0.1:6443"),
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
		} else if IBMCos.SignatureVersion != "" {
			return nil, util.NewPersistentError("InvalidSignatureVersion",
				fmt.Sprintf("Invalid s3 signature version %q for backing store %q",
					IBMCos.SignatureVersion, r.BackingStore.Name))
		}
		if IBMCos.Endpoint == "" {
			u := url.URL{
				Scheme: "https",
				Host:   fmt.Sprintf("127.0.0.1:6443"),
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

	if !util.IsStringGraphicOrSpacesCharsOnly(conn.Identity) || !util.IsStringGraphicOrSpacesCharsOnly(conn.Secret) {
		return nil, util.NewPersistentError("InvalidSecret",
			fmt.Sprintf("Invalid secret containing non graphic characters (perhaps not base64 encoded?) %q", r.Secret.Name))
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

	// TODO we only support creation here, but not updates
	if r.ExternalConnectionInfo != nil {
		return nil
	}
	if r.AddExternalConnectionParams == nil {
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

	// TODO we only support creation here, but not updates
	if r.PoolInfo != nil {
		return nil
	}

	poolName := ""

	if r.CreateHostsPoolParams != nil {
		err := r.NBClient.CreateHostsPoolAPI(*r.CreateHostsPoolParams)
		if err != nil {
			return err
		}
		poolName = r.CreateHostsPoolParams.Name
	}

	if r.CreateCloudPoolParams != nil {
		err := r.NBClient.CreateCloudPoolAPI(*r.CreateCloudPoolParams)
		if err != nil {
			return err
		}
		poolName = r.CreateCloudPoolParams.Name
	}

	if poolName != "" {
		err := r.NBClient.UpdateAllBucketsDefaultPool(nb.UpdateDefaultPoolParams{
			PoolName: poolName,
		})
		if err != nil {
			return err
		}
	}

	return nil
}
