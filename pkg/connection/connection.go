package connection

import (
	"fmt"
	"strings"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/constants"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StoreType string

const (
	StoreTypeS3Compatible StoreType = "s3-compatible"
	StoreTypeIBMCOS       StoreType = "ibm-cos"
)

type matchedStore struct {
	isBackingStore bool
	Store          client.Object
	identity       string
	secret         string
	authMethod     nb.CloudAuthMethod
	bucket         string
	endpointType   nb.EndpointType
}

func (m *matchedStore) name() string {
	if m.Store == nil {
		return ""
	}
	name := m.Store.GetName()
	if m.isBackingStore {
		return "BackingStore/" + name
	}
	return "NamespaceStore/" + name
}

type connEntry struct {
	name         string
	endpointType nb.EndpointType
}

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connection",
		Short: "Manage backingstores/namespacestores connections",
	}
	cmd.AddCommand(
		CmdUpdate(),
	)
	return cmd
}

// CmdUpdate returns a CLI command
func CmdUpdate() *cobra.Command {
	cmd := &cobra.Command{
		Use: "update",
		Short: `Update connection properties across all matching backing stores and namespace stores, 
		currently the only supported property is endpoint`,
		Run: RunUpdate,
	}
	cmd.Flags().String("old-endpoint", "", "The current endpoint URL to replace")
	cmd.Flags().String("new-endpoint", "", "The new endpoint URL to set")
	return cmd
}

// RunUpdate implements the endpoint update flow
func RunUpdate(cmd *cobra.Command, args []string) {
	log := util.Logger()
	oldEndpoint, _ := cmd.Flags().GetString("old-endpoint")
	newEndpoint, _ := cmd.Flags().GetString("new-endpoint")
	oldEndpoint = strings.TrimSpace(oldEndpoint)
	newEndpoint = strings.TrimSpace(newEndpoint)

	// Validate flags
	if oldEndpoint == "" || newEndpoint == "" {
		log.Fatalf("both --old-endpoint and --new-endpoint are required")
	}

	// List and filter all matching stores using oldEndpoint
	matched := findMatchingStores(oldEndpoint)
	if len(matched) == 0 {
		log.Fatalf("no matching stores found for endpoint %q", oldEndpoint)
	}

	log.Infof("Found %d store(s) matching endpoint %q:", len(matched), oldEndpoint)
	for _, m := range matched {
		log.Infof("  - %s", m.name())
	}

	// Read secrets and buckets
	if err := loadSecretsAndBuckets(matched); err != nil {
		log.Fatalf("failed to read secrets: %s", err)
	}

	// Pre-validate each store's bucket on the new endpoint
	sysClient, err := system.ConnectAuto()
	if err != nil {
		log.Fatalf("failed to connect to NooBaa system: %s", err)
	}
	nbClient := sysClient.NBClient

	failedValidations := validateStores(nbClient, matched, newEndpoint)
	if len(failedValidations) > 0 {
		log.Errorf("Pre-validation failed for the following stores:")
		for _, f := range failedValidations {
			log.Errorf("  - %s", f)
		}
		log.Fatalf("Aborting. No changes have been made.")
	}
	log.Infof("All stores passed pre-validation on new endpoint %q", newEndpoint)

	// Set pause annotation on all matching stores
	if errs := setPauseAnnotation(matched, true); len(errs) > 0 {
		removePauseAnnotations(matched)
		log.Fatalf("failed to set pause annotation: %s", errs)
	}
	log.Infof("Paused reconciliation for all matching stores")

	// Patch all CR specs with new endpoint
	var patched []matchedStore
	if err := patchEndpoints(matched, newEndpoint, &patched); err != nil {
		log.Errorf("failed to patch endpoints: %s", err)
		rollback(patched, oldEndpoint, matched, nil, nil)
		log.Fatalf("Rollback complete. Endpoint update aborted.")
	}
	log.Infof("Patched all %d store specs with new endpoint", len(patched))

	// Discover unique connections by name from core
	uniqueConns, err := findConnections(nbClient, oldEndpoint)
	if err != nil {
		log.Errorf("failed to find connections: %s", err)
		rollback(patched, oldEndpoint, matched, nil, nil)
		log.Fatalf("Rollback complete. Endpoint update aborted.")
	}
	log.Infof("Found %d unique connection(s) to update in NooBaa core", len(uniqueConns))

	// Update external connections, tracking successes for rollback
	var updatedConns []connEntry
	for _, entry := range uniqueConns {
		err := nbClient.UpdateExternalConnectionAPI(nb.UpdateExternalConnectionParams{
			Name: entry.name,
			EndpointInfo: &nb.EndpointInfo{
				Endpoint:     newEndpoint,
				EndpointType: entry.endpointType,
			},
		})
		if err != nil {
			log.Errorf("failed to update connection %q: %s", entry.name, err)
			rollback(patched, oldEndpoint, matched, nbClient, updatedConns)
			log.Fatalf("Rollback complete. Endpoint update aborted.")
		}
		updatedConns = append(updatedConns, entry)
		log.Infof("Updated connection %q in NooBaa core", entry.name)
	}

	if errs := setPauseAnnotation(matched, false); len(errs) > 0 {
		log.Warnf("failed to remove pause annotations: %s (stores were updated successfully, but annotations may need manual cleanup)", errs)
	}

	log.Infof("Endpoint update completed successfully:")
	log.Infof("  Old endpoint: %s", oldEndpoint)
	log.Infof("  New endpoint: %s", newEndpoint)
	log.Infof("  Stores updated: %d", len(matched))
	log.Infof("  Connections updated: %d", len(uniqueConns))
}

