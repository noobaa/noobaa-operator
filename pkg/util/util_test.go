package util

import (
	"testing"
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
