package bucketclass

import (
	"context"
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/validations"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
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

	NBClient   nb.Client
	SystemInfo *nb.SystemInfo

	BucketClass *nbv1.BucketClass
	NooBaa      *nbv1.NooBaa
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a bucket class
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {

	r := &Reconciler{
		Request:     req,
		Client:      client,
		Scheme:      scheme,
		Recorder:    recorder,
		Ctx:         context.TODO(),
		Logger:      logrus.WithField("bucketclass", req.Namespace+"/"+req.Name),
		BucketClass: util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_bucketclass_cr_yaml).(*nbv1.BucketClass),
		NooBaa:      util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
	}

	// Set Namespace
	r.BucketClass.Namespace = r.Request.Namespace
	r.NooBaa.Namespace = options.Namespace

	// Set Names
	r.BucketClass.Name = r.Request.Name
	r.NooBaa.Name = options.SystemName

	return r
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {
	var err error = nil
	res := reconcile.Result{}
	log := r.Logger
	log.Infof("Start BucketClass Reconcile...")

	systemFound := system.CheckSystem(r.NooBaa)

	if !util.KubeCheck(r.BucketClass) {
		log.Infof("❌ BucketClass %q not found or deleted.", r.BucketClass.Name)
		return res, err
	}

	if r.BucketClass.DeletionTimestamp != nil {
		err = r.ReconcileDeletion()
		return res, err
	}

	if !systemFound {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return res, nil
	}

	if util.EnsureCommonMetaFields(r.BucketClass, nbv1.Finalizer) {
		if !util.KubeUpdate(r.BucketClass) {
			log.Errorf("❌ BucketClass %q failed to add mandatory meta fields", r.BucketClass.Name)

			res.RequeueAfter = 3 * time.Second
			return res, nil
		}
	}

	err = r.ReconcilePhases()
	if err != nil {
		if perr, isPERR := err.(*util.PersistentError); isPERR {
			r.SetPhase(nbv1.BucketClassPhaseRejected, perr.Reason, perr.Message)
			log.Errorf("❌ Persistent Error: %s", err)
			if r.Recorder != nil {
				r.Recorder.Eventf(r.BucketClass, corev1.EventTypeWarning, perr.Reason, perr.Message)
			}
		} else {
			res.RequeueAfter = 3 * time.Second
			// leave current phase as is
			r.SetPhase("", "TemporaryError", err.Error())
			log.Warnf("⏳ Temporary Error: %s", err)
		}
	} else {
		if r.BucketClass.Status.Mode != "OPTIMAL" && r.BucketClass.Status.Mode != "" {
			if r.Recorder != nil {
				r.Recorder.Eventf(r.BucketClass, corev1.EventTypeWarning, r.BucketClass.Status.Mode, r.BucketClass.Status.Mode)
			}
		}
		r.SetPhase(
			nbv1.BucketClassPhaseReady,
			"BucketClassPhaseReady",
			"noobaa operator completed reconcile - bucket class is ready",
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
	if err := r.ReconcilePhaseConfiguring(); err != nil {
		return err
	}

	return nil
}

// SetPhase updates the status phase and conditions
func (r *Reconciler) SetPhase(phase nbv1.BucketClassPhase, reason string, message string) {

	c := &r.BucketClass.Status.Conditions

	if phase == "" {
		r.Logger.Infof("SetPhase: temporary error during phase %q", r.BucketClass.Status.Phase)
		util.SetProgressingCondition(c, reason, message)
		return
	}

	r.Logger.Infof("SetPhase: %s", phase)
	r.BucketClass.Status.Phase = phase
	switch phase {
	case nbv1.BucketClassPhaseReady:
		util.SetAvailableCondition(c, reason, message)
	case nbv1.BucketClassPhaseRejected:
		util.SetErrorCondition(c, reason, message)
	default:
		util.SetProgressingCondition(c, reason, message)
	}
}

// UpdateStatus updates the bucket class status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() error {
	err := r.Client.Status().Update(r.Ctx, r.BucketClass)
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
		nbv1.BucketClassPhaseVerifying,
		"BucketClassPhaseVerifying",
		"noobaa operator started phase 1/2 - \"Verifying\"",
	)

	err := validations.ValidateBucketClass(r.BucketClass)
	if err != nil {
		return util.NewPersistentError("ValidationError", err.Error())
	}

	if r.NooBaa.UID == "" {
		return util.NewPersistentError("MissingSystem",
			fmt.Sprintf("NooBaa system %q not found or deleted", r.NooBaa.Name))
	}
	if r.BucketClass.Spec.PlacementPolicy != nil {
		for i := range r.BucketClass.Spec.PlacementPolicy.Tiers {
			tier := &r.BucketClass.Spec.PlacementPolicy.Tiers[i]
			for _, backingStoreName := range tier.BackingStores {
				backStore := &nbv1.BackingStore{
					TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
					ObjectMeta: metav1.ObjectMeta{
						Name:      backingStoreName,
						Namespace: r.NooBaa.Namespace,
					},
				}
				if !util.KubeCheck(backStore) {
					return util.NewPersistentError("MissingBackingStore",
						fmt.Sprintf("NooBaa BackingStore %q not found or deleted", backingStoreName))
				}
				if backStore.Status.Phase == nbv1.BackingStorePhaseRejected {
					return util.NewPersistentError("RejectedBackingStore",
						fmt.Sprintf("NooBaa BackingStore %q is in rejected phase", backingStoreName))
				}
				if backStore.Status.Phase != nbv1.BackingStorePhaseReady {
					return fmt.Errorf("NooBaa BackingStore %q is not yet ready", backingStoreName)
				}
			}
		}
	}
	if r.BucketClass.Spec.NamespacePolicy != nil {
		namespaceStoresArr := validations.GetBucketclassNamespaceStoreArray(r.BucketClass)
		// check that namespace stores exists and their phase it ready
		for _, name := range namespaceStoresArr {
			nsStore := &nbv1.NamespaceStore{
				TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      name,
					Namespace: r.NooBaa.Namespace,
				},
			}
			if !util.KubeCheck(nsStore) {
				return util.NewPersistentError("MissingNamespaceStore",
					fmt.Sprintf("NooBaa NamespaceStore %q not found or deleted", name))
			}
			if nsStore.Status.Phase == nbv1.NamespaceStorePhaseRejected {
				return util.NewPersistentError("RejectedNamespaceStore",
					fmt.Sprintf("NooBaa NamespaceStore %q is in rejected phase", name))
			}
			if nsStore.Status.Phase != nbv1.NamespaceStorePhaseReady {
				return fmt.Errorf("NooBaa NamespaceStore %q is not yet ready", name)
			}
		}
	}

	return nil
}

