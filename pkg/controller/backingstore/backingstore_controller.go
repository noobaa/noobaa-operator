package backingstore

import (
	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/backingstore"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
			func(req reconcile.Request) (reconcile.Result, error) {
				return backingstore.NewReconciler(
					req.NamespacedName,
					mgr.GetClient(),
					mgr.GetScheme(),
					mgr.GetEventRecorderFor("noobaa-operator"),
				).Reconcile()
			}),
	})
	if err != nil {
		return err
	}

	// Watch for changes on resources to trigger reconcile

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &nbv1.BackingStore{}}

	err = c.Watch(&source.Kind{Type: &nbv1.BackingStore{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.PersistentVolumeClaim{}}, ownerHandler)
	if err != nil {
		return err
	}

	// err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.Funcs{
	// 	CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
	// 		fmt.Println("JAJA: Create", e)
	// 		ownerHandler.Create(e, q)
	// 	},
	// 	UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
	// 		fmt.Println("JAJA: Update", e)
	// 		ownerHandler.Update(e, q)
	// 	},
	// 	DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
	// 		fmt.Println("JAJA: Delete", e)
	// 		ownerHandler.Delete(e, q)
	// 	},
	// 	GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
	// 		fmt.Println("JAJA: Generic", e)
	// 		ownerHandler.Generic(e, q)
	// 	},
	// })
	if err != nil {
		return err
	}

	return nil
}
