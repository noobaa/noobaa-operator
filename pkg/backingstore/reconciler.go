package backingstore

import (
	"context"
	"time"

	"github.com/noobaa/noobaa-operator/build/_output/bundle"
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/pkg/nb"
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

	NooBaa       *nbv1.NooBaa
	BackingStore *nbv1.BackingStore
	Secret       *corev1.Secret
}

// NewReconciler initializes a reconciler to be used for loading or reconciling a noobaa system
func NewReconciler(
	req types.NamespacedName,
	client client.Client,
	scheme *runtime.Scheme,
	recorder record.EventRecorder,
) *Reconciler {
	s := &Reconciler{
		Request:  req,
		Client:   client,
		Scheme:   scheme,
		Recorder: recorder,
		Ctx:      context.TODO(),
		Logger:   logrus.WithFields(logrus.Fields{"ns": req.Namespace, "sys": req.Name}),
		NooBaa:   util.KubeObject(bundle.File_deploy_crds_noobaa_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa),
	}
	util.SecretResetStringDataFromData(s.Secret)

	// Set Namespace
	s.NooBaa.Namespace = s.Request.Namespace
	s.BackingStore.Namespace = s.Request.Namespace
	s.Secret.Namespace = s.Request.Namespace

	// Set Names
	s.NooBaa.Name = s.Request.Name
	s.BackingStore.Name = s.Request.Name + "-core"
	s.Secret.Name = s.Request.Name + "-mgmt"

	return s
}

// Reconcile reads that state of the cluster for a System object,
// and makes changes based on the state read and what is in the System.Spec.
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *Reconciler) Reconcile() (reconcile.Result, error) {

	log := r.Logger.WithField("func", "Reconcile")
	log.Infof("Start ...")

	util.KubeCheck(r.NooBaa)
	if r.NooBaa.UID == "" {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return reconcile.Result{}, nil
	}

	err := util.CombineErrors(
		r.RunReconcile(),
		r.UpdateStatus(),
	)
	if util.IsPersistentError(err) {
		log.Errorf("❌ Persistent Error: %s", err)
		return reconcile.Result{}, nil
	}
	if err != nil {
		log.Warnf("⏳ Temporary Error: %s", err)
		return reconcile.Result{RequeueAfter: 2 * time.Second}, nil
	}
	log.Infof("✅ Done")
	return reconcile.Result{}, nil
}

// RunReconcile runs the reconcile flow and populates System.Status.
func (r *Reconciler) RunReconcile() error {
	return nil
}

// UpdateStatus updates the backing store status in kubernetes from the memory
func (r *Reconciler) UpdateStatus() error {
	log := r.Logger.WithField("func", "UpdateStatus")
	log.Infof("Updating backing store status")
	return r.Client.Status().Update(r.Ctx, r.BackingStore)
}