// ReconcilePhaseConfiguring updates existing buckets to match the changes in bucket class
func (r *Reconciler) ReconcilePhaseConfiguring() error {

	r.SetPhase(
		nbv1.BucketClassPhaseConfiguring,
		"BucketClassPhaseConfiguring",
		"noobaa operator started phase 2/2 - \"Configuring\"",
	)

	objectBuckets := &nbv1.ObjectBucketList{}
	obcSelector, _ := labels.Parse("noobaa-domain=" + options.SubDomainNS())
	util.KubeList(objectBuckets, &client.ListOptions{LabelSelector: obcSelector})

	var bucketNames []string
	for i := range objectBuckets.Items {
		ob := &objectBuckets.Items[i]
		bucketClass := ob.Spec.AdditionalState["bucketclass"]
		bucketClassGeneration := ob.Spec.AdditionalState["bucketclassgeneration"]
		bucketName := ob.Spec.Endpoint.BucketName
		if bucketClass != r.BucketClass.Name {
			continue
		}
		if bucketClassGeneration == fmt.Sprintf("%d", r.BucketClass.Generation) {
			continue
		}
		bucketNames = append(bucketNames, bucketName)
	}

	if len(bucketNames) == 0 {
		return nil
	}

	sysClient, err := system.Connect(false)
	if err != nil {
		return err
	}
	r.NBClient = sysClient.NBClient

	if err := r.UpdateBucketClass(bucketNames); err != nil {
		return err
	}

	return nil
}

// ReconcileDeletion handles the deletion of a bucket class using the noobaa api
func (r *Reconciler) ReconcileDeletion() error {

	// Set the phase to let users know the operator has noticed the deletion request
	if r.BucketClass.Status.Phase != nbv1.BucketClassPhaseDeleting {
		r.SetPhase(
			nbv1.BucketClassPhaseDeleting,
			"BucketClassPhaseDeleting",
			"noobaa operator started deletion",
		)
		err := r.UpdateStatus()
		if err != nil {
			return err
		}
	}

	r.Logger.Infof("BucketClass %q remove finalizer", r.BucketClass.Name)
	return r.FinalizeDeletion()
}

// FinalizeDeletion removed the finalizer and updates in order to let the bucket class get reclaimed by kubernetes
func (r *Reconciler) FinalizeDeletion() error {
	util.RemoveFinalizer(r.BucketClass, nbv1.Finalizer)
	if !util.KubeUpdate(r.BucketClass) {
		return fmt.Errorf("BucketClass %q failed to remove finalizer %q", r.BucketClass.Name, nbv1.Finalizer)
	}
	return nil
}

