package noobaa

import (
	nbv1 "github.com/noobaa/noobaa-operator/v2/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v2/pkg/nb"
	"github.com/noobaa/noobaa-operator/v2/pkg/options"
	"github.com/noobaa/noobaa-operator/v2/pkg/system"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NotificationSource specifies a queue of notifications
type NotificationSource struct {
	Queue workqueue.RateLimitingInterface
}

// Start will setup s.Queue field
func (s *NotificationSource) Start(handler handler.EventHandler, q workqueue.RateLimitingInterface, predicates ...predicate.Predicate) error {
	s.Queue = q
	return nil
}

// Add creates a Controller and adds it to the Manager.
// The Manager will set fields on the Controller and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {

	// Create a controller that runs reconcile on noobaa system

	c, err := controller.New("noobaa-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler: reconcile.Func(
			func(req reconcile.Request) (reconcile.Result, error) {
				return system.NewReconciler(
					req.NamespacedName,
					mgr.GetClient(),
					mgr.GetScheme(),
					mgr.GetRecorder("noobaa-operator"),
				).Reconcile()
			}),
	})
	if err != nil {
		return err
	}

	// Watch for changes on resources to trigger reconcile

	ownerHandler := &handler.EnqueueRequestForOwner{IsController: true, OwnerType: &nbv1.NooBaa{}}

	err = c.Watch(&source.Kind{Type: &nbv1.NooBaa{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, ownerHandler)
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, ownerHandler)
	if err != nil {
		return err
	}

	// Watch for StorageClass changes to trigger reconcile and recreate it when deleted
	err = c.Watch(&source.Kind{Type: &storagev1.StorageClass{}}, &handler.EnqueueRequestsFromMapFunc{
		ToRequests: handler.ToRequestsFunc(func(mo handler.MapObject) []reconcile.Request {
			sc, ok := mo.Object.(*storagev1.StorageClass)
			if !ok || sc.Provisioner != options.ObjectBucketProvisionerName() {
				return nil
			}
			return []reconcile.Request{{
				NamespacedName: types.NamespacedName{
					Name:      options.SystemName,
					Namespace: options.Namespace,
				},
			}}
		}),
	})
	if err != nil {
		return err
	}
	// watch on notificationSource in order to keep the controller work queue
	notificationSource := &NotificationSource{}
	err = c.Watch(notificationSource, nil)
	if err != nil {
		return err
	}
	// handler for global RPC message and ,simply trigger a reconcile on every message
	nb.GlobalRPC.Handler = func(req *nb.RPCMessage) (interface{}, error) {
		logrus.Infof("RPC Handle: %+v", req)
		notificationSource.Queue.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		}})
		return nil, nil
	}

	return nil
}
