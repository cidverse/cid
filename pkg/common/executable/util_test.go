package executable

import (
	"testing"
	"time"
)

func TestReplaceCommandPlaceholders(t *testing.T) {
	env := map[string]string{
		"HOME": "/user/home",
		"USER": "john.doe",
	}

	tests := []struct {
		input string
		want  string
	}{
		{
			input: "echo 'Hello, {USER}!'",
			want:  "echo 'Hello, john.doe!'",
		},
		{
			input: "ls {HOME}",
			want:  "ls /user/home",
		},
		{
			input: "kubectl logs my-pod --since-time={TIMESTAMP_RFC3339}",
			want:  "kubectl logs my-pod --since-time=" + time.Now().Format(time.RFC3339),
		},
		{
			input: "echo 'Missing placeholder {NOT_FOUND}'",
			want:  "echo 'Missing placeholder {NOT_FOUND}'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ReplaceCommandPlaceholders(tt.input, env)
			if got != tt.want {
				t.Errorf("ReplaceCommandPlaceholders() = %q, want %q", got, tt.want)
			}
		})
	}
}
