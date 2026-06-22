package util

import (
	"fmt"
	"strings"
	"testing"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func TestGoogleCredentialsFromStoreSecret(t *testing.T) {
	validExternalAccountJSON := fmt.Sprintf(
		`{"type":"external_account","service_account_impersonation_url":%q}`,
		googleImpersonationURL(googleWIFServiceAccountEmail),
	)
	validServiceAccountJSON := fmt.Sprintf(`{"type":"service_account","private_key_id":%q}`, googleServiceAccountKeyID)

	tests := []struct {
		name          string
		secret        *corev1.Secret
		expectedJSON  string
		expectedErr   bool
	}{
		{
			name: "only WIF key",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "gcp-wif"},
				StringData: map[string]string{GoogleCredentialsJson: validExternalAccountJSON},
			},
			expectedJSON: validExternalAccountJSON,
		},
		{
			name: "only classic key",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "gcp-classic"},
				StringData: map[string]string{GoogleServiceAccountPrivateKeyJson: validServiceAccountJSON},
			},
			expectedJSON: validServiceAccountJSON,
		},
		{
			name: "both keys",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "gcp-both"},
				StringData: map[string]string{
					GoogleCredentialsJson:              validExternalAccountJSON,
					GoogleServiceAccountPrivateKeyJson: validServiceAccountJSON,
				},
			},
			expectedErr: true,
		},
		{
			name: "CCO service_account.json with WIF key",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "noobaa-gcp-bucket-creds", Namespace: "openshift-storage"},
				StringData: map[string]string{
					"service_account.json":  validExternalAccountJSON,
					GoogleCredentialsJson: validExternalAccountJSON,
				},
			},
			expectedJSON: validExternalAccountJSON,
		},
		{
			name: "CCO service_account.json with classic key",
			secret: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{Name: "noobaa-gcp-bucket-creds", Namespace: "openshift-storage"},
				StringData: map[string]string{
					"service_account.json":             validServiceAccountJSON,
					GoogleServiceAccountPrivateKeyJson: validServiceAccountJSON,
				},
			},
			expectedJSON: validServiceAccountJSON,
		},
		{
			name:        "nil secret",
			secret:      nil,
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotJSON, err := GoogleCredentialsFromStoreSecret(tc.secret)
			if tc.expectedErr {
				if err == nil {
					t.Fatal("GoogleCredentialsFromStoreSecret expected error")
				}
				if !IsValidationError(err) {
					t.Fatalf("GoogleCredentialsFromStoreSecret expected ValidationError, got %T", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("GoogleCredentialsFromStoreSecret unexpected error: %v", err)
			}
			if gotJSON != tc.expectedJSON {
				t.Fatalf("expected JSON %q, got %q", tc.expectedJSON, gotJSON)
			}
		})
	}
}

