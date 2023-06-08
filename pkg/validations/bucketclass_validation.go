package validations

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateBucketClass bucket class
func ValidateBucketClass(bc *nbv1.BucketClass) error {
	if bc == nil {
		return nil
	}
	if bc.Spec.NamespacePolicy != nil {
		if err := ValidateNSFSSingleBC(bc); err != nil {
			return err
		}
	}
	if bc.Spec.PlacementPolicy != nil {
		if err := ValidateTiersNumber(bc.Spec.PlacementPolicy.Tiers); err != nil {
			return err
		}
	}

	return ValidateQuotaConfig(bc.Name, bc.Spec.Quota)
}

// ValidateImmutLabelChange validates that immutable labels are not changed
func ValidateImmutLabelChange(bc *nbv1.BucketClass, oldBC *nbv1.BucketClass, immuts map[string]struct{}) error {
	if bc == nil || oldBC == nil {
		return nil
	}

	for immutableLabel := range immuts {
		val1 := oldBC.GetLabels()[immutableLabel]
		val2 := bc.GetLabels()[immutableLabel]

		if val1 != val2 {
			return util.ValidationError{
				Msg: fmt.Sprintf("immutable label %q cannot be changed", immutableLabel),
			}
		}
	}

	return nil
}

// ValidateQuotaConfig validates the quota values
func ValidateQuotaConfig(bcName string, quota *nbv1.Quota) error {
	if quota == nil {
		return nil
	}

	return util.ValidateQuotaConfig(bcName, quota.MaxSize, quota.MaxObjects)
}

// ValidateTiersNumber validates the provided number of tiers is 1 or 2
func ValidateTiersNumber(tiers []nbv1.Tier) error {
	if len(tiers) != 1 && len(tiers) != 2 {
		return util.ValidationError{
			Msg: "unsupported number of tiers, bucketclass supports only 1 or 2 tiers",
		}
	}
	return nil
}

// GetBucketclassNamespaceStoreArray returns an array of namespacestores of the provided bc
func GetBucketclassNamespaceStoreArray(namespacePolicy *nbv1.NamespacePolicy) []string {
	log := util.Logger()
	log.Infof("creating namespace store array %+v from namespace policy", namespacePolicy)
	var namespaceStoresArr []string

	switch namespacePolicy.Type {
	case nbv1.NSBucketClassTypeCache:
		namespaceStoresArr = append(namespaceStoresArr, namespacePolicy.Cache.HubResource)
	case nbv1.NSBucketClassTypeMulti:
		if namespacePolicy.Multi.WriteResource == "" {
			namespaceStoresArr = namespacePolicy.Multi.ReadResources
		} else {
			namespaceStoresArr = append(namespacePolicy.Multi.ReadResources, namespacePolicy.Multi.WriteResource)
		}
	case nbv1.NSBucketClassTypeSingle:
		namespaceStoresArr = append(namespaceStoresArr, namespacePolicy.Single.Resource)
	}
	log.Infof("created namespace store array successfully %+v", namespaceStoresArr)
	return namespaceStoresArr
}

// ValidatePlacementPolicy validates backingstore existance and readiness
func ValidatePlacementPolicy(placementPolicy *nbv1.PlacementPolicy, namespace string) error {
	log := util.Logger()
	log.Infof("validating placement policy %+v", placementPolicy)
	if placementPolicy == nil {
		return nil
	}

	for i := range placementPolicy.Tiers {
		tier := &placementPolicy.Tiers[i]
		for _, backingStoreName := range tier.BackingStores {
			backStore := &nbv1.BackingStore{
				TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
				ObjectMeta: metav1.ObjectMeta{
					Name:      backingStoreName,
					Namespace: namespace,
				},
			}
			if !util.KubeCheck(backStore) {
				return util.NewPersistentError("MissingBackingStore",
					fmt.Sprintf("NooBaa BackingStore %q not found or deleted", backingStoreName))
			}
			if backStore.Status.Phase == nbv1.BackingStorePhaseRejected {
				return util.NewPersistentError("RejectedBackingStore",
					fmt.Sprintf("NooBaa BackingStore %q is in rejected phase", backingStoreName))
			}
			if backStore.Status.Phase != nbv1.BackingStorePhaseReady {
				return fmt.Errorf("NooBaa BackingStore %q is not yet ready", backingStoreName)
			}
		}
	}
	log.Infof("validated placement policy successfully %+v", placementPolicy)
	return nil
}

// ValidateNamespacePolicy validates namespacestores existance and readiness
func ValidateNamespacePolicy(namespacePolicy *nbv1.NamespacePolicy, namespace string) error {
	log := util.Logger()
	log.Infof("validating namespace policy %+v", namespacePolicy)
	if namespacePolicy == nil {
		return nil
	}

	namespaceStoresArr := GetBucketclassNamespaceStoreArray(namespacePolicy)
	// check that namespace stores exists and their phase it ready
	for _, name := range namespaceStoresArr {
		nsStore := &nbv1.NamespaceStore{
			TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
		}
		if !util.KubeCheck(nsStore) {
			return util.NewPersistentError("MissingNamespaceStore",
				fmt.Sprintf("NooBaa NamespaceStore %q not found or deleted", name))
		}
		if nsStore.Status.Phase == nbv1.NamespaceStorePhaseRejected {
			return util.NewPersistentError("RejectedNamespaceStore",
				fmt.Sprintf("NooBaa NamespaceStore %q is in rejected phase", name))
		}
		if nsStore.Status.Phase != nbv1.NamespaceStorePhaseReady {
			return fmt.Errorf("NooBaa NamespaceStore %q is not yet ready", name)
		}
	}
	log.Infof("validated namespace policy successfully %+v", namespacePolicy)
	return nil
}

// ValidateNSFSSingleBC validates that bucketclass configured to NS of type NSFS it will only be of type Single.
func ValidateNSFSSingleBC(bc *nbv1.BucketClass) error {
	if bc.Spec.NamespacePolicy.Type == nbv1.NSBucketClassTypeSingle {
		return nil
	}
	namespaceStoresArr := GetBucketclassNamespaceStoreArray(bc.Spec.NamespacePolicy)
	for _, name := range namespaceStoresArr {
		nsStore := &nbv1.NamespaceStore{
			TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: bc.Namespace,
			},
		}
		if !util.KubeCheck(nsStore) {
			return fmt.Errorf("failed to KubeCheck namespacestore in NSFS single bucketclass type validation")
		}
		if nsStore.Spec.Type == nbv1.NSStoreTypeNSFS {
			return util.ValidationError{
				Msg: fmt.Sprintf("invalid namespaceStore types, nsfs namespacestore %q is allowed on bucketclass of type single", name),
			}
		}
	}
	return nil
}