/*
Lists the BackingStore and NamespaceStore resources in the
configured namespace and returns the stores that match the specified endpoint.
*/
func findMatchingStores(endpoint string) []matchedStore {
	log := util.Logger()
	// List all stores in the current namespace
	backingStoreList := &nbv1.BackingStoreList{}
	namespaceStoreList := &nbv1.NamespaceStoreList{}
	ok := util.KubeList(backingStoreList, client.InNamespace(options.Namespace))
	if !ok {
		log.Fatalf("error listing backing stores in namespace %s", options.Namespace)
	}
	ok = util.KubeList(namespaceStoreList, client.InNamespace(options.Namespace))
	if !ok {
		log.Fatalf("error listing namespace stores in namespace %s", options.Namespace)
	}

	// Filter matching stores
	var matched []matchedStore
	for i := range backingStoreList.Items {
		bs := &backingStoreList.Items[i]
		if m, ok := matchBackingStore(bs, endpoint); ok {
			matched = append(matched, m)
		}
	}
	for i := range namespaceStoreList.Items {
		ns := &namespaceStoreList.Items[i]
		if m, ok := matchNamespaceStore(ns, endpoint); ok {
			matched = append(matched, m)
		}
	}
	return matched
}

/*
Validates the external connection configuration for each
matched store for the specified endpoint.
It returns a list of validation errors, with each error identifying the
store that failed validation.
*/
func validateStores(nbClient nb.Client, matched []matchedStore, newEndpoint string) []string {
	var failedValidations []string
	for _, m := range matched {
		params := nb.CheckExternalConnectionParams{
			Endpoint:     newEndpoint,
			EndpointType: m.endpointType,
			Identity:     nb.MaskedString(m.identity),
			Secret:       nb.MaskedString(m.secret),
			AuthMethod:   m.authMethod,
		}
		if m.isBackingStore {
			params.Bucket = m.bucket
		}
		reply, err := nbClient.CheckExternalConnectionAPI(params)
		if err != nil {
			failedValidations = append(failedValidations, fmt.Sprintf("%s: %v", m.name(), err))
			continue
		}
		if reply.Status != nb.ExternalConnectionSuccess {
			failedValidations = append(failedValidations, fmt.Sprintf("%s (bucket %q): status=%s error=%s",
				m.name(), m.bucket, reply.Status, reply.Error.Message))
		}
	}
	return failedValidations
}