func TestValidateGoogleCredentialsJSONType(t *testing.T) {
	validExternalAccountJSON := fmt.Sprintf(
		`{"type":"external_account","service_account_impersonation_url":%q}`,
		googleImpersonationURL(googleWIFServiceAccountEmail),
	)
	validServiceAccountJSON := fmt.Sprintf(`{"type":"service_account","private_key_id":%q}`, googleServiceAccountKeyID)

	tests := []struct {
		name        string
		json        string
		expectSTS   bool
		expectedErr bool
		errContains string
	}{
		{
			name:      "service_account for GCP long-lived",
			json:      validServiceAccountJSON,
			expectSTS: false,
		},
		{
			name:      "external_account for GCP WIF (STS)",
			json:      validExternalAccountJSON,
			expectSTS: true,
		},
		{
			name:        "GCP WIF (STS) secret used with GCP long-lived command",
			json:        validExternalAccountJSON,
			expectSTS:   false,
			expectedErr: true,
			errContains: "service_account",
		},
		{
			name:        "GCP long-lived secret used with GCP WIF (STS) command",
			json:        validServiceAccountJSON,
			expectSTS:   true,
			expectedErr: true,
			errContains: "external_account",
		},
		{
			name:        "invalid JSON",
			json:        `{`,
			expectSTS:   false,
			expectedErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGoogleCredentialsJSONType(tt.json, tt.expectSTS)
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %v", tt.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
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

func TestGoogleWIFAudience(t *testing.T) {
	got := GoogleWIFAudience("123456789", "my-pool", "my-provider")
	want := "//iam.googleapis.com/projects/123456789/locations/global/workloadIdentityPools/my-pool/providers/my-provider"
	if got != want {
		t.Fatalf("expected audience %q, got %q", want, got)
	}
}

func TestBuildGoogleWIFCredentialsJSON(t *testing.T) {
	const (
		projectNumber       = "123456789"
		poolID              = "my-pool"
		providerID          = "my-provider"
		serviceAccountEmail = "noobaa-wif-sa@my-project.iam.gserviceaccount.com"
	)

	t.Run("valid parameters", func(t *testing.T) {
		credentialsJSON, err := BuildGoogleWIFCredentialsJSON(
			projectNumber, poolID, providerID, serviceAccountEmail,
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		isExternalAccount, identity, err := ParseGoogleCredentials(credentialsJSON)
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if !isExternalAccount {
			t.Fatal("expected external_account credentials")
		}
		if identity != serviceAccountEmail {
			t.Fatalf("expected identity %q, got %q", serviceAccountEmail, identity)
		}
		if !strings.Contains(credentialsJSON, GoogleWIFAudience(projectNumber, poolID, providerID)) {
			t.Fatalf("expected audience in credentials json: %s", credentialsJSON)
		}
		if !strings.Contains(credentialsJSON, WebIdentityTokenPath) {
			t.Fatalf("expected token path in credentials json: %s", credentialsJSON)
		}
	})

	t.Run("ignores surrounding whitespace on parameters", func(t *testing.T) {
		credentialsJSON, err := BuildGoogleWIFCredentialsJSON(
			" "+projectNumber+" ",
			" "+poolID+" ",
			" "+providerID+" ",
			" "+serviceAccountEmail+" ",
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expectedAudience := GoogleWIFAudience(projectNumber, poolID, providerID)
		if !strings.Contains(credentialsJSON, expectedAudience) {
			t.Fatalf("expected audience %q in json, got: %s", expectedAudience, credentialsJSON)
		}
		if strings.Contains(credentialsJSON, " "+projectNumber+" ") {
			t.Fatalf("json must not contain untrimmed project number with spaces")
		}

		_, identity, err := ParseGoogleCredentials(credentialsJSON)
		if err != nil {
			t.Fatalf("unexpected parse error: %v", err)
		}
		if identity != serviceAccountEmail {
			t.Fatalf("expected trimmed identity %q, got %q", serviceAccountEmail, identity)
		}
	})

	t.Run("missing required parameters", func(t *testing.T) {
		_, err := BuildGoogleWIFCredentialsJSON("", "", "", "")
		if err == nil {
			t.Fatal("expected error when required params are missing")
		}
	})
}

func TestGcpProjectIDFromServiceAccountEmail(t *testing.T) {
	tests := []struct {
		name              string
		email             string
		expectedProjectID string
		expectedErr       bool
	}{
		{
			name:              "valid GCP service account email",
			email:             "noobaa-wif-sa@my-project.iam.gserviceaccount.com",
			expectedProjectID: "my-project",
		},
		{
			name:        "missing at sign",
			email:       "invalid-email",
			expectedErr: true,
		},
		// GcpProjectIDFromServiceAccountEmail only strips .iam.gserviceaccount.com;
		// other domains pass through the part after @ unchanged (no format validation).
		{
			name:              "non-GCP domain",
			email:             "sa@other-domain.com",
			expectedProjectID: "other-domain.com",
		},
		{
			name:        "empty domain after at sign",
			email:       "sa@",
			expectedErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GcpProjectIDFromServiceAccountEmail(tc.email)
			if tc.expectedErr {
				if err == nil {
					t.Fatal("GcpProjectIDFromServiceAccountEmail expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("GcpProjectIDFromServiceAccountEmail unexpected error: %v", err)
			}
			if got != tc.expectedProjectID {
				t.Fatalf("GcpProjectIDFromServiceAccountEmail expected project ID %q, got %q", tc.expectedProjectID, got)
			}
		})
	}
}

func TestValidateGCPWIFParams(t *testing.T) {
	valid := struct {
		projectNumber       string
		poolID              string
		providerID          string
		serviceAccountEmail string
	}{
		projectNumber:       "123456789",
		poolID:              "my-pool",
		providerID:          "my-provider",
		serviceAccountEmail: "noobaa-wif-sa@my-project.iam.gserviceaccount.com",
	}

	tests := []struct {
		name                string
		projectNumber       string
		poolID              string
		providerID          string
		serviceAccountEmail string
		expectedErr         bool
		errContains         string
	}{
		{
			name: "none set",
		},
		{
			name:                "all four valid",
			projectNumber:       valid.projectNumber,
			poolID:              valid.poolID,
			providerID:          valid.providerID,
			serviceAccountEmail: valid.serviceAccountEmail,
		},
		{
			name:          "only project-number",
			projectNumber: valid.projectNumber,
			expectedErr:   true,
			errContains:   "all four are required",
		},
		{
			name:          "three of four",
			projectNumber: valid.projectNumber,
			poolID:        valid.poolID,
			providerID:    valid.providerID,
			expectedErr:   true,
			errContains:   "all four are required",
		},
		{
			name:                "non-numeric project number",
			projectNumber:       "my-project",
			poolID:              valid.poolID,
			providerID:          valid.providerID,
			serviceAccountEmail: valid.serviceAccountEmail,
			expectedErr:         true,
			errContains:         "numeric GCP project number",
		},
		{
			name:                "pool-id too short",
			projectNumber:       valid.projectNumber,
			poolID:              "abc",
			providerID:          valid.providerID,
			serviceAccountEmail: valid.serviceAccountEmail,
			expectedErr:         true,
			errContains:         "pool-id must be 4-32",
		},
		{
			name:                "pool-id reserved gcp- prefix",
			projectNumber:       valid.projectNumber,
			poolID:              "gcp-pool",
			providerID:          valid.providerID,
			serviceAccountEmail: valid.serviceAccountEmail,
			expectedErr:         true,
			errContains:         "gcp-",
		},
		{
			name:                "provider-id invalid characters",
			projectNumber:       valid.projectNumber,
			poolID:              valid.poolID,
			providerID:          "My_Provider",
			serviceAccountEmail: valid.serviceAccountEmail,
			expectedErr:         true,
			errContains:         "provider-id must contain only lowercase letters, digits, and hyphens",
		},
		{
			name:                "invalid service account email suffix",
			projectNumber:       valid.projectNumber,
			poolID:              valid.poolID,
			providerID:          valid.providerID,
			serviceAccountEmail: "noobaa-wif-sa@my-project.example.com",
			expectedErr:         true,
			errContains:         ".iam.gserviceaccount.com",
		},
		{
			name:                "service account email missing project ID",
			projectNumber:       valid.projectNumber,
			poolID:              valid.poolID,
			providerID:          valid.providerID,
			serviceAccountEmail: "noobaa-wif-sa@.iam.gserviceaccount.com",
			expectedErr:         true,
			errContains:         "missing project ID",
		},
		{
			name:                "service account email extra at-sign in domain",
			projectNumber:       valid.projectNumber,
			poolID:              valid.poolID,
			providerID:          valid.providerID,
			serviceAccountEmail: "noobaa-wif-sa@my@project.iam.gserviceaccount.com",
			expectedErr:         true,
			errContains:         "exactly one @",
		},
		{
			name:                "whitespace-only counts as set — all four required",
			projectNumber:       " ",
			poolID:              " ",
			providerID:          " ",
			serviceAccountEmail: " ",
			expectedErr:         true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateGCPWIFParams(tt.projectNumber, tt.poolID, tt.providerID, tt.serviceAccountEmail)
			if tt.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Fatalf("expected error containing %q, got %v", tt.errContains, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
