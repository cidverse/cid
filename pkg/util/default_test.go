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

func TestFirstNonEmpty(t *testing.T) {
	tests := []struct {
		strings []string
		result  string
	}{
		{[]string{"", "hello", "world"}, "hello"},
		{[]string{"", "", "hello"}, "hello"},
		{[]string{"", "", ""}, ""},
		{[]string{"hello", "world", "golang"}, "hello"},
		{[]string{"golang", "world", "hello"}, "golang"},
	}

	for _, test := range tests {
		res := FirstNonEmpty(test.strings)
		if res != test.result {
			t.Errorf("FirstNonEmpty(%v) = %v, want %v", test.strings, res, test.result)
		}
	}
}
