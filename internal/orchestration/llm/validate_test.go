package llm

import (
	"context"
	"strings"
	"testing"
)

func TestPingProvider_OpenRouter_InvalidKey(t *testing.T) {
	t.Parallel()
	err := PingProvider(context.Background(), "openrouter", "openai_compatible", "sk-or-v1-invalid-test-key", "", "")
	if err == nil {
		t.Fatal("expected error for invalid OpenRouter key")
	}
	if !strings.Contains(err.Error(), "OpenRouter") {
		t.Fatalf("expected OpenRouter-specific error, got: %v", err)
	}
}

func TestPingProvider_Ollama_SkipsPing(t *testing.T) {
	t.Parallel()
	if err := PingProvider(context.Background(), "ollama", "ollama", "", "http://localhost:11434", ""); err != nil {
		t.Fatalf("ollama should skip ping: %v", err)
	}
}
