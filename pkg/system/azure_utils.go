package system

import (
	"fmt"
	"log"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"k8s.io/utils/ptr"
)

func (r *Reconciler) getStorageAccountsClient() (*armstorage.AccountsClient, error) {
	if r.IsAzureSTSCluster {
		workloadIdentitycred, err := r.getAuthorizerForWorkloadIdentity()
		if err != nil {
			return nil, err
		}
		return armstorage.NewAccountsClient(
			r.AzureContainerCreds.StringData["azure_subscription_id"],
			workloadIdentitycred,
			&arm.ClientOptions{},
		)
	} else {
		clientSecretCrede, err := r.getAuthorizerForSecretCredential()
		if err != nil {
			return nil, err
		}
		return armstorage.NewAccountsClient(
			r.AzureContainerCreds.StringData["azure_subscription_id"],
			clientSecretCrede,
			&arm.ClientOptions{},
		)
	}
}

func (r *Reconciler) getAccountPrimaryKey(accountName, accountGroupName string) string {
	response, err := r.GetAccountKeys(accountName, accountGroupName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	if len(response.Keys) == 0 || response.Keys[0].Value == nil {
		log.Fatalf("no storage account keys returned for accountName:  %s", accountName)
	}
	return *response.Keys[0].Value
}

// Helper function to create pointers to values as required by Go SDK
func toPtr[T any](v T) *T {
	return &v
}

// CreateStorageAccount starts creation of a new storage account and waits for
// the account to be created.
func (r *Reconciler) CreateStorageAccount(accountName, accountGroupName string) (armstorage.AccountsClientCreateResponse, error) {
	var storageAccount armstorage.AccountsClientCreateResponse
	accountsClient, err := r.getStorageAccountsClient()
	if err != nil {
		return storageAccount, err
	}
	// we used to call storage.AccountCheckNameAvailabilityParameters here to make sure the name is available
	// removed it because when using a newer API version (2019-06-01), this call produced some irrelevant errors sometimes
	// if the name is not available, CreateStorageAccount will return an error, and a different name will be used next time
	enableHTTPSTrafficOnly := true
	allowBlobPublicAccess := false
	future, err := accountsClient.BeginCreate(
		r.Ctx,
		accountGroupName,
		accountName,
		armstorage.AccountCreateParameters{
			SKU: &armstorage.SKU{
				Name: toPtr(armstorage.SKUNameStandardLRS),
			},
			Kind:     toPtr(armstorage.KindStorageV2),
			Location: ptr.To(r.AzureContainerCreds.StringData["azure_region"]),
			Properties: &armstorage.AccountPropertiesCreateParameters{
				EnableHTTPSTrafficOnly: &enableHTTPSTrafficOnly,
				AllowBlobPublicAccess:  &allowBlobPublicAccess,
				MinimumTLSVersion:      toPtr(armstorage.MinimumTLSVersionTLS12),
			},
		},
		&armstorage.AccountsClientBeginCreateOptions{},
	)

	if err != nil {
		return storageAccount, fmt.Errorf("failed to start creating storage account: %+v", err)
	}

	storageAccount, err = future.PollUntilDone(r.Ctx, &runtime.PollUntilDoneOptions{})
	if err != nil {
		return storageAccount, fmt.Errorf("failed to finish creating storage account: %+v", err)
	}

	return storageAccount, err
}

// GetStorageAccount gets details on the specified storage account
func (r *Reconciler) GetStorageAccount(accountName, accountGroupName string) (armstorage.AccountsClientGetPropertiesResponse, error) {
	storageAccountsClient, err := r.getStorageAccountsClient()
	if err != nil {
		return armstorage.AccountsClientGetPropertiesResponse{}, err
	}
	return storageAccountsClient.GetProperties(r.Ctx, accountGroupName, accountName, &armstorage.AccountsClientGetPropertiesOptions{})
}

// DeleteStorageAccount deletes an existing storate account
func (r *Reconciler) DeleteStorageAccount(accountName, accountGroupName string) (armstorage.AccountsClientDeleteResponse, error) {
	storageAccountsClient, err := r.getStorageAccountsClient()
	if err != nil {
		return armstorage.AccountsClientDeleteResponse{}, err
	}
	return storageAccountsClient.Delete(r.Ctx, accountGroupName, accountName, &armstorage.AccountsClientDeleteOptions{})
}

// CheckAccountNameAvailability checks if the storage account name is available.
// Storage account names must be unique across Azure and meet other requirements.
func (r *Reconciler) CheckAccountNameAvailability(accountName string) (armstorage.AccountsClientCheckNameAvailabilityResponse, error) {
	storageAccountsClient, err := r.getStorageAccountsClient()
	if err != nil {
		return armstorage.AccountsClientCheckNameAvailabilityResponse{}, err
	}
	paramaccountName := armstorage.AccountCheckNameAvailabilityParameters{
		Name: ptr.To(accountName),
		Type: ptr.To("Microsoft.Storage/storageAccounts"),
	}

	result, err := storageAccountsClient.CheckNameAvailability(
		r.Ctx,
		paramaccountName,
		&armstorage.AccountsClientCheckNameAvailabilityOptions{})
	return result, err
}

// GetAccountKeys gets the storage account keys
func (r *Reconciler) GetAccountKeys(accountName, accountGroupName string) (armstorage.AccountsClientListKeysResponse, error) {
	accountsClient, err := r.getStorageAccountsClient()
	if err != nil {
		return armstorage.AccountsClientListKeysResponse{}, err
	}
	return accountsClient.ListKeys(r.Ctx, accountGroupName, accountName, &armstorage.AccountsClientListKeysOptions{})
}

func (r *Reconciler) getContainerURL(accountName, accountGroupName, containerName string) azblob.ContainerURL {
	key := r.getAccountPrimaryKey(accountName, accountGroupName)
	c, _ := azblob.NewSharedKeyCredential(accountName, key)
	p := azblob.NewPipeline(c, azblob.PipelineOptions{
		Telemetry: azblob.TelemetryOptions{Value: "Go-http-client/1.1"},
	})
	u, _ := url.Parse(fmt.Sprintf(`https://%s.blob.core.windows.net`, accountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

// CreateContainer creates a new container with the specified name in the specified account
func (r *Reconciler) CreateContainer(accountName, accountGroupName, containerName string) (azblob.ContainerURL, error) {
	c := r.getContainerURL(accountName, accountGroupName, containerName)

	_, err := c.Create(
		r.Ctx,
		azblob.Metadata{},
		azblob.PublicAccessNone)
	return c, err
}

// GetContainer gets info about an existing container.
func (r *Reconciler) GetContainer(accountName, accountGroupName, containerName string) (azblob.ContainerURL, error) {
	c := r.getContainerURL(accountName, accountGroupName, containerName)

	_, err := c.GetProperties(r.Ctx, azblob.LeaseAccessConditions{})
	return c, err
}

// DeleteContainer deletes the named container.
func (r *Reconciler) DeleteContainer(accountName, accountGroupName, containerName string) error {
	c := r.getContainerURL(accountName, accountGroupName, containerName)

	_, err := c.Delete(r.Ctx, azblob.ContainerAccessConditions{})
	return err
}

// Get authorizer for the WorkloadIdentity Credential,
// It will use the authorizer using federated we identity token
func (r *Reconciler) getAuthorizerForWorkloadIdentity() (*azidentity.WorkloadIdentityCredential, error) {

	// Workload Identity Federation
	workloadIdentityCredential, err := azidentity.NewWorkloadIdentityCredential(&azidentity.WorkloadIdentityCredentialOptions{
		ClientID:      r.AzureContainerCreds.StringData["azure_client_id"],
		TenantID:      r.AzureContainerCreds.StringData["azure_tenant_id"],
		TokenFilePath: r.webIdentityTokenPath,
		ClientOptions: policy.ClientOptions{},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create WorkloadIdentityCredential: %+v", err)
	}
	return workloadIdentityCredential, nil
}

// Get authorizer for the normal Credential,
// It will use the authorizer using client id and secret
func (r *Reconciler) getAuthorizerForSecretCredential() (*azidentity.ClientSecretCredential, error) {
	// Service Principal with Secret
	clientSecret := r.AzureContainerCreds.StringData["azure_client_secret"]
	clientSecretCredential, err := azidentity.NewClientSecretCredential(
		r.AzureContainerCreds.StringData["azure_tenant_id"],
		r.AzureContainerCreds.StringData["azure_client_id"],
		clientSecret,
		&azidentity.ClientSecretCredentialOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Credential: %+v", err)
	}
	return clientSecretCredential, nil

}
