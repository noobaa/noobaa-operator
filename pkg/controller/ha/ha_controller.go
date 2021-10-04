package hac

import (
	"context"

	"github.com/noobaa/noobaa-operator/v5/pkg/hac"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
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

// nodeIsTainted checks if a kubernetes node is tainted
// Noe: for the context of this function, "PreferNoSchedule" is
// not considered a taint, since it allows pods to continue
// function normally.
// On the other hand, "NoSchedule" and "NoExecute" would
// result in true return value
func nodeIsTainted(node *corev1.Node) bool {
	for _, t := range node.Spec.Taints {
		if t.Effect != corev1.TaintEffectPreferNoSchedule {
			return true
		}
	}

	return false
}

// deletePodsOnStartup - during start up delete NooBaa pods
// that might be stuck on a failing node
func deletePodsOnStartup(client client.Client) error {
	// fetch the cluster nodes from the api server
	nodeList := &corev1.NodeList{}
	if !util.KubeList(nodeList) {
		return errors.Errorf("failed to list nodes")
	}

	for _, node := range nodeList.Items {
		if !nodeIsReady(&node) || nodeIsTainted(&node) {
			pd := hac.PodDeleter{Client: client, NodeName: node.Name}
			if err := pd.DeletePodsOnNode(); err != nil {
				return errors.Errorf("failed to delete noobaa pods on the node %v in namespace %v", node.Name, options.Namespace)	
			}
		}
	}
	return nil
}

// nodeNeedsEvacuationPredicate selects nodes that were ready, but became unreachable
// or tainted
func nodeNeedsEvacuationPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(event.CreateEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldNode := e.ObjectOld.(*corev1.Node)
			newNode := e.ObjectNew.(*corev1.Node)
			if nodeIsReady(oldNode) && !nodeIsReady(newNode) {
				return true
			}

			if !nodeIsTainted(oldNode) && nodeIsTainted(newNode) {
				return true
			}

			return false
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
			func(ctx context.Context,req reconcile.Request) (reconcile.Result, error) {
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
	if err := c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{}, nodeNeedsEvacuationPredicate()); err != nil {
		return err
	}

	// delete pods that might be stuck on a failing node when operator first starts
	// handles cases like failure of a node running the operator pod
	return deletePodsOnStartup(mgr.GetClient())
}
