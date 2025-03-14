package obc

import (
	"context"
	"reflect"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/noobaa/noobaa-operator/v5/pkg/obc"
	"k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add starts running the noobaa bucket provisioner
func Add(mgr manager.Manager) error {

	// Creating informer to watch OBC
	informer, err := mgr.GetCache().GetInformer(context.TODO(), &nbv1.ObjectBucketClaim{})
	if err != nil {
		return err
	}

	_, _ = informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldOBC, okOld := oldObj.(*nbv1.ObjectBucketClaim)
			newOBC, okNew := newObj.(*nbv1.ObjectBucketClaim)
			if !okOld || !okNew {
				return
			}

			// supports only updating of bucket tagging with OBC labels
			if !reflect.DeepEqual(oldOBC.Labels, newOBC.Labels) {
				sysClient, err := system.Connect(false)
				if err != nil {
					util.Logger().Errorf("Failed to connect to the system for OBC %s: %v", newOBC.Name, err)
					return
				}
				if err := obc.UpdateBucketTagging(sysClient, newOBC); err != nil {
					util.Logger().Errorf("Failed to update bucket tagging for OBC %s: %v", newOBC.Name, err)
					return
				}
			}
		},
	})

	return obc.RunProvisioner(
		mgr.GetClient(),
		mgr.GetScheme(),
		mgr.GetEventRecorderFor("noobaa-operator"),
	)
}
