package noobaa

import (
	"context"
	"strings"

	cnpgv1 "github.com/cloudnative-pg/cloudnative-pg/api/v1"
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
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// cnpgClusterNameSuffix is appended to the NooBaa CR name to form the CNPG Cluster name.
// Must stay in sync with pgClusterSuffix in pkg/system/db_reconciler.go.
const cnpgClusterNameSuffix = "-db-pg-cluster"

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
					mgr.GetEventRecorder("noobaa-operator"),
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

	if err := watchCNPGCluster(c, mgr, logEventsPredicate); err != nil {
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

// watchCNPGCluster watches CNPG Cluster deletion so dbRecovery starts immediately
// after the user deletes noobaa-db-pg-cluster. The watch is skipped when the CNPG
// CRD is not installed (KMS/kind tests, standalone DB) so the operator can still start.
func watchCNPGCluster(c controller.Controller, mgr manager.Manager, logEventsPredicate util.LogEventsPredicate) error {
	if !util.KubeList(&cnpgv1.ClusterList{}, client.InNamespace(options.Namespace)) {
		logrus.Info("CNPG Cluster CRD is not available, skipping Cluster watch")
		return nil
	}

	return c.Watch(source.Kind[client.Object](mgr.GetCache(), &cnpgv1.Cluster{},
		handler.EnqueueRequestsFromMapFunc(mapCNPGClusterToNooBaa),
		&cnpgClusterDeletePredicate{}, &logEventsPredicate))
}

// mapCNPGClusterToNooBaa enqueues the NooBaa system named by the CNPG Cluster name
// (<noobaa-name>-db-pg-cluster) in the same namespace.
func mapCNPGClusterToNooBaa(_ context.Context, obj client.Object) []reconcile.Request {
	name := obj.GetName()
	if !strings.HasSuffix(name, cnpgClusterNameSuffix) {
		return nil
	}
	return []reconcile.Request{{
		NamespacedName: types.NamespacedName{
			Name:      options.SystemName,
			Namespace: options.Namespace,
		},
	}}
}

// cnpgClusterDeletePredicate queues reconcile only when a CNPG Cluster is deleted.
// Recovery from dbRecovery is triggered by deleting the Cluster; without this watch
// the operator stays idle until some other NooBaa-owned resource changes.
type cnpgClusterDeletePredicate struct {
	predicate.Funcs
}

func (p cnpgClusterDeletePredicate) Create(event.CreateEvent) bool { return false }
func (p cnpgClusterDeletePredicate) Delete(e event.DeleteEvent) bool {
	if e.Object != nil {
		logrus.Infof("Delete event detected for CNPG cluster %s (%s), queuing Reconcile",
			e.Object.GetName(), e.Object.GetNamespace())
	}
	return true
}
func (p cnpgClusterDeletePredicate) Update(event.UpdateEvent) bool   { return false }
func (p cnpgClusterDeletePredicate) Generic(event.GenericEvent) bool { return false }
