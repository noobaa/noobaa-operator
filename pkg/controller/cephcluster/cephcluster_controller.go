package cephcluster

import (
	"context"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	cephv1 "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"
	"github.com/sirupsen/logrus"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
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
	// Create a controller to react to ceph cluster changes
	if !util.KubeList(&cephv1.CephClusterList{}, client.InNamespace(options.Namespace)) {
		return nil
	}

	logrus.Info("CephCluster CR is detected in the cluster, adding a ceph cluster controller")
	c, err := controller.New("noobaa-controller", mgr, controller.Options{
		MaxConcurrentReconciles: 1,
		Reconciler:              reconcile.Func(doReconcile),
		SkipNameValidation: &[]bool{true}[0],
	})
	if err != nil {
		return err
	}

	// Predicate that allow us to log event that are being queued
	logEventsPredicate := util.LogEventsPredicate{}

	// Watch for cephcluster resource changes
	err = c.Watch(source.Kind[client.Object](mgr.GetCache(), &cephv1.CephCluster{}, &handler.EnqueueRequestForObject{},
		&CephCapacityChangedPredicate{}, &logEventsPredicate))
	if err != nil {
		return err
	}

	return nil
}

// React to cephcluster capacity changes
func doReconcile(context context.Context, req reconcile.Request) (reconcile.Result, error) {
	log := logrus.WithField("cephcluster", req.Namespace+"/"+req.Name)
	res := reconcile.Result{}

	nooBaa := nbv1.NooBaa{}
	nooBaa.Name = options.SystemName
	nooBaa.Namespace = req.Namespace
	if !system.CheckSystem(&nooBaa) {
		log.Infof("NooBaa not found or already deleted. Skip reconcile.")
		return res, nil
	}

	cephCluster := cephv1.CephCluster{}
	cephCluster.Name = req.Name
	cephCluster.Namespace = req.Namespace
	if !util.KubeCheck(&cephCluster) {
		log.Fatalf(`❌ Could not find CephCluster CR %q in namespace %q`,
			cephCluster.Name, cephCluster.Namespace)

		return res, nil
	}

	// Ensure we have ceph status to read from
	if cephCluster.Status.CephStatus == nil {
		return res, nil
	}
	cephCapacity := &cephCluster.Status.CephStatus.Capacity

	// Get a noobaa client
	sysClient, err := system.Connect(false)
	if err != nil {
		logrus.Errorf("Could not connect to system %+v", err)
		return res, err
	}
	nbClient := sysClient.NBClient

	backingStoreList := nbv1.BackingStoreList{}
	util.KubeList(&backingStoreList, client.InNamespace(options.Namespace))
	for _, bs := range backingStoreList.Items {
		if bs.Spec.S3Compatible != nil && bs.ObjectMeta.Annotations != nil {
			if _, ok := bs.ObjectMeta.Annotations["rgw"]; ok {
				avail := nb.UInt64ToBigInt(cephCapacity.AvailableBytes)
				err := nbClient.UpdateCloudPoolAPI(nb.UpdateCloudPoolParams{
					Name:              bs.Name,
					AvailableCapacity: &avail,
				})
				if err != nil {
					return res, err
				}
			}
		}
	}

	return res, nil
}

// CephCapacityChangedPredicate will filter an
type CephCapacityChangedPredicate struct {
	predicate.Funcs
}

// Create will fire whenever the contoller is registered or a new ceph cluster is create
func (p CephCapacityChangedPredicate) Create(e event.CreateEvent) bool {
	return true
}

// Update implements the update event trap for LabelsChangedPredicate
func (p CephCapacityChangedPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		return false
	}
	oldCephCluster, oldCastOk := e.ObjectOld.(*cephv1.CephCluster)
	newCephCluster, newCastOk := e.ObjectNew.(*cephv1.CephCluster)
	if !oldCastOk || !newCastOk {
		return false
	}

	oldCephStatus := oldCephCluster.Status.CephStatus
	newCephStatus := newCephCluster.Status.CephStatus
	if oldCephStatus == nil {
		return newCephStatus != nil
	}

	return oldCephStatus.Capacity.AvailableBytes != newCephStatus.Capacity.AvailableBytes
}
