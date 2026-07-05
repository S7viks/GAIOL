package keys

import (
	"context"
	"errors"
	"sort"
	"strings"

	"gaiol/internal/database"
)

// Credential kinds understood by the TS orchestrator (contract v1 credentials payload).
const (
	CredentialKindOpenAICompatible = "openai_compatible"
	CredentialKindAnthropic        = "anthropic_messages"
	CredentialKindGemini           = "gemini"
	CredentialKindOllama           = "ollama"
)

// ProviderCredential is one resolved provider for the orchestrate credentials payload.
// APIKey is a secret: never log it.
type ProviderCredential struct {
	ID      string
	Kind    string
	APIKey  string
	BaseURL string
}

// CredentialModel is one routable model in the tenant's provider pool.
type CredentialModel struct {
	Provider    string
	ModelID     string
	DisplayName string
}

// TenantCredentials is the tenant's full provider pool, ready to forward to the
// TS orchestrator on every user chat request.
type TenantCredentials struct {
	Providers []ProviderCredential
	Models    []CredentialModel
}

// builtinCredentialMeta maps a built-in provider id to the adapter kind and base URL the
// TS orchestrator uses. Base URLs mirror orchestrator/src/config/adapters.ts; empty base
// URL means the TS adapter default applies (Anthropic, Gemini).
var builtinCredentialMeta = map[string]struct {
	Kind    string
	BaseURL string
}{
	"openrouter":  {Kind: CredentialKindOpenAICompatible, BaseURL: "https://openrouter.ai/api/v1"},
	"openai":      {Kind: CredentialKindOpenAICompatible, BaseURL: "https://api.openai.com/v1"},
	"groq":        {Kind: CredentialKindOpenAICompatible, BaseURL: "https://api.groq.com/openai/v1"},
	"together":    {Kind: CredentialKindOpenAICompatible, BaseURL: "https://api.together.xyz/v1"},
	"huggingface": {Kind: CredentialKindOpenAICompatible, BaseURL: "https://router.huggingface.co/v1"},
	"anthropic":   {Kind: CredentialKindAnthropic},
	"google":      {Kind: CredentialKindGemini},
}

// CredentialMetaForBuiltin returns adapter kind and default base URL for a built-in provider id.
func CredentialMetaForBuiltin(provider string) (kind, baseURL string, ok bool) {
	cm, ok := builtinCredentialMeta[normalizeProvider(provider)]
	if !ok {
		return "", "", false
	}
	return cm.Kind, cm.BaseURL, true
}