// updateNamespaceBucketClass updates all namespace buckets that are assigned to a BucketClass
func (r *Reconciler) updateNamespaceBucketClass(bucketNames []string) error {
	log := r.Logger

	if r.BucketClass == nil {
		return fmt.Errorf("BucketClass not loaded %#v", r)
	}

	if r.BucketClass.Spec.NamespacePolicy != nil {
		namespacePolicyType := r.BucketClass.Spec.NamespacePolicy.Type
		var readResources []nb.NamespaceResourceFullConfig
		createBucketParams := &nb.CreateBucketParams{}
		createBucketParams.Namespace = &nb.NamespaceBucketInfo{}

		if namespacePolicyType == nbv1.NSBucketClassTypeSingle {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Single.Resource}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Single.Resource})

		} else if namespacePolicyType == nbv1.NSBucketClassTypeMulti {
			if r.BucketClass.Spec.NamespacePolicy.Multi.WriteResource != "" {
				createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
					Resource: r.BucketClass.Spec.NamespacePolicy.Multi.WriteResource}
			}

			for i := range r.BucketClass.Spec.NamespacePolicy.Multi.ReadResources {
				rr := r.BucketClass.Spec.NamespacePolicy.Multi.ReadResources[i]
				readResources = append(readResources, nb.NamespaceResourceFullConfig{Resource: rr})
			}

			createBucketParams.Namespace.ReadResources = readResources

		} else if namespacePolicyType == nbv1.NSBucketClassTypeCache {
			createBucketParams.Namespace.WriteResource = nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Cache.HubResource}
			createBucketParams.Namespace.ReadResources = append(readResources, nb.NamespaceResourceFullConfig{
				Resource: r.BucketClass.Spec.NamespacePolicy.Cache.HubResource})
			createBucketParams.Namespace.Caching = &nb.CacheSpec{TTLMs: r.BucketClass.Spec.NamespacePolicy.Cache.Caching.TTL}
			//cachePrefix := r.BucketClass.Spec.NamespacePolicy.Cache.Prefix
		}

		for i := range bucketNames {
			createBucketParams.Name = bucketNames[i]
			err := r.NBClient.UpdateBucketAPI(*createBucketParams)

			if err != nil {
				return fmt.Errorf("Failed to update obcs %q with error: %v", bucketNames[i], err)
			}
		}
		log.Infof("✅ Successfully updated namespace bucket class and obcs %q", r.BucketClass.Name)
	}
	return nil
}

// UpdateBucketClass updates all buckets that are assigned to a BucketClass
func (r *Reconciler) UpdateBucketClass(bucketNames []string) error {
	log := r.Logger

	if r.BucketClass == nil {
		return fmt.Errorf("BucketClass not loaded %#v", r)
	}

	if r.BucketClass.Spec.PlacementPolicy == nil &&
		r.BucketClass.Spec.NamespacePolicy != nil {
		return r.updateNamespaceBucketClass(bucketNames)
	}

	policyTiers := []nb.TierItem{}
	tiers := []nb.TierInfo{}

	for i := range r.BucketClass.Spec.PlacementPolicy.Tiers {
		tier := &r.BucketClass.Spec.PlacementPolicy.Tiers[i]
		// Tier is irrelevant and will be populated in the BE
		policyTiers = append(policyTiers, nb.TierItem{Order: int64(i), Tier: "TEMP"})
		// we assume either mirror or spread but no mix and the bucket class controller rejects mixed classes.
		placement := "SPREAD"
		if tier.Placement == nbv1.TierPlacementMirror {
			placement = "MIRROR"
		}
		// Name is irrelevant and will be populated in the BE
		tiers = append(tiers, nb.TierInfo{Name: "TEMP", AttachedPools: tier.BackingStores, DataPlacement: placement})
	}

	result, err := r.NBClient.UpdateBucketClass(nb.UpdateBucketClassParams{
		Name: r.BucketClass.Name,
		// Name is irrelevant and will be populated in the BE
		Policy: nb.TieringPolicyInfo{Name: "TEMP", Tiers: policyTiers},
		Tiers:  tiers,
	})

	if err != nil {
		return fmt.Errorf("Failed to update bucket class %q with error: %v - Can't revert changes", r.BucketClass.Name, err)
	}

	if result.ShouldRevert {
		r.BucketClass.Spec.PlacementPolicy.Tiers = []nbv1.Tier{}
		for _, t := range result.RevertToPolicy.Tiers {
			placement := nbv1.TierPlacementSpread
			if t.DataPlacement == "MIRROR" {
				placement = nbv1.TierPlacementMirror
			}
			r.BucketClass.Spec.PlacementPolicy.Tiers = append(r.BucketClass.Spec.PlacementPolicy.Tiers,
				nbv1.Tier{Placement: placement, BackingStores: t.AttachedPools})
		}
		util.KubeUpdate(r.BucketClass)
		return util.NewPersistentError("InvalidConfReverting", fmt.Sprintf("Unable to change bucketclass due to error: %v", result.ErrorMessage))
		// return fmt.Errorf("Failed to update bucket class %q with error: %v - Reverting back", r.BucketClass.Name, result.ErrorMessage)
	}

	log.Infof("✅ Successfully updated bucket class %q", r.BucketClass.Name)
	return nil
}
