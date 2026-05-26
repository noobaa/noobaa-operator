package validations

import (
	"fmt"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ValidateAccountDefaultResource is the public entry point for validating that a NooBaaAccount's
// DefaultResource is not an archive NamespaceStore.
func ValidateAccountDefaultResource(na nbv1.NooBaaAccount) error {
	if na.Spec.DefaultResource == "" {
		return nil
	}
	isBackingStore, _ := checkResourceBackingStore(na.Spec.DefaultResource)
	isNamespaceStore, namespaceStoreObj := checkResourceNamespaceStore(na.Spec.DefaultResource)

	if !isBackingStore && !isNamespaceStore {
		return util.NewPersistentError("MissingDefaultResource",
			fmt.Sprintf("Account %q default resource %q was not found", na.Name, na.Spec.DefaultResource))
	} else if isBackingStore && isNamespaceStore {
		return util.NewPersistentError("MissingDefaultResource",
			fmt.Sprintf("BackingStore and NamespaceStore should not have the same name: %q", na.Spec.DefaultResource))
	}
	if util.IsArchiveNamespaceStore(namespaceStoreObj) {
		return util.NewPersistentError("InvalidDefaultResource",
			fmt.Sprintf("NamespaceStore %q is an archive store and cannot be used as a default resource", namespaceStoreObj.Name))
	}
	return nil
}

// ValidateRemoveNSFSConfig validates that the NSFS config was not removed
func ValidateRemoveNSFSConfig(na nbv1.NooBaaAccount, oldNa nbv1.NooBaaAccount) error {
	if oldNa.Spec.NsfsAccountConfig != nil && na.Spec.NsfsAccountConfig == nil {
		return util.ValidationError{
			Msg: "Removing the NsfsAccountConfig is unsupported",
		}
	}
	return nil
}

// ValidateNSFSConfig validates that the NSFS config files were set properly
func ValidateNSFSConfig(na nbv1.NooBaaAccount) error {
	nsfsConf := na.Spec.NsfsAccountConfig

	if nsfsConf == nil {
		return nil
	}

	//UID validation
	if *nsfsConf.UID < 0 {
		return util.ValidationError{
			Msg: "UID must be a whole positive number",
		}
	}

	//GID validation
	if *nsfsConf.GID < 0 {
		return util.ValidationError{
			Msg: "GID must be a whole positive number",
		}
	}

	return nil
}

// checkResourceBackingStore checks if a resourceName exists and if BackingStore
// returns true if the resource exists and is a BackingStore, and also returns the BackingStore object if it exists
func checkResourceBackingStore(resourceName string) (bool, *nbv1.BackingStore) {
	// check that a backing store exists
	resourceBackingStore := &nbv1.BackingStore{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: options.Namespace,
		},
	}
	res := util.KubeCheckQuiet(resourceBackingStore)
	return res, resourceBackingStore
}

// checkResourceNamespaceStore checks if a resourceName exists and if NamespaceStore
// returns true if the resource exists and is a NamespaceStore, and also returns the NamespaceStore object if it exists
func checkResourceNamespaceStore(resourceName string) (bool, *nbv1.NamespaceStore) {
	// check that a namespace store exists
	resourceNamespaceStore := &nbv1.NamespaceStore{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: options.Namespace,
		},
	}
	res := util.KubeCheckQuiet(resourceNamespaceStore)
	return res, resourceNamespaceStore
}
