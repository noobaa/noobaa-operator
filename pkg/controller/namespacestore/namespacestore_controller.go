package namespacestore

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	// Create a controller that runs reconcile on noobaa namespace store

	c, err := controller.New("noobaa-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(context context.Context, req reconcile.Request) (reconcile.Result, error) {
				return namespacestore.NewReconciler(
					req.NamespacedName,
					mgr.GetClient(),
					mgr.GetScheme(),
					mgr.GetEventRecorderFor("noobaa-operator"),
				).Reconcile()
			}),
		SkipNameValidation: &[]bool{true}[0],
	})
	if err != nil {
		return err
	}

	// Predicate that allow us to log event that are being queued
	logEventsPredicate := util.LogEventsPredicate{}

	// Predicate that allows events that only change spec, labels or finalizers and will log any allowed events
	// This will stop infinite reconciles that triggered by status or irrelevant metadata changes
	namespaceStorePredicate := util.ComposePredicates(
		predicate.GenerationChangedPredicate{},
		util.LabelsChangedPredicate{},
		util.FinalizersChangedPredicate{},
		namespaceStoreModeChangedPredicate{},
	)
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.NamespaceStore{}, &handler.EnqueueRequestForObject{},
		namespaceStorePredicate, &logEventsPredicate))
	if err != nil {
		return err
	}

	secretsHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		return namespacestore.MapSecretToNamespaceStores(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		})
	})
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.Secret{}, secretsHandler, logEventsPredicate))
	if err != nil {
		return err
	}

	return nil
}

// namespaceStoreModeChangedPredicate will only allow events that changed Status.Mode.ModeCode.
// This predicate should be used only for NamespaceStore objects!
type namespaceStoreModeChangedPredicate struct {
	predicate.Funcs
}

// Update implements the update event trap for LabelsChangedPredicate
func (p namespaceStoreModeChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	oldNamespaceStore, oldCastOk := e.ObjectOld.(*nbv1.NamespaceStore)
	newNamespaceStore, newCastOk := e.ObjectNew.(*nbv1.NamespaceStore)
	if !oldCastOk || !newCastOk {
		return false
	}
	return oldNamespaceStore.Status.Mode.ModeCode != newNamespaceStore.Status.Mode.ModeCode
}
