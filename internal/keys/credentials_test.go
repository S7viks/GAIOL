package keys

import "testing"

// Every built-in provider that stores an API key must map to a credential kind so it
// reaches the orchestrator credentials payload. Ollama is covered separately (custom
// provider row with kind "ollama").
func TestBuiltinCredentialMetaCoversCatalog(t *testing.T) {
	for _, p := range BuiltinProviders {
		if p.ID == "ollama" {
			if _, ok := builtinCredentialMeta[p.ID]; ok {
				t.Errorf("ollama must not be in builtinCredentialMeta (it resolves via tenant_providers)")
			}
			continue
		}
		meta, ok := builtinCredentialMeta[p.ID]
		if !ok {
			t.Errorf("provider %q has no credential mapping; it would be dropped from the orchestrate payload", p.ID)
			continue
		}
		switch meta.Kind {
		case CredentialKindOpenAICompatible:
			if meta.BaseURL == "" {
				t.Errorf("provider %q: openai_compatible credentials require an explicit base_url", p.ID)
			}
		case CredentialKindAnthropic, CredentialKindGemini:
			// TS adapter defaults apply; base URL optional.
		default:
			t.Errorf("provider %q: unexpected credential kind %q", p.ID, meta.Kind)
		}
	}
}

func TestCredentialProviderIDNormalization(t *testing.T) {
	cases := map[string]string{
		"openai":      "openai",
		"OpenAI":      "openai",
		"gemini":      "google",
		"google":      "google",
		"claude":      "anthropic",
		"anthropic":   "anthropic",
		"openrouter":  "openrouter",
		"groq":        "groq",
		"together":    "together",
		"huggingface": "huggingface",
		"ollama":      "ollama",
		"my-custom":   "my-custom",
		"":            "",
		"  ":          "",
	}
	for in, want := range cases {
		if got := credentialProviderID(in); got != want {
			t.Errorf("credentialProviderID(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCredentialsContainModel(t *testing.T) {
	models := []CredentialModel{
		{Provider: "openai", ModelID: "gpt-4o-mini"},
		{Provider: "google", ModelID: "gemini-2.0-flash"},
	}
	if !credentialsContainModel(models, "openai:gpt-4o-mini") {
		t.Error("expected openai:gpt-4o-mini to be present")
	}
	// gemini alias must collapse to google.
	if !credentialsContainModel(models, "gemini:gemini-2.0-flash") {
		t.Error("expected gemini:gemini-2.0-flash to match google provider entry")
	}
	if credentialsContainModel(models, "anthropic:claude-3-5-sonnet-20241022") {
		t.Error("did not expect anthropic model to be present")
	}
	if credentialsContainModel(models, "not-a-model-id") {
		t.Error("malformed default_model_id must not match")
	}
}

// Every built-in provider must have at least one catalog default so a freshly
// connected provider is immediately routable in the credentials payload.
func TestDefaultModelsExistForAllBuiltinProviders(t *testing.T) {
	for _, p := range BuiltinProviders {
		if len(DefaultModelsForProvider(p.ID)) == 0 {
			t.Errorf("provider %q has no default model suggestions", p.ID)
		}
	}
}
