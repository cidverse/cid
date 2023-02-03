package version

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHighestReleaseType(t *testing.T) {
	tests := []struct {
		input    []ReleaseType
		expected ReleaseType
	}{
		{[]ReleaseType{ReleasePatch, ReleasePatch, ReleasePatch}, ReleasePatch},
		{[]ReleaseType{ReleaseMinor, ReleasePatch, ReleasePatch}, ReleaseMinor},
		{[]ReleaseType{ReleaseMinor, ReleaseMinor, ReleasePatch}, ReleaseMinor},
		{[]ReleaseType{ReleaseMajor, ReleaseMinor, ReleasePatch}, ReleaseMajor},
		{[]ReleaseType{ReleaseMajor, ReleaseMajor, ReleasePatch}, ReleaseMajor},
		{[]ReleaseType{ReleaseMajor, ReleaseMajor, ReleaseMajor}, ReleaseMajor},
	}

	for _, test := range tests {
		result := HighestReleaseType(test.input)
		if result != test.expected {
			t.Errorf("HighestReleaseType(%v) = %v, expected %v", test.input, result, test.expected)
		}
	}
}

func TestIsValidSemver(t *testing.T) {
	assert.True(t, IsValidSemver("v1.2.3"))
	assert.True(t, IsValidSemver("1.2.3"))
	assert.False(t, IsValidSemver("1.2.a"))
	assert.False(t, IsValidSemver("v1.2.3-"))
	assert.False(t, IsValidSemver("v1.2.3+"))
	assert.True(t, IsValidSemver("v0.0.0"))
	assert.True(t, IsValidSemver("v0.0.0-beta.1"))
	assert.True(t, IsValidSemver("v0.0.0-beta.1+123"))
}

func TestIsStable(t *testing.T) {
	assert.True(t, IsStable("v1.2.3"))
	assert.True(t, IsStable("v1.2.3+6123"))
	assert.False(t, IsStable("v1.2.3-rc.1"))
	assert.False(t, IsStable("v1.2.3-alpha.1+6123"))
	assert.False(t, IsStable("v1.2.3-beta.1"))
	assert.False(t, IsStable("v1.2.3-beta.1+6123"))
	assert.False(t, IsStable("v1.2.3-beta.1-6123"))
	assert.False(t, IsStable("v1.2.3-rc.1+6123"))
	assert.False(t, IsStable("v1.2.3-rc.1-6123"))
	assert.False(t, IsStable("v1.2.3-alpha.1"))
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      error
	}{
		{"1.2.3", "1.2.3", nil},
		{"1.2.3-alpha", "1.2.3-alpha", nil},
		{"1.2.3+build", "1.2.3+build", nil},
		{"1.2.3-alpha+build", "1.2.3-alpha+build", nil},
		{"invalid.format", "", fmt.Errorf(`malformed version: invalid.format`)},
		{"v1.2.3", "1.2.3", nil},
	}

	for _, test := range tests {
		got, err := Format(test.input)
		if got != test.expected {
			t.Errorf("Format(%q) = %q, want %q", test.input, got, test.expected)
		}
		if err != nil && err.Error() != test.err.Error() {
			t.Errorf("Format(%q) returned error %q, want %q", test.input, err, test.err)
		}
	}
}

func TestCompare(t *testing.T) {
	assert.Equal(t, 0, Compare("v1.2.3", "v1.2.3"))
	assert.Equal(t, -1, Compare("v1.2.3", "v2.0.0"))
	assert.Equal(t, 1, Compare("v1.6.0", "v1.2.3"))
	assert.Equal(t, 0, Compare("v1.2.3+6123", "v1.2.3+6123"))
	assert.Equal(t, 0, Compare("v1.2.3-rc.1", "v1.2.3-rc.1"))
	assert.Equal(t, 0, Compare("v0.0.0", "v0.0.0"))
	assert.Equal(t, -1, Compare("v0.0.0", "v0.0.1"))
	assert.Equal(t, 1, Compare("v0.0.1", "v0.0.0"))
}

func TestFulfillsConstraint(t *testing.T) {
	assert.True(t, FulfillsConstraint("v1.0.0", ">=0.0.0"))
	assert.True(t, FulfillsConstraint("v1.2.3", ">=1.2.3"))
	assert.False(t, FulfillsConstraint("v1.2.3", ">=1.2.4"))
	assert.True(t, FulfillsConstraint("v1.2.3", ">1.2.2"))
	assert.False(t, FulfillsConstraint("v1.2.3", ">1.2.3"))
	assert.True(t, FulfillsConstraint("v1.2.3", "<=1.2.3"))
	assert.False(t, FulfillsConstraint("v1.2.3", "<=1.2.2"))
	assert.True(t, FulfillsConstraint("v1.2.3", "<1.2.4"))
	assert.False(t, FulfillsConstraint("v1.2.3", "<1.2.3"))
	assert.True(t, FulfillsConstraint("v1.2.3", "1.2.3"))
	assert.False(t, FulfillsConstraint("v1.2.3", "1.2.2"))
	assert.False(t, FulfillsConstraint("v1.2.3", "1.2.4"))
}

func TestBumpMajor(t *testing.T) {
	bumped, err := Bump("1.2.3", ReleaseMajor)
	assert.NoError(t, err)
	assert.Equal(t, "2.0.0", bumped)
}

func TestBumpMinor(t *testing.T) {
	bumped, err := Bump("1.2.3", ReleaseMinor)
	assert.NoError(t, err)
	assert.Equal(t, "1.3.0", bumped)
}

func TestBumpPatch(t *testing.T) {
	bumped, err := Bump("1.2.3", ReleasePatch)
	assert.NoError(t, err)
	assert.Equal(t, "1.2.4", bumped)
}
