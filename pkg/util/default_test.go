package util

import (
	"testing"
)

func TestGetOrDefault(t *testing.T) {
	testCases := []struct {
		value        string
		defaultValue string
		expected     string
	}{
		{"foo", "bar", "foo"},
		{"", "bar", "bar"},
		{"", "", ""},
	}

	for _, tc := range testCases {
		result := GetStringOrDefault(tc.value, tc.defaultValue)
		if result != tc.expected {
			t.Errorf("GetStringOrDefault returned unexpected result: got %v, want %v", result, tc.expected)
		}
	}
}
