package bucketclass

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

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

	// Create a controller that runs reconcile on noobaa bucket class

	c, err := controller.New("noobaa-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(context context.Context, req reconcile.Request) (reconcile.Result, error) {
				return bucketclass.NewReconciler(
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
	bucketClassPredicate := util.ComposePredicates(
		predicate.GenerationChangedPredicate{},
		util.LabelsChangedPredicate{},
		util.FinalizersChangedPredicate{},
	)

	// Watch for changes on resources to trigger reconcile
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.BucketClass{}, &handler.EnqueueRequestForObject{},
		ignoreUnmatchedProvisioner(options.Namespace), bucketClassPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}

	backingStoreHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		return bucketclass.MapBackingstoreToBucketclasses(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		})
	},
	)
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.BackingStore{}, backingStoreHandler,
		util.IgnoreIfNotInNamespace(options.Namespace), logEventsPredicate))
	if err != nil {
		return err
	}

	namespaceStoreHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
		return bucketclass.MapNamespacestoreToBucketclasses(types.NamespacedName{
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
		})
	},
	)
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.NamespaceStore{}, namespaceStoreHandler,
		util.IgnoreIfNotInNamespace(options.Namespace), logEventsPredicate))
	if err != nil {
		return err
	}

	return nil
}

func ignoreUnmatchedProvisioner(noobaaOperatorNamespace string) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			return isObjectForProvisioner(e.Object, noobaaOperatorNamespace)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return isObjectForProvisioner(e.Object, noobaaOperatorNamespace)
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return isObjectForProvisioner(e.ObjectNew, noobaaOperatorNamespace)
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return isObjectForProvisioner(e.Object, noobaaOperatorNamespace)
		},
	}
}

func isObjectForProvisioner(obj client.Object, noobaaOperatorNamespace string) bool {
	noobaaOperatorLabel := "noobaa-operator"
	provisionerLable, ok := obj.GetLabels()[noobaaOperatorLabel]
	if !ok {
		// If the object doesn't have the provisioner label, it is only for the provisioner
		// if it is in the provisioner namespace
		return obj.GetNamespace() == noobaaOperatorNamespace
	}

	// If the object has the provisioner label, it is only for the provisioner
	// if the label value is the provisioner namespace
	return provisionerLable == noobaaOperatorNamespace
}
