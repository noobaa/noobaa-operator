package hac

// NooBaa Component Rescheduling
// see doc/high-availability-controller.md

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// Name of this controller
const Name = "high-availability-controller"

// HAC is a high availability controller that watches the nodes Ready -> NotReady transitions
// and deletes noobaa pods on the node that became unvailable
type HAC struct {
	Request  types.NamespacedName
	Client   client.Client
	Ctx      context.Context
	Logger   *logrus.Entry
	Node     *corev1.Node
}

// NewHAC initializes a high availability controller reconciler
func NewHAC(
	req types.NamespacedName,
	client client.Client,
) *HAC {
	hac := &HAC{
		Request:          req,
		Client:           client,
		Ctx:              context.Background(),
		Logger:           logrus.WithField(Name, req.Name),
	}

	// the node that failed
	hac.Node = &corev1.Node{}
	hac.Node.Name = req.Name

	return hac
}

// Reconcile is called when a node in the cluster transitions
// from Ready to NotReady state
// go over the noobaa pods running on this node and force their deletion
func (hac *HAC) Reconcile() (reconcile.Result, error) {
	res := reconcile.Result{}
	log := hac.Logger
	log.Warningf("❌ node %v became NotReady", hac.Node.Name)

	// looking for noobaa pods running on the failing node in the watched namespace
	labelOption := client.MatchingLabels{"app": "noobaa"}
	namespaceOption := client.InNamespace(options.Namespace)
	nodeOption := client.MatchingFields{"spec.nodeName": hac.Node.Name}

	// fetch the noobaa pods from the api server
	podList := &corev1.PodList{}
	if !util.KubeList(podList, labelOption, namespaceOption, nodeOption) {
		return res, errors.Errorf("failed to list noobaa pods on the node %v in namespace %v", hac.Node.Name, options.Namespace)
	}

	// delete the found pods
	var gracePeriod int64 = 0
	deleteOpts := client.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	for _, pod := range podList.Items {
		if err := hac.Client.Delete(hac.Ctx, &pod, &deleteOpts); err != nil {
			log.Warningf("❌ pod %v/%v, node %v: deletion failed: %v", pod.Namespace, pod.Name, hac.Node.Name, err)
			return res, err
		}
		log.Infof("✅ pod %v/%v, node %v: deletion succeeded", pod.Namespace, pod.Name, hac.Node.Name)
	}

	return res, nil
}
