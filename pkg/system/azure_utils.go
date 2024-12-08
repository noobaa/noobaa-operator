package system

import (
	"fmt"
	"log"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

func (r *Reconciler) getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(r.AzureContainerCreds.StringData["azure_subscription_id"])
	auth, _ := r.GetResourceManagementAuthorizer()
	storageAccountsClient.Authorizer = auth
	err := storageAccountsClient.AddToUserAgent("Go-http-client/1.1")
	if err != nil {
		log.Fatalf("got error on storageAccountsClient.AddToUserAgent %v", err)
	}
	return storageAccountsClient
}

func (r *Reconciler) getAccountPrimaryKey(accountName, accountGroupName string) string {
	response, err := r.GetAccountKeys(accountName, accountGroupName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

// CreateStorageAccount starts creation of a new storage account and waits for
// the account to be created.
func (r *Reconciler) CreateStorageAccount(accountName, accountGroupName string) (storage.Account, error) {
	var s storage.Account
	storageAccountsClient := r.getStorageAccountsClient()

	// we used to call storage.AccountCheckNameAvailabilityParameters here to make sure the name is available
	// removed it because when using a newer API version (2019-06-01), this call produced some irrelevant errors sometimes
	// if the name is not available, CreateStorageAccount will return an error, and a different name will be used next time

	enableHTTPSTrafficOnly := true
	allowBlobPublicAccess := false
	future, err := storageAccountsClient.Create(
		r.Ctx,
		accountGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Kind:     storage.StorageV2,
			Location: to.StringPtr(r.AzureContainerCreds.StringData["azure_region"]),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{
				EnableHTTPSTrafficOnly: &enableHTTPSTrafficOnly,
				AllowBlobPublicAccess:  &allowBlobPublicAccess,
				MinimumTLSVersion:      storage.TLS12,
			},
		})

	if err != nil {
		return s, fmt.Errorf("failed to start creating storage account: %+v", err)
	}

	err = future.WaitForCompletionRef(r.Ctx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("failed to finish creating storage account: %+v", err)
	}

	return future.Result(storageAccountsClient)
}

// GetStorageAccount gets details on the specified storage account
func (r *Reconciler) GetStorageAccount(accountName, accountGroupName string) (storage.Account, error) {
	storageAccountsClient := r.getStorageAccountsClient()
	return storageAccountsClient.GetProperties(r.Ctx, accountGroupName, accountName, storage.AccountExpandBlobRestoreStatus)
}

// DeleteStorageAccount deletes an existing storate account
func (r *Reconciler) DeleteStorageAccount(accountName, accountGroupName string) (autorest.Response, error) {
	storageAccountsClient := r.getStorageAccountsClient()
	return storageAccountsClient.Delete(r.Ctx, accountGroupName, accountName)
}

// CheckAccountNameAvailability checks if the storage account name is available.
// Storage account names must be unique across Azure and meet other requirements.
func (r *Reconciler) CheckAccountNameAvailability(accountName string) (storage.CheckNameAvailabilityResult, error) {
	storageAccountsClient := r.getStorageAccountsClient()
	result, err := storageAccountsClient.CheckNameAvailability(
		r.Ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	return result, err
}

// GetAccountKeys gets the storage account keys
func (r *Reconciler) GetAccountKeys(accountName, accountGroupName string) (storage.AccountListKeysResult, error) {
	accountsClient := r.getStorageAccountsClient()
	return accountsClient.ListKeys(r.Ctx, accountGroupName, accountName, storage.Kerb)
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

// Environment returns an `azure.Environment{...}` for the current cloud.
func (r *Reconciler) Environment() *azure.Environment {
	env, _ := azure.EnvironmentFromName("AzurePublicCloud")
	return &env
}

// GetResourceManagementAuthorizer gets an OAuthTokenAuthorizer for Azure Resource Manager
func (r *Reconciler) GetResourceManagementAuthorizer() (autorest.Authorizer, error) {
	return r.getAuthorizerForResource(r.Environment().ResourceManagerEndpoint)
}

func (r *Reconciler) getAuthorizerForResource(resource string) (autorest.Authorizer, error) {
	var a autorest.Authorizer
	var err error

	oauthConfig, err := adal.NewOAuthConfig(
		r.Environment().ActiveDirectoryEndpoint, r.AzureContainerCreds.StringData["azure_tenant_id"])
	if err != nil {
		return nil, err
	}

	token, err := adal.NewServicePrincipalToken(
		*oauthConfig, r.AzureContainerCreds.StringData["azure_client_id"], r.AzureContainerCreds.StringData["azure_client_secret"], resource)
	if err != nil {
		return nil, err
	}
	a = autorest.NewBearerAuthorizer(token)

	return a, err
}
