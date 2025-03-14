package util

import (
	"testing"
)

func TestResolvePlaceholders(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		env      map[string]string
		expected string
	}{
		{
			name:     "Single replacement",
			input:    "Hello, $USER!",
			env:      map[string]string{"USER": "Mom"},
			expected: "Hello, Mom!",
		},
		{
			name:     "Multiple replacements",
			input:    "User: $USER, Env: $ENV",
			env:      map[string]string{"USER": "Mom", "ENV": "prod"},
			expected: "User: Mom, Env: prod",
		},
		{
			name:     "Braced variable replacement",
			input:    "Welcome, ${USER}!",
			env:      map[string]string{"USER": "Bob"},
			expected: "Welcome, Bob!",
		},
		{
			name:     "Mixed $VAR and ${VAR}",
			input:    "$USER is working on ${PROJECT}",
			env:      map[string]string{"USER": "Charlie", "PROJECT": "GoApp"},
			expected: "Charlie is working on GoApp",
		},
		{
			name:     "Variable not found (keep original)",
			input:    "Hello, $UNKNOWN!",
			env:      map[string]string{"USER": "Mom"},
			expected: "Hello, $UNKNOWN!",
		},
		{
			name:     "Braced variable not found (keep original)",
			input:    "Welcome, ${UNKNOWN}!",
			env:      map[string]string{"USER": "Mom"},
			expected: "Welcome, ${UNKNOWN}!",
		},
		{
			name:     "Adjacent variables",
			input:    "$USER$ROLE",
			env:      map[string]string{"USER": "Dev", "ROLE": "Ops"},
			expected: "DevOps",
		},
		{
			name:     "Dollar sign without variable (should remain)",
			input:    "Price is $5 per unit",
			env:      map[string]string{"PRICE": "10"},
			expected: "Price is $5 per unit",
		},
		{
			name:     "Partial match without full variable name",
			input:    "This is $USER_NAME",
			env:      map[string]string{"USER": "Admin"},
			expected: "This is $USER_NAME",
		},
		{
			name:     "Handles underscores in variables",
			input:    "${APP_ENV} mode",
			env:      map[string]string{"APP_ENV": "production"},
			expected: "production mode",
		},
		{
			name:     "Handles empty input string",
			input:    "",
			env:      map[string]string{"ANY": "Value"},
			expected: "",
		},
		{
			name:     "Handles empty variable values",
			input:    "API_KEY=${API_KEY}",
			env:      map[string]string{"API_KEY": ""},
			expected: "API_KEY=",
		},
		{
			name:     "Handles only dollar sign",
			input:    "$",
			env:      map[string]string{"DOLLAR": "money"},
			expected: "$",
		},
		{
			name:     "Handles variable at the start",
			input:    "$VERSION is released",
			env:      map[string]string{"VERSION": "1.2.3"},
			expected: "1.2.3 is released",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResolveEnvPlaceholders(tt.input, tt.env)
			if result != tt.expected {
				t.Errorf("resolvePlaceholders(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
