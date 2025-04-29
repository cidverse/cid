package cargotoml

import (
	"bytes"
	"testing"
)

func TestPatchVersion(t *testing.T) {
	original := []byte(`[package]
name = "my-crate"
version = "0.1.0"
description = "A sample crate"

[dependencies]
serde = "1.0"
`)

	expected := []byte(`[package]
name = "my-crate"
version = "1.2.3"
description = "A sample crate"

[dependencies]
serde = "1.0"
`)

	result, err := PatchVersion(original, "1.2.3")
	if err != nil {
		t.Fatalf("PatchVersion returned error: %v", err)
	}

	if !bytes.Equal(result, expected) {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, result)
	}
}

func TestPatchVersion_NoPackageSection(t *testing.T) {
	invalid := []byte(`[dependencies]
serde = "1.0"
`)

	_, err := PatchVersion(invalid, "2.0.0")
	if err == nil {
		t.Errorf("Expected error when no [package] section is present")
	}
}
