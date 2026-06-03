package util

import (
	"fmt"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
)

const (
	googleWIFServiceAccountEmail = "noobaa-wif-sa@my-project.iam.gserviceaccount.com"
	googleServiceAccountKeyID    = "key-id-123"
)

// TestMapAlternateKeysValue a tiny test for util.MapAlternateKeysValue()
func TestMapAlternateKeysValue(t *testing.T) {
	keyVal := "XXXXX"
	altKey := "aws_access_key_id"
	canonicalKey := "AWS_ACCESS_KEY_ID"

	actualAltValue := MapAlternateKeysValue(map[string]string{
		altKey: keyVal,
	}, canonicalKey)
	if actualAltValue != keyVal {
		t.Fatalf("Alternative key: expected %q, got %q", keyVal, actualAltValue)
	}

	actualCanonicalValue := MapAlternateKeysValue(map[string]string{
		canonicalKey: keyVal,
	}, canonicalKey)
	if actualCanonicalValue != keyVal {
		t.Fatalf("Canonical key: expected %q, got %q", keyVal, actualCanonicalValue)
	}
}

func TestIsRemoteObcAnnotation(t *testing.T) {
	tests := []struct {
		name        string
		annotations map[string]string
		expected    bool
	}{
		{
			name:        "nil annotations",
			annotations: nil,
			expected:    false,
		},
		{
			name:        "missing annotation",
			annotations: map[string]string{"other": "true"},
			expected:    false,
		},
		{
			name:        "explicit false",
			annotations: map[string]string{"remote-obc-creation": "false"},
			expected:    false,
		},
		{
			name:        "explicit true",
			annotations: map[string]string{"remote-obc-creation": "true"},
			expected:    true,
		},
		{
			name:        "mixed case true",
			annotations: map[string]string{"remote-obc-creation": "TrUe"},
			expected:    true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			actual := IsRemoteObcAnnotation(testCase.annotations)
			if actual != testCase.expected {
				t.Fatalf("expected %v, got %v", testCase.expected, actual)
			}
		})
	}
}

// TestIsAzureSTSClusterBS verifies Azure STS backing store detection for backingstore
func TestIsAzureSTSClusterBS(t *testing.T) {
	clientID := "test-client-id"
	tenantID := "test-tenant-id"

	tests := []struct {
		name         string
		backingStore *nbv1.BackingStore
		expected     bool
	}{
		{
			name: "azure-blob with ClientId returns true",
			backingStore: &nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						ClientId:            &clientID,
						TenantId:            &tenantID,
					},
				},
			},
			expected: true,
		},
		{
			name: "azure-blob without ClientId returns false",
			backingStore: &nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
					},
				},
			},
			expected: false,
		},
		{
			name: "aws-s3 type returns false",
			backingStore: &nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAWSS3,
					AWSS3: &nbv1.AWSS3Spec{
						TargetBucket: "bucket",
					},
				},
			},
			expected: false,
		},
		{
			name: "nil AzureBlob spec returns false",
			backingStore: &nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
				},
			},
			expected: false,
		},
		{
			name: "azure-blob with ClientId, TenantId, SubscriptionId and ResourcegroupId returns true",
			backingStore: &nbv1.BackingStore{
				Spec: nbv1.BackingStoreSpec{
					Type: nbv1.StoreTypeAzureBlob,
					AzureBlob: &nbv1.AzureBlobSpec{
						TargetBlobContainer: "container",
						ClientId:            &clientID,
						TenantId:            &tenantID,
						SubscriptionId:      ptrString("sub-id"),
						ResourcegroupId:     ptrString("rg-id"),
					},
				},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := IsAzureSTSClusterBS(tc.backingStore)
			if actual != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}

func ptrString(s string) *string {
	return &s
}

// googleImpersonationURL returns the impersonation URL for the given email
func googleImpersonationURL(email string) string {
	return GoogleImpersonationURLPrefix + email + GoogleImpersonationURLSuffix
}

// TestParseGoogleCredentials tests ParseGoogleCredentials function
func TestParseGoogleCredentials(t *testing.T) {
	validExternalAccountJSON := fmt.Sprintf(
		`{"type":"external_account","service_account_impersonation_url":%q}`,
		googleImpersonationURL(googleWIFServiceAccountEmail),
	)

	tests := []struct {
		name                    string
		json                    string
		expectedExternalAccount bool
		expectedIdentity        string
		expectedErr             bool
	}{
		{
			name:                    "external_account",
			json:                    validExternalAccountJSON,
			expectedExternalAccount: true,
			expectedIdentity:        googleWIFServiceAccountEmail,
		},
		{
			name:                    "service_account",
			json:                    fmt.Sprintf(`{"type":"service_account","private_key_id":%q}`, googleServiceAccountKeyID),
			expectedExternalAccount: false,
			expectedIdentity:        googleServiceAccountKeyID,
		},
		{
			name:        "invalid JSON",
			json:        `{`,
			expectedErr: true,
		},
		{
			name:        "propagates identity error",
			json:        `{"type":"service_account"}`,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isExternalAccount, identity, err := ParseGoogleCredentials(tc.json)
			if tc.expectedErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if isExternalAccount != tc.expectedExternalAccount {
				t.Fatalf("expected isExternalAccount=%v, got %v", tc.expectedExternalAccount, isExternalAccount)
			}
			if identity != tc.expectedIdentity {
				t.Fatalf("expected identity %q, got %q", tc.expectedIdentity, identity)
			}
		})
	}
}

