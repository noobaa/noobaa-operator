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

// PodDeleter encapsulates an operation of
// deleting pods on node
type PodDeleter struct {
	Client   client.Client
	NodeName string
}

// DeletePodsOnNode force delete NooBaa pods on the given node
func (pd *PodDeleter) DeletePodsOnNode() error {
	log := logrus.WithField(Name, pd.NodeName)

	// looking for noobaa pods running on the failing node in the watched namespace
	labelOption := client.MatchingLabels{"app": "noobaa"}
	namespaceOption := client.InNamespace(options.Namespace)
	nodeOption := client.MatchingFields{"spec.nodeName": pd.NodeName}

	// fetch the noobaa pods from the api server
	podList := &corev1.PodList{}
	if !util.KubeList(podList, labelOption, namespaceOption, nodeOption) {
		return errors.Errorf("failed to list noobaa pods on the node %v in namespace %v", pd.NodeName, options.Namespace)
	}

	// delete the found pods
	var gracePeriod int64 = 0
	deleteOpts := client.DeleteOptions{GracePeriodSeconds: &gracePeriod}
	for _, pod := range podList.Items {
		if err := pd.Client.Delete(context.Background(), &pod, &deleteOpts); err != nil {
			log.Warningf("❌ pod %v/%v, node %v: deletion failed: %v", pod.Namespace, pod.Name, pd.NodeName, err)
			return err
		}
		log.Infof("✅ pod %v/%v, node %v: deletion succeeded", pod.Namespace, pod.Name, pd.NodeName)
	}

	return nil
}

// HAC is a high availability controller that watches the nodes Ready -> NotReady transitions
// and deletes noobaa pods on the node that became unvailable
type HAC struct {
	*PodDeleter
}

// NewHAC initializes a high availability controller reconciler
func NewHAC(
	req types.NamespacedName,
	client client.Client,
) *HAC {
	// k8s client and the node that failed
	pd := &PodDeleter{client, req.Name}

	return &HAC{pd}
}

// Reconcile is called when a node in the cluster transitions
// from Ready to NotReady state
// go over the noobaa pods running on this node and force their deletion
func (hac *HAC) Reconcile() (reconcile.Result, error) {
	res := reconcile.Result{}
	log := logrus.WithField(Name, hac.NodeName)
	log.Warningf("❌ node %v became NotReady", hac.NodeName)

	if err := hac.DeletePodsOnNode(); err != nil {
		return res, errors.Errorf("failed to delete noobaa pods on the node %v in namespace %v", hac.NodeName, options.Namespace)
	}

	return res, nil
}
