package utils

import "testing"

func TestNormalizeAIBaseURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no version suffix",
			input:    "https://api.openai.com",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "with trailing slash",
			input:    "https://api.openai.com/",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "already has v1",
			input:    "https://api.openai.com/v1",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "already has v1 with trailing slash",
			input:    "https://api.openai.com/v1/",
			expected: "https://api.openai.com/v1",
		},
		{
			name:     "has v2",
			input:    "https://api.openai.com/v2",
			expected: "https://api.openai.com/v2",
		},
		{
			name:     "has v3 with trailing slash",
			input:    "https://api.openai.com/v3/",
			expected: "https://api.openai.com/v3",
		},
		{
			name:     "has v10",
			input:    "https://api.openai.com/v10",
			expected: "https://api.openai.com/v10",
		},
		{
			name:     "has non-version path",
			input:    "https://api.openai.com/api",
			expected: "https://api.openai.com/api/v1",
		},
		{
			name:     "complex path without version",
			input:    "https://api.openai.com/chat/completions",
			expected: "https://api.openai.com/chat/completions/v1",
		},
		{
			name:     "complex path with version",
			input:    "https://api.openai.com/chat/v2",
			expected: "https://api.openai.com/chat/v2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeAIBaseURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAIBaseURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
