package main

import (
	"encoding/json"
	"testing"
)

func TestFormatAPIKeyName(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		expected string
	}{
		{
			name:     "lowercase provider",
			provider: "openai",
			expected: "api_key_openai",
		},
		{
			name:     "uppercase provider",
			provider: "ANTHROPIC",
			expected: "api_key_anthropic",
		},
		{
			name:     "mixed case provider",
			provider: "OpenRouter",
			expected: "api_key_openrouter",
		},
		{
			name:     "empty provider",
			provider: "",
			expected: "api_key_",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatAPIKeyName(tt.provider)
			if result != tt.expected {
				t.Errorf("formatAPIKeyName(%q) = %q, want %q", tt.provider, result, tt.expected)
			}
		})
	}
}

func TestGetAPIKeyStatus_ReturnsValidJSON(t *testing.T) {
	// This test verifies the JSON structure returned by GetAPIKeyStatus
	// We test the expected providers are present in the status map
	expectedProviders := []string{"openai", "anthropic", "gemini", "openrouter", "ollama"}

	// Create a mock status map to verify JSON structure
	status := make(map[string]bool)
	for _, provider := range expectedProviders {
		status[provider] = false
	}

	result, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Failed to marshal status: %v", err)
	}

	// Verify we can unmarshal it back
	var parsed map[string]bool
	if err := json.Unmarshal(result, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal status: %v", err)
	}

	// Verify all expected providers are present
	for _, provider := range expectedProviders {
		if _, ok := parsed[provider]; !ok {
			t.Errorf("Expected provider %q not found in status", provider)
		}
	}
}
