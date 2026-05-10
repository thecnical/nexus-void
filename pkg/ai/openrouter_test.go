package ai

import (
	"os"
	"testing"
)

func TestNewOpenRouter(t *testing.T) {
	os.Setenv("OPENROUTER_API_KEY", "test-key")
	defer os.Unsetenv("OPENROUTER_API_KEY")

	client := NewOpenRouter()
	if client.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got %q", client.APIKey)
	}
	if client.Model != DefaultModel {
		t.Errorf("expected model %q, got %q", DefaultModel, client.Model)
	}
}

func TestNewOpenRouterFallback(t *testing.T) {
	os.Setenv("OPENROUTER_KEY", "fallback-key")
	defer os.Unsetenv("OPENROUTER_KEY")

	client := NewOpenRouter()
	if client.APIKey != "fallback-key" {
		t.Errorf("expected fallback key, got %q", client.APIKey)
	}
}

func TestOpenRouterNoKey(t *testing.T) {
	os.Unsetenv("OPENROUTER_API_KEY")
	os.Unsetenv("OPENROUTER_KEY")

	client := NewOpenRouter()
	_, err := client.Chat("system", "user")
	if err == nil {
		t.Error("expected error for missing API key")
	}
}
