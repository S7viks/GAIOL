package orchestration

import (
	"log"
	"os"
	"strings"

	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
	"gaiol/internal/orchestration/llm"
)

// Runtime is per-request or env-backed registry + adapters.
type Runtime struct {
	Registry []ModelRegistryEntry
	Adapters map[string]llm.ProviderAdapter
}

func sampleRegistry() []ModelRegistryEntry {
	acc55 := 0.55
	acc70 := 0.70
	acc62 := 0.62
	return []ModelRegistryEntry{
		{ModelID: "mock-fast", ProviderID: "mock", RemoteName: "mock-fast", Capabilities: []string{"general"}, CostIndex: 0.1, LatencyPriorMs: 50, AccuracyPrior: &acc55, Available: true},
		{ModelID: "mock-strong", ProviderID: "mock", RemoteName: "mock-strong", Capabilities: []string{"reasoning", "general"}, CostIndex: 0.4, LatencyPriorMs: 120, AccuracyPrior: &acc70, Available: true},
		{ModelID: "mock-code", ProviderID: "mock", RemoteName: "mock-code", Capabilities: []string{"code"}, CostIndex: 0.25, LatencyPriorMs: 90, AccuracyPrior: &acc62, Available: true},
	}
}

func mockAdapters() map[string]llm.ProviderAdapter {
	return map[string]llm.ProviderAdapter{
		"mock": llm.NewMockAdapter("mock"),
	}
}

// RuntimeFromEnv builds registry/adapters from process environment (no-auth dev).
func RuntimeFromEnv() *Runtime {
	registry := liveRegistryFromEnv()
	adapters := make(map[string]llm.ProviderAdapter)
	if len(registry) == 0 {
		return &Runtime{Registry: sampleRegistry(), Adapters: mockAdapters()}
	}
	for _, e := range registry {
		if _, ok := adapters[e.ProviderID]; ok {
			continue
		}
		switch e.ProviderID {
		case "openai-compatible", "openai", "openrouter", "groq", "together", "huggingface":
			key := envFirst("OPENAI_API_KEY", "OPENROUTER_API_KEY", "GROQ_API_KEY", "HUGGINGFACE_API_KEY", "TOGETHER_API_KEY")
			base := strings.TrimSpace(os.Getenv("OPENAI_BASE_URL"))
			if a, err := llm.AdapterForCredential(e.ProviderID, "openai_compatible", key, base); err == nil {
				adapters[e.ProviderID] = a
			}
		case "anthropic":
			if a, err := llm.AdapterForCredential("anthropic", "anthropic_messages", os.Getenv("ANTHROPIC_API_KEY"), ""); err == nil {
				adapters[e.ProviderID] = a
			}
		case "google-gemini", "google":
			if a, err := llm.AdapterForCredential("google", "gemini", envFirst("GEMINI_API_KEY", "GOOGLE_API_KEY"), ""); err == nil {
				adapters[e.ProviderID] = a
			}
		}
	}
	if len(adapters) == 0 {
		return &Runtime{Registry: sampleRegistry(), Adapters: mockAdapters()}
	}
	return &Runtime{Registry: registry, Adapters: adapters}
}

func envFirst(keys ...string) string {
	for _, k := range keys {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

func liveRegistryFromEnv() []ModelRegistryEntry {
	var out []ModelRegistryEntry
	acc := func(v float64) *float64 { return &v }
	if strings.TrimSpace(os.Getenv("OPENAI_API_KEY")) != "" {
		out = append(out, ModelRegistryEntry{
			ModelID: "openai-primary", ProviderID: "openai-compatible",
			RemoteName: envOr("OPENAI_ORCHESTRATOR_MODEL", "gpt-4o-mini"),
			Capabilities: []string{"general", "reasoning", "code"}, CostIndex: 0.35, LatencyPriorMs: 900, AccuracyPrior: acc(0.72), Available: true,
		})
	}
	if strings.TrimSpace(os.Getenv("ANTHROPIC_API_KEY")) != "" {
		out = append(out, ModelRegistryEntry{
			ModelID: "anthropic-primary", ProviderID: "anthropic",
			RemoteName: envOr("ANTHROPIC_ORCHESTRATOR_MODEL", "claude-3-5-haiku-20241022"),
			Capabilities: []string{"general", "reasoning", "code"}, CostIndex: 0.38, LatencyPriorMs: 950, AccuracyPrior: acc(0.73), Available: true,
		})
	}
	if k := envFirst("GEMINI_API_KEY", "GOOGLE_API_KEY"); k != "" {
		out = append(out, ModelRegistryEntry{
			ModelID: "gemini-primary", ProviderID: "google-gemini",
			RemoteName: envOr("GEMINI_ORCHESTRATOR_MODEL", "gemini-2.0-flash"),
			Capabilities: []string{"general", "reasoning", "code"}, CostIndex: 0.3, LatencyPriorMs: 850, AccuracyPrior: acc(0.71), Available: true,
		})
	}
	if strings.TrimSpace(os.Getenv("OPENROUTER_API_KEY")) != "" {
		out = append(out, ModelRegistryEntry{
			ModelID: "openrouter:meta-llama/llama-3.1-8b-instruct:free", ProviderID: "openrouter",
			RemoteName: "meta-llama/llama-3.1-8b-instruct:free",
			Capabilities: []string{"general"}, CostIndex: 0.05, LatencyPriorMs: 600, AccuracyPrior: acc(0.65), Available: true,
		})
	}
	return out
}

func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}

// RuntimeFromCredentials builds per-request runtime from v1 credentials.
func RuntimeFromCredentials(creds *orchestratorv1.RequestCredentialsV1) (*Runtime, error) {
	if creds == nil {
		return nil, nil
	}
	adapters := make(map[string]llm.ProviderAdapter)
	for _, p := range creds.Providers {
		id := strings.TrimSpace(p.ID)
		if id == "" || adapters[id] != nil {
			continue
		}
		a, err := llm.AdapterForCredential(id, p.Kind, p.APIKey, p.BaseURL)
		if err != nil {
			log.Printf("orchestration: skip provider %q: %v", id, err)
			continue
		}
		if a != nil {
			adapters[id] = a
		}
	}
	registry := make([]ModelRegistryEntry, 0, len(creds.Models))
	seen := make(map[string]struct{})
	acc := 0.7
	accPtr := &acc
	for _, m := range creds.Models {
		provider := strings.TrimSpace(m.Provider)
		remote := strings.TrimSpace(m.ModelID)
		if provider == "" || remote == "" {
			continue
		}
		if adapters[provider] == nil {
			continue
		}
		modelID := provider + ":" + remote
		if _, ok := seen[modelID]; ok {
			continue
		}
		seen[modelID] = struct{}{}
		registry = append(registry, ModelRegistryEntry{
			ModelID: modelID, ProviderID: provider, RemoteName: remote,
			Capabilities: []string{"general", "reasoning", "code"}, CostIndex: 0.3, LatencyPriorMs: 800, AccuracyPrior: accPtr, Available: true,
		})
	}
	rt := &Runtime{Registry: registry, Adapters: adapters}
	return rt, nil
}

func entryByID(registry []ModelRegistryEntry, id string) *ModelRegistryEntry {
	for i := range registry {
		if registry[i].ModelID == id {
			return &registry[i]
		}
	}
	return nil
}
