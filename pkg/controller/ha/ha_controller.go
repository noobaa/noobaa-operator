package hac

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/hac"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// nodeIsReady checks if a kubernetes node is ready
func nodeIsReady(node *corev1.Node) bool {
	for _, c := range node.Status.Conditions {
		if c.Type == corev1.NodeReady && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

// nodeNotReadyPredicate selects nodes that were ready, but became unreachable
func nodeNotReadyPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode := e.ObjectOld.(*corev1.Node)
			newNode := e.ObjectNew.(*corev1.Node)
			return nodeIsReady(oldNode) && !nodeIsReady(newNode)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(event.GenericEvent) bool {
			return false
		},
	}
}

// Add creates a nodewatcher Controller and adds it to the Manager.
func Add(mgr manager.Manager) error {

	opts := controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(req reconcile.Request) (reconcile.Result, error) {
				return hac.NewHAC(
					req.NamespacedName,
					mgr.GetClient(),
				).Reconcile()
			}),
	}

	c, err := controller.New(hac.Name, mgr, opts)
	if err != nil {
		return err
	}

	// start watching node state transitions
	return c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}, nodeNotReadyPredicate())
}
