package noobaa

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// NotificationSource specifies a queue of notifications
type NotificationSource struct {
	Queue workqueue.TypedRateLimitingInterface[reconcile.Request]
}

// Start will setup s.Queue field
func (s *NotificationSource) Start(context context.Context, q workqueue.TypedRateLimitingInterface[reconcile.Request]) error {
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
			func(context context.Context, req reconcile.Request) (reconcile.Result, error) {
				return system.NewReconciler(
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

	// Predicate that filter events that noobaa is not their owner
	filterForOwnerPredicate := util.FilterForOwner{
		OwnerType: &nbv1.NooBaa{},
		Scheme:    mgr.GetScheme(),
	}

	// Predicate that allows events that only change spec, labels or finalizers will log any allowed events
	// This will stop infinite reconciles that triggered by status or irrelevant metadata changes
	noobaaPredicate := util.ComposePredicates(
		predicate.GenerationChangedPredicate{},
		util.LabelsChangedPredicate{},
		util.FinalizersChangedPredicate{},
	)

	// Watch for changes on resources to trigger reconcile
	ownerHandler := handler.EnqueueRequestForOwner(
		mgr.GetScheme(),
		mgr.GetRESTMapper(),
		&nbv1.NooBaa{},
		handler.OnlyControllerOwner(),
	)

	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &nbv1.NooBaa{}, &handler.EnqueueRequestForObject{},
		noobaaPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &appsv1.StatefulSet{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.Service{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.Pod{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &appsv1.Deployment{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &corev1.ConfigMap{}, ownerHandler, &filterForOwnerPredicate, &logEventsPredicate))
	if err != nil {
		return err
	}

	storageClassHandler := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, mo client.Object) []reconcile.Request {
		sc, ok := mo.(*storagev1.StorageClass)
		if !ok || sc.Provisioner != options.ObjectBucketProvisionerName() {
			return nil
		}
		return []reconcile.Request{{
			NamespacedName: types.NamespacedName{
				Name:      options.SystemName,
				Namespace: options.Namespace,
			},
		}}
	},
	)

	// Watch for StorageClass changes to trigger reconcile and recreate it when deleted
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &storagev1.StorageClass{}, storageClassHandler, &logEventsPredicate))
	if err != nil {
		return err
	}
	// watch on notificationSource in order to keep the controller work queue
	notificationSource := &NotificationSource{}
	err = c.Watch(notificationSource)
	if err != nil {
		return err
	}

	// handler for global RPC message and ,simply trigger a reconcile on every message
	nb.GlobalRPC.Handler = func(req *nb.RPCMessage) (interface{}, error) {
		logrus.Infof("RPC Handle: {Op: %s, API: %s, Method: %s, Error: %s, Params: %+v}", req.Op, req.API, req.Method, req.Error, req.Params)
		notificationSource.Queue.AddRateLimited(reconcile.Request{NamespacedName: types.NamespacedName{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		}})
		return nil, nil
	}

	return nil
}