// ResolveTenantCredentials loads the tenant's complete provider pool from the database:
// built-in provider keys (provider_api_keys), custom providers incl. Ollama
// (tenant_providers), and active tenant models (tenant_models).
//
// This is the single credential resolver for user inference: every chat surface goes
// Go shell -> ResolveTenantCredentials -> TS POST /v1/orchestrate. Returns a pool with
// zero providers (not an error) when the tenant has not connected any provider yet.
func ResolveTenantCredentials(ctx context.Context, db *database.Client, tenantID string) (*TenantCredentials, error) {
	if db == nil || db.Client == nil || strings.TrimSpace(tenantID) == "" {
		return nil, errors.New("database + tenant_id are required")
	}

	legacyKeys, err := LoadProviderKeysForTenant(ctx, db, tenantID)
	if err != nil {
		return nil, err
	}
	customProviders, err := LoadCustomProvidersForTenant(ctx, db, tenantID)
	if err != nil {
		return nil, err
	}
	tenantModels, err := LoadTenantModelsForTenant(ctx, db, tenantID)
	if err != nil {
		return nil, err
	}

	out := &TenantCredentials{}

	// Built-in providers in catalog order (deterministic payloads).
	for _, meta := range BuiltinProviders {
		apiKey := strings.TrimSpace(legacyKeys[meta.ID])
		if apiKey == "" {
			continue
		}
		cm, ok := builtinCredentialMeta[meta.ID]
		if !ok {
			continue
		}
		out.Providers = append(out.Providers, ProviderCredential{
			ID:      meta.ID,
			Kind:    cm.Kind,
			APIKey:  strings.TrimSpace(apiKey),
			BaseURL: cm.BaseURL,
		})
	}

	// Custom providers (tenant_providers), sorted by key for determinism.
	customKeys := make([]string, 0, len(customProviders))
	for pk := range customProviders {
		customKeys = append(customKeys, pk)
	}
	sort.Strings(customKeys)
	for _, pk := range customKeys {
		cfg := customProviders[pk]
		if pk == "ollama" {
			baseURL := strings.TrimSpace(cfg.BaseURL)
			if baseURL == "" {
				baseURL = "http://localhost:11434"
			}
			out.Providers = append(out.Providers, ProviderCredential{
				ID:      "ollama",
				Kind:    CredentialKindOllama,
				BaseURL: baseURL,
			})
			continue
		}
		switch strings.TrimSpace(strings.ToLower(cfg.ProviderType)) {
		case "", "openai_compatible":
			out.Providers = append(out.Providers, ProviderCredential{
				ID:      pk,
				Kind:    CredentialKindOpenAICompatible,
				APIKey:  strings.TrimSpace(cfg.APIKey),
				BaseURL: strings.TrimSpace(cfg.BaseURL),
			})
		case "anthropic_messages":
			out.Providers = append(out.Providers, ProviderCredential{
				ID:      pk,
				Kind:    CredentialKindAnthropic,
				APIKey:  strings.TrimSpace(cfg.APIKey),
				BaseURL: strings.TrimSpace(cfg.BaseURL),
			})
		}
	}

	resolved := make(map[string]bool, len(out.Providers))
	for _, p := range out.Providers {
		resolved[p.ID] = true
	}

	// Active tenant models whose provider is in the resolved pool.
	seenProviderModels := make(map[string]bool)
	for _, m := range tenantModels {
		pk := credentialProviderID(m.ProviderKey)
		if pk == "" || !m.IsActive || strings.TrimSpace(m.ModelID) == "" || !resolved[pk] {
			continue
		}
		out.Models = append(out.Models, CredentialModel{
			Provider:    pk,
			ModelID:     strings.TrimSpace(m.ModelID),
			DisplayName: strings.TrimSpace(m.DisplayName),
		})
		seenProviderModels[pk] = true
	}

	// Connected providers without registered models still get a catalog default so the pool
	// stays routable (mirror of EnsureDefaultModelsForProvider, without a DB write).
	for _, p := range out.Providers {
		if seenProviderModels[p.ID] {
			continue
		}
		suggestions := DefaultModelsForProvider(p.ID)
		if len(suggestions) == 0 {
			continue
		}
		out.Models = append(out.Models, CredentialModel{
			Provider:    p.ID,
			ModelID:     suggestions[0].ModelID,
			DisplayName: suggestions[0].DisplayName,
		})
	}

	return out, nil
}

// credentialProviderID normalizes a tenant_models provider_key to a credential provider id.
// Built-in aliases collapse (gemini -> google, claude -> anthropic); custom keys pass through.
func credentialProviderID(providerKey string) string {
	pk := strings.TrimSpace(strings.ToLower(providerKey))
	if pk == "" {
		return ""
	}
	if n := normalizeProvider(pk); n != "" {
		return n
	}
	return pk
}

func credentialsContainModel(models []CredentialModel, defaultModelID string) bool {
	parts := strings.SplitN(defaultModelID, ":", 2)
	if len(parts) != 2 {
		return false
	}
	provider := credentialProviderID(parts[0])
	modelID := strings.TrimSpace(parts[1])
	for _, m := range models {
		if m.Provider == provider && m.ModelID == modelID {
			return true
		}
	}
	return false
}