/*
findConnections retrieves the external connections configured in the system
and returns the unique connections associated with the specified endpoint.
The returned map is keyed by connection name.
*/
func findConnections(nbClient nb.Client, endpoint string) (map[string]connEntry, error) {
	systemInfo, err := nbClient.ReadSystemAPI()
	if err != nil {
		return nil, fmt.Errorf("failed to read system info: %s", err)
	}

	uniqueConns := map[string]connEntry{}
	for i := range systemInfo.Accounts {
		account := &systemInfo.Accounts[i]
		for j := range account.ExternalConnections.Connections {
			conn := &account.ExternalConnections.Connections[j]
			if conn.Endpoint == endpoint {
				if _, exists := uniqueConns[conn.Name]; !exists {
					uniqueConns[conn.Name] = connEntry{
						name:         conn.Name,
						endpointType: conn.EndpointType,
					}
				}
			}
		}
	}
	return uniqueConns, nil
}

func matchBackingStore(bs *nbv1.BackingStore, oldEndpoint string) (matchedStore, bool) {
	if bs != nil {
		switch bs.Spec.Type {
		case nbv1.StoreTypeS3Compatible:
			if bs.Spec.S3Compatible != nil && bs.Spec.S3Compatible.Endpoint == oldEndpoint {
				return matchedStore{isBackingStore: true, Store: bs, endpointType: nb.EndpointTypeS3Compat}, true
			}
		case nbv1.StoreTypeIBMCos:
			if bs.Spec.IBMCos != nil && bs.Spec.IBMCos.Endpoint == oldEndpoint {
				return matchedStore{isBackingStore: true, Store: bs, endpointType: nb.EndpointTypeIBMCos}, true
			}
		}
	}
	return matchedStore{}, false
}

func matchNamespaceStore(ns *nbv1.NamespaceStore, oldEndpoint string) (matchedStore, bool) {
	if ns != nil {
		switch ns.Spec.Type {
		case nbv1.NSStoreTypeS3Compatible:
			if ns.Spec.S3Compatible != nil && ns.Spec.S3Compatible.Endpoint == oldEndpoint {
				return matchedStore{isBackingStore: false, Store: ns, endpointType: nb.EndpointTypeS3Compat}, true
			}
		case nbv1.NSStoreTypeIBMCos:
			if ns.Spec.IBMCos != nil && ns.Spec.IBMCos.Endpoint == oldEndpoint {
				return matchedStore{isBackingStore: false, Store: ns, endpointType: nb.EndpointTypeIBMCos}, true
			}
		}
	}
	return matchedStore{}, false
}

