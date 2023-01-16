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
func GetBucketclassNamespaceStoreArray(bc *nbv1.BucketClass) []string {
	var namespaceStoresArr []string

	switch bc.Spec.NamespacePolicy.Type {
	case nbv1.NSBucketClassTypeCache:
		namespaceStoresArr = append(namespaceStoresArr, bc.Spec.NamespacePolicy.Cache.HubResource)
	case nbv1.NSBucketClassTypeMulti:
		if bc.Spec.NamespacePolicy.Multi.WriteResource == "" {
			namespaceStoresArr = bc.Spec.NamespacePolicy.Multi.ReadResources
		} else {
			namespaceStoresArr = append(bc.Spec.NamespacePolicy.Multi.ReadResources,
				bc.Spec.NamespacePolicy.Multi.WriteResource)
		}
	case nbv1.NSBucketClassTypeSingle:
		namespaceStoresArr = append(namespaceStoresArr, bc.Spec.NamespacePolicy.Single.Resource)
	}

	return namespaceStoresArr
}

// ValidateNSFSSingleBC validates that bucketclass configured to NS of type NSFS it will only be of type Single.
func ValidateNSFSSingleBC(bc *nbv1.BucketClass) error {
	if bc.Spec.NamespacePolicy.Type == nbv1.NSBucketClassTypeSingle {
		return nil
	}
	namespaceStoresArr := GetBucketclassNamespaceStoreArray(bc)
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
