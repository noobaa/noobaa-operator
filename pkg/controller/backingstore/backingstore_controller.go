package backingstore

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/backingstore"
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

	// Create a controller that runs reconcile on noobaa backing store

	c, err := controller.New("noobaa-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(context context.Context, req reconcile.Request) (reconcile.Result, error) {
				return backingstore.NewReconciler(
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

	// Predicate that filter events by their owner
	filterForOwnerPredicate := util.FilterForOwner{
		OwnerType: &nbv1.BackingStore{},
		Scheme:    mgr.GetScheme(),
	}

	// Predicate that filter events that noobaa is not their owner
	filterForNoobaaOwnerPredicate := util.FilterForOwner{
		OwnerType: &nbv1.NooBaa{},
		Scheme:    mgr.GetScheme(),
	}

	// Predicate that allows events that only change spec, labels or finalizers and will log any allowed events
	// This will stop infinite reconciles that triggered by status or irrelevant metadata changes
	backingStorePredicate := util.ComposePredicates(
		predicate.GenerationChangedPredicate{},
		util.LabelsChangedPredicate{},
		util.FinalizersChangedPredicate{},
		backingStoreModeChangedPredicate{},
	)

	// Watch for changes on resources to trigger reconcile
	ownerHandler := handler.EnqueueRequestForOwner(
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		&nbv1.BackingStore{},
		handler.OnlyControllerOwner(),
	)

	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.BackingStore{}, &handler.EnqueueRequestForObject{}, backingStorePredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.Pod{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.PersistentVolumeClaim{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.ConfigMap{}, ownerHandler, &filterForNoobaaOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}

	// setting another handler to watch events on secrets that not necessarily owned by the Backingstore.
	// only one OwnerReference can be a controller see:
	// https://github.com/kubernetes-sigs/controller-runtime/blob/master/pkg/controller/controllerutil/controllerutil.go#L54
	secretsHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		return backingstore.MapSecretToBackingStores(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		})
	})
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.Secret{}, secretsHandler, logEventsPredicate))
	if err != nil {
		return err
	}

	// Setting another handler to watch events on noobaa system that are not necessarily owned by the backingstore.
	// For example: modify toleration in Noobaa CR should pass to the backingstores.
	noobaaHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		return backingstore.MapNoobaaToBackingStores(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		})
	})
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.NooBaa{}, noobaaHandler, logEventsPredicate))
	if err != nil {
		return err
	}

	return nil
}

// backingStoreModeChangedPredicate will only allow events that changed Status.Mode.ModeCode.
// This predicate should be used only for BackingsStore objects!
type backingStoreModeChangedPredicate struct {
	predicate.Funcs
}

// Update implements the update event trap for LabelsChangedPredicate
func (p backingStoreModeChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	oldBackingStore, oldCastOk := e.ObjectOld.(*nbv1.BackingStore)
	newBackingStore, newCastOk := e.ObjectNew.(*nbv1.BackingStore)
	if !oldCastOk || !newCastOk {
		return false
	}
	return oldBackingStore.Status.Mode.ModeCode != newBackingStore.Status.Mode.ModeCode
}
