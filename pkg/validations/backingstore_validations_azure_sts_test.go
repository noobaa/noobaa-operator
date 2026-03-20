package validations

import (
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

// TestValidateAzureSTSCredentials verifies Azure STS credentials validation for backingstore.
// When secret name is empty, clientId and tenantId must be provided for Azure STS (web identity).
func TestValidateAzureSTSCredentials(t *testing.T) {
	clientID := "azure-client-id"
	tenantID := "azure-tenant-id"
	subscriptionID := "azure-subscription-id"
	resourceGroupID := "azure-resource-group"

	tests := []struct {
		name        string
		backingStore nbv1.BackingStore
		wantErr     bool
		errMsg      string
	}{
		{
			name: "allow when clientId and tenantId are set",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						Secret:             corev1.SecretReference{Name: "", Namespace: "test"},
						ClientId:            &clientID,
						TenantId:            &tenantID,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "allow when clientId, tenantId, subscriptionId and resourcegroupId are set",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						Secret:              corev1.SecretReference{Name: "", Namespace: "test"},
						ClientId:            &clientID,
						TenantId:            &tenantID,
						SubscriptionId:     &subscriptionID,
						ResourcegroupId:    &resourceGroupID,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "deny when secret is empty and clientId is nil",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						Secret:              corev1.SecretReference{Name: "", Namespace: "test"},
						TenantId:            &tenantID,
					},
				},
			},
			wantErr: true,
			errMsg:  "please provide secret name or Azure STS credentials (clientId and tenantId)",
		},
		{
			name: "deny when secret is empty and tenantId is nil",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						Secret:              corev1.SecretReference{Name: "", Namespace: "test"},
						ClientId:            &clientID,
					},
				},
			},
			wantErr: true,
			errMsg:  "please provide secret name or Azure STS credentials (clientId and tenantId)",
		},
		{
			name: "deny when secret is empty and both clientId and tenantId are nil",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						Secret:              corev1.SecretReference{Name: "", Namespace: "test"},
					},
				},
			},
			wantErr: true,
			errMsg:  "please provide secret name or Azure STS credentials (clientId and tenantId)",
		},
		{
			name: "nil AzureBlob spec - ValidateAzureSTSCredentials does not error",
			backingStore: nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: nil,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAzureSTSCredentials(tt.backingStore)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAzureSTSCredentials() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" && (err == nil || !strings.Contains(err.Error(), tt.errMsg)) {
				t.Errorf("ValidateAzureSTSCredentials() error = %v, want message containing %q", err, tt.errMsg)
			}
		})
	}
}
