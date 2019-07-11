package bucketclass

import (
	nbv1 "github.com/noobaa/noobaa-operator/pkg/apis/noobaa/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
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
			func(req reconcile.Request) (reconcile.Result, error) {
				return reconcile.Result{}, nil
			}),
	})
	if err != nil {
		return err
	}

	// Watch for changes on resources to trigger reconcile

	primaryHandler := &handler.EnqueueRequestForObject{}

	err = c.Watch(&source.Kind{Type: &nbv1.BucketClass{}}, primaryHandler)
	if err != nil {
		return err
	}

	return nil
}