// TestGoogleIdentityFromCredentials tests googleIdentityFromCredentials function
func TestGoogleIdentityFromCredentials(t *testing.T) {
	tests := []struct {
		name             string
		creds            *googleCredentialsJSON
		expectedIdentity string
		expectedErr      bool
	}{
		{
			name: "external_account",
			creds: &googleCredentialsJSON{
				Type:                           "external_account",
				ServiceAccountImpersonationURL: googleImpersonationURL(googleWIFServiceAccountEmail),
			},
			expectedIdentity: googleWIFServiceAccountEmail,
		},
		{
			name: "service_account",
			creds: &googleCredentialsJSON{
				Type:         "service_account",
				PrivateKeyID: googleServiceAccountKeyID,
			},
			expectedIdentity: googleServiceAccountKeyID,
		},
		{
			name: "service_account missing private_key_id",
			creds: &googleCredentialsJSON{
				Type: "service_account",
			},
			expectedErr: true,
		},
		{
			name: "missing type",
			creds: &googleCredentialsJSON{
				PrivateKeyID: googleServiceAccountKeyID,
			},
			expectedErr: true,
		},
		{
			name: "unsupported type",
			creds: &googleCredentialsJSON{
				Type: "authorized_user",
			},
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			identity, err := googleIdentityFromCredentials(tc.creds)
			if tc.expectedErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if identity != tc.expectedIdentity {
				t.Fatalf("expected identity %q, got %q", tc.expectedIdentity, identity)
			}
		})
	}
}

// TestGoogleIdentityFromImpersonationURL tests googleIdentityFromImpersonationURL function
func TestGoogleIdentityFromImpersonationURL(t *testing.T) {
	tests := []struct {
		name             string
		impersonationURL string
		expectedIdentity string
		expectedErr      bool
	}{
		{
			name:             "valid url",
			impersonationURL: googleImpersonationURL(googleWIFServiceAccountEmail),
			expectedIdentity: googleWIFServiceAccountEmail,
		},
		{
			name:             "invalid prefix or suffix",
			impersonationURL: "https://example.com",
			expectedErr:      true,
		},
		{
			name:             "empty url",
			impersonationURL: "",
			expectedErr:      true,
		},
		{
			name:             "empty email",
			impersonationURL: GoogleImpersonationURLPrefix + GoogleImpersonationURLSuffix,
			expectedErr:      true,
		},
		{
			name:             "malformed email",
			impersonationURL: googleImpersonationURL("not-an-email"),
			expectedErr:      true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			identity, err := googleIdentityFromImpersonationURL(tc.impersonationURL)
			if tc.expectedErr {
				if err == nil {
					t.Fatalf("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if identity != tc.expectedIdentity {
				t.Fatalf("expected identity %q, got %q", tc.expectedIdentity, identity)
			}
		})
	}
}
