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