func loadSecretsAndBuckets(stores []matchedStore) error {
	for i := range stores {
		m := &stores[i]
		if m.isBackingStore {
			if err := loadBackingStoreSecretsAndBucket(m); err != nil {
				return err
			}
		} else {
			if err := loadNamespaceStoreSecretsAndBucket(m); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadBackingStoreSecretsAndBucket(m *matchedStore) error {
	bs := m.Store
	storeName := bs.GetName()
	secretRef, err := util.GetBackingStoreSecret(bs.(*nbv1.BackingStore))
	if err != nil {
		return fmt.Errorf("BackingStore %q: %w", storeName, err)
	}
	secret, err := util.GetSecretFromSecretReference(secretRef)
	if err != nil {
		return fmt.Errorf("BackingStore %q: %w", storeName, err)
	}
	if secret == nil {
		return fmt.Errorf("BackingStore %q: secret not found", storeName)
	}

	extractStoreCredentials(m, secret)

	bucket, err := util.GetBackingStoreTargetBucket(bs.(*nbv1.BackingStore))
	if err != nil {
		return fmt.Errorf("BackingStore %q: %w", storeName, err)
	}
	m.bucket = bucket
	return nil
}

func extractStoreCredentials(m *matchedStore, secret *corev1.Secret) {
	switch m.Store.(type) {
	case *nbv1.BackingStore:
		bs := m.Store.(*nbv1.BackingStore)
		switch bs.Spec.Type {
		case nbv1.StoreTypeS3Compatible:
			m.identity, m.secret = extractCredentials(
				secret,
				StoreTypeS3Compatible,
			)
			if bs.Spec.S3Compatible != nil {
				m.authMethod = nb.CloudAuthMethod(
					extractSignatureVersion(bs.Spec.S3Compatible.SignatureVersion),
				)
			}
		case nbv1.StoreTypeIBMCos:
			m.identity, m.secret = extractCredentials(
				secret,
				StoreTypeIBMCOS,
			)
			if bs.Spec.IBMCos != nil {
				m.authMethod = nb.CloudAuthMethod(
					extractSignatureVersion(bs.Spec.IBMCos.SignatureVersion),
				)
			}
		}
	case *nbv1.NamespaceStore:
		ns := m.Store.(*nbv1.NamespaceStore)
		switch ns.Spec.Type {
		case nbv1.NSStoreTypeS3Compatible:
			m.identity, m.secret = extractCredentials(
				secret,
				StoreTypeS3Compatible,
			)
			if ns.Spec.S3Compatible != nil {
				m.authMethod = nb.CloudAuthMethod(
					extractSignatureVersion(ns.Spec.S3Compatible.SignatureVersion),
				)
			}
		case nbv1.NSStoreTypeIBMCos:
			m.identity, m.secret = extractCredentials(
				secret,
				StoreTypeIBMCOS,
			)
			if ns.Spec.IBMCos != nil {
				m.authMethod = nb.CloudAuthMethod(
					extractSignatureVersion(ns.Spec.IBMCos.SignatureVersion),
				)
			}
		}
	}
}

func extractCredentials(secret *corev1.Secret, storeType StoreType) (string, string) {
	identityKey := ""
	secretKey := ""
	switch storeType {
	case StoreTypeS3Compatible:
		identityKey = "AWS_ACCESS_KEY_ID"
		secretKey = "AWS_SECRET_ACCESS_KEY"
	case StoreTypeIBMCOS:
		identityKey = "IBM_COS_ACCESS_KEY_ID"
		secretKey = "IBM_COS_SECRET_ACCESS_KEY"
	}
	return secret.StringData[identityKey], secret.StringData[secretKey]
}

func loadNamespaceStoreSecretsAndBucket(m *matchedStore) error {
	ns := m.Store
	storeName := ns.GetName()

	secretRef, err := util.GetNamespaceStoreSecret(ns.(*nbv1.NamespaceStore))
	if err != nil {
		return fmt.Errorf("NamespaceStore %q: %w", storeName, err)
	}
	secret, err := util.GetSecretFromSecretReference(secretRef)
	if err != nil {
		return fmt.Errorf("NamespaceStore %q: %w", storeName, err)
	}
	if secret == nil {
		return fmt.Errorf("NamespaceStore %q: secret not found", storeName)
	}

	extractStoreCredentials(m, secret)

	bucket, err := util.GetNamespaceStoreTargetBucket(ns.(*nbv1.NamespaceStore))
	if err != nil {
		return fmt.Errorf("NamespaceStore %q: %w", storeName, err)
	}
	m.bucket = bucket
	return nil
}

func extractSignatureVersion(sv nbv1.S3SignatureVersion) string {
	svString := ""
	switch sv {
	case nbv1.S3SignatureVersionV4:
		svString = "AWS_V4"
	case nbv1.S3SignatureVersionV2:
		svString = "AWS_V2"
	}
	return svString
}

func setPauseAnnotation(stores []matchedStore, pause bool) (errors []error) {
	for i := range stores {
		m := &stores[i]
		store := m.Store
		if store == nil {
			return append(errors, fmt.Errorf("encountered empty value for store"))
		}
		storeName := m.name()
		annotations := store.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}

		if pause {
			annotations[constants.PauseReconcile] = "true"
		} else {
			delete(annotations, constants.PauseReconcile)
		}
		store.SetAnnotations(annotations)

		if !util.KubeUpdate(store) {
			errors = append(errors, fmt.Errorf("failed to update Store %q annotations", storeName))
			if pause {
				return
			}
		}
	}
	return errors
}

func patchEndpoints(stores []matchedStore, newEndpoint string, patched *[]matchedStore) error {
	for i := range stores {
		m := &stores[i]
		if m.isBackingStore {
			bs := m.Store.(*nbv1.BackingStore)
			switch bs.Spec.Type {
			case nbv1.StoreTypeS3Compatible:
				bs.Spec.S3Compatible.Endpoint = newEndpoint
			case nbv1.StoreTypeIBMCos:
				bs.Spec.IBMCos.Endpoint = newEndpoint
			}
			m.Store = bs
		} else {
			ns := m.Store.(*nbv1.NamespaceStore)
			switch ns.Spec.Type {
			case nbv1.NSStoreTypeS3Compatible:
				ns.Spec.S3Compatible.Endpoint = newEndpoint
			case nbv1.NSStoreTypeIBMCos:
				ns.Spec.IBMCos.Endpoint = newEndpoint
			}
			m.Store = ns
		}

		if !util.KubeUpdate(m.Store) {
			return fmt.Errorf("failed to patch Store %q", m.name())
		}
		*patched = append(*patched, *m)
	}
	return nil
}

// rollback reverts already-patched stores to the old endpoint, reverts any
// already-updated core connections, and removes pause annotations from all stores.
// nbClient and updatedConns may be nil if no connections were updated yet.
func rollback(patched []matchedStore, oldEndpoint string, allStores []matchedStore, nbClient nb.Client, updatedConns []connEntry) {
	log := util.Logger()
	log.Infof("Rolling back %d patched store(s) to endpoint %q", len(patched), oldEndpoint)
	for i := range patched {
		m := &patched[i]
		if m.isBackingStore {
			bs := m.Store.(*nbv1.BackingStore)
			switch bs.Spec.Type {
			case nbv1.StoreTypeS3Compatible:
				bs.Spec.S3Compatible.Endpoint = oldEndpoint
			case nbv1.StoreTypeIBMCos:
				bs.Spec.IBMCos.Endpoint = oldEndpoint
			}
			m.Store = bs
		} else {
			ns := m.Store.(*nbv1.NamespaceStore)
			switch ns.Spec.Type {
			case nbv1.NSStoreTypeS3Compatible:
				ns.Spec.S3Compatible.Endpoint = oldEndpoint
			case nbv1.NSStoreTypeIBMCos:
				ns.Spec.IBMCos.Endpoint = oldEndpoint
			}
			m.Store = ns
		}
		if !util.KubeUpdate(m.Store) {
			log.Errorf("failed to rollback Store %q", m.name())
		}
	}

	if nbClient != nil && len(updatedConns) > 0 {
		log.Infof("Rolling back %d updated connection(s) to endpoint %q", len(updatedConns), oldEndpoint)
		var failedReverts []string
		for _, entry := range updatedConns {
			err := nbClient.UpdateExternalConnectionAPI(nb.UpdateExternalConnectionParams{
				Name: entry.name,
				EndpointInfo: &nb.EndpointInfo{
					Endpoint:     oldEndpoint,
					EndpointType: entry.endpointType,
				},
			})
			if err != nil {
				failedReverts = append(failedReverts, entry.name)
				log.Errorf("failed to revert connection %q: %s", entry.name, err)
			}
		}
		if len(failedReverts) > 0 {
			log.Errorf("MANUAL ACTION REQUIRED: the following connections could not be reverted and still point to the new endpoint: %v", failedReverts)
		}
	}

	if errs := setPauseAnnotation(allStores, false); len(errs) > 0 {
		log.Errorf("failed to remove pause annotations during rollback: %s", errs)
	}
}

func removePauseAnnotations(stores []matchedStore) []error {
	log := util.Logger()
	if errs := setPauseAnnotation(stores, false); len(errs) > 0 {
		log.Errorf("failed to remove pause annotations: %s", errs)
		return errs
	}
	return nil
}
