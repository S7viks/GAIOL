package keys

import "testing"

func TestNormalizeProviderID(t *testing.T) {
	tests := map[string]string{
		"openrouter": "openrouter",
		"OpenAI":     "openai",
		"claude":     "anthropic",
		"Gemini":     "google",
		"groq":       "groq",
		"together":   "together",
		"ollama":     "ollama",
		"invalid":    "",
	}
	for in, want := range tests {
		if got := NormalizeProviderID(in); got != want {
			t.Errorf("NormalizeProviderID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestDefaultModelsForProvider(t *testing.T) {
	if len(DefaultModelsForProvider("openai")) == 0 {
		t.Fatal("expected openai default models")
	}
	if len(DefaultModelsForProvider("nope")) != 0 {
		t.Fatal("expected empty for unknown provider")
	}
}

func TestPingModelForProvider(t *testing.T) {
	if got := PingModelForProvider("openrouter"); got != "openai/gpt-4o-mini" {
		t.Fatalf("openrouter ping model = %q, want openai/gpt-4o-mini", got)
	}
	if got := PingModelForProvider("nope"); got != "" {
		t.Fatalf("unknown provider ping model = %q, want empty", got)
	}
}
