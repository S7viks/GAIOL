package keys

import "strings"

// DefaultModelSuggestion is used when auto-registering models after a provider key is saved.
type DefaultModelSuggestion struct {
	ModelID     string
	DisplayName string
}

// BuiltinProviderMeta describes a first-class provider in the dashboard dropdown.
type BuiltinProviderMeta struct {
	ID                 string
	Label              string
	HelpURL            string
	RequiresAPIKey     bool
	OptionalBaseURL    bool
	DefaultBaseURL     string
	PlaceholderKeyHint string
}

// BuiltinProviders is the canonical list for Settings / onboarding UI and API validation.
var BuiltinProviders = []BuiltinProviderMeta{
	{ID: "openrouter", Label: "OpenRouter", HelpURL: "https://openrouter.ai/keys", RequiresAPIKey: true, PlaceholderKeyHint: "sk-or-…"},
	{ID: "openai", Label: "OpenAI", HelpURL: "https://platform.openai.com/api-keys", RequiresAPIKey: true, PlaceholderKeyHint: "sk-…"},
	{ID: "anthropic", Label: "Anthropic (Claude)", HelpURL: "https://console.anthropic.com/settings/keys", RequiresAPIKey: true, PlaceholderKeyHint: "sk-ant-…"},
	{ID: "google", Label: "Google (Gemini)", HelpURL: "https://aistudio.google.com/apikey", RequiresAPIKey: true, PlaceholderKeyHint: "AI…"},
	{ID: "groq", Label: "Groq", HelpURL: "https://console.groq.com/keys", RequiresAPIKey: true, PlaceholderKeyHint: "gsk_…"},
	{ID: "together", Label: "Together AI", HelpURL: "https://api.together.xyz/settings/api-keys", RequiresAPIKey: true, PlaceholderKeyHint: "…"},
	{ID: "huggingface", Label: "Hugging Face", HelpURL: "https://huggingface.co/settings/tokens", RequiresAPIKey: true, PlaceholderKeyHint: "hf_…"},
	{ID: "ollama", Label: "Ollama (local)", HelpURL: "https://ollama.com", RequiresAPIKey: false, OptionalBaseURL: true, DefaultBaseURL: "http://localhost:11434", PlaceholderKeyHint: "optional"},
}

var defaultModelsByProvider = map[string][]DefaultModelSuggestion{
	"openrouter": {
		{ModelID: "anthropic/claude-opus-4", DisplayName: "Claude Opus 4 (OpenRouter)"},
		{ModelID: "anthropic/claude-sonnet-4", DisplayName: "Claude Sonnet 4 (OpenRouter)"},
		{ModelID: "anthropic/claude-3.5-sonnet", DisplayName: "Claude 3.5 Sonnet (OpenRouter)"},
		{ModelID: "openai/gpt-4.1", DisplayName: "GPT-4.1 (OpenRouter)"},
		{ModelID: "openai/gpt-4o", DisplayName: "GPT-4o (OpenRouter)"},
		{ModelID: "openai/o1", DisplayName: "o1 (OpenRouter)"},
		{ModelID: "google/gemini-2.5-pro-preview", DisplayName: "Gemini 2.5 Pro (OpenRouter)"},
		{ModelID: "google/gemini-2.5-flash-preview", DisplayName: "Gemini 2.5 Flash (OpenRouter)"},
		{ModelID: "meta-llama/llama-4-maverick", DisplayName: "Llama 4 Maverick (OpenRouter)"},
		{ModelID: "deepseek/deepseek-r1", DisplayName: "DeepSeek R1 (OpenRouter)"},
		{ModelID: "deepseek/deepseek-chat", DisplayName: "DeepSeek V3 (OpenRouter)"},
		{ModelID: "x-ai/grok-3", DisplayName: "Grok 3 (OpenRouter)"},
		{ModelID: "mistralai/mistral-large-latest", DisplayName: "Mistral Large (OpenRouter)"},
		{ModelID: "qwen/qwen-max", DisplayName: "Qwen Max (OpenRouter)"},
	},
	"openai": {
		{ModelID: "gpt-4.1", DisplayName: "GPT-4.1"},
		{ModelID: "gpt-4.1-mini", DisplayName: "GPT-4.1 mini"},
		{ModelID: "gpt-4o", DisplayName: "GPT-4o"},
		{ModelID: "gpt-4o-mini", DisplayName: "GPT-4o mini"},
		{ModelID: "o1", DisplayName: "OpenAI o1"},
		{ModelID: "o3-mini", DisplayName: "OpenAI o3-mini"},
	},
	"anthropic": {
		{ModelID: "claude-opus-4-20250514", DisplayName: "Claude Opus 4"},
		{ModelID: "claude-sonnet-4-20250514", DisplayName: "Claude Sonnet 4"},
		{ModelID: "claude-3-7-sonnet-20250219", DisplayName: "Claude 3.7 Sonnet"},
		{ModelID: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet"},
		{ModelID: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku"},
	},
	"google": {
		{ModelID: "gemini-2.5-pro", DisplayName: "Gemini 2.5 Pro"},
		{ModelID: "gemini-2.5-flash", DisplayName: "Gemini 2.5 Flash"},
		{ModelID: "gemini-2.0-flash", DisplayName: "Gemini 2.0 Flash"},
		{ModelID: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash"},
	},
	"groq": {
		{ModelID: "llama-3.3-70b-versatile", DisplayName: "Llama 3.3 70B"},
		{ModelID: "mixtral-8x7b-32768", DisplayName: "Mixtral 8x7B"},
		{ModelID: "deepseek-r1-distill-llama-70b", DisplayName: "DeepSeek R1 Distill 70B"},
	},
	"together": {
		{ModelID: "meta-llama/Llama-3.3-70B-Instruct-Turbo", DisplayName: "Llama 3.3 70B Turbo"},
		{ModelID: "meta-llama/Llama-3.1-8B-Instruct", DisplayName: "Llama 3.1 8B Instruct"},
		{ModelID: "mistralai/Mixtral-8x7B-Instruct-v0.1", DisplayName: "Mixtral 8x7B Instruct"},
	},
	"huggingface": {
		{ModelID: "meta-llama/Llama-3.1-8B-Instruct", DisplayName: "Llama 3.1 8B Instruct"},
		{ModelID: "Qwen/Qwen2.5-7B-Instruct", DisplayName: "Qwen 2.5 7B Instruct"},
		{ModelID: "mistralai/Mistral-7B-Instruct-v0.3", DisplayName: "Mistral 7B Instruct"},
	},
	"ollama": {
		{ModelID: "llama3.3", DisplayName: "Llama 3.3"},
		{ModelID: "qwen2.5:72b", DisplayName: "Qwen 2.5 72B"},
		{ModelID: "llama3.2", DisplayName: "Llama 3.2"},
		{ModelID: "mistral", DisplayName: "Mistral"},
	},
}

// DefaultModelsForProvider returns suggested models for auto-registration.
func DefaultModelsForProvider(provider string) []DefaultModelSuggestion {
	p := normalizeProvider(provider)
	if p == "" {
		return nil
	}
	out := defaultModelsByProvider[p]
	if len(out) == 0 {
		return nil
	}
	dup := make([]DefaultModelSuggestion, len(out))
	copy(dup, out)
	return dup
}

// IsBuiltinProvider reports whether p is a supported built-in provider id.
func IsBuiltinProvider(p string) bool {
	return normalizeProvider(p) != ""
}

// IsLocalProvider is true when the provider does not require a remote API key (e.g. Ollama).
func IsLocalProvider(p string) bool {
	p = normalizeProvider(p)
	return p == "ollama"
}

func normalizeProvider(p string) string {
	p = strings.TrimSpace(strings.ToLower(p))
	switch p {
	case "openrouter", "huggingface", "openai", "anthropic", "groq", "together", "ollama":
		return p
	case "google", "gemini":
		return "google"
	case "claude":
		return "anthropic"
	case "gpt", "chatgpt":
		return "openai"
	}
	return ""
}

// NormalizeProviderID exports provider normalization for HTTP handlers.
func NormalizeProviderID(p string) string {
	return normalizeProvider(p)
}

// pingModelByProvider lists lightweight models used to verify API keys before save.
// Prefer cheap/widely-available models over frontier defaults in defaultModelsByProvider.
var pingModelByProvider = map[string]string{
	"openrouter":  "openai/gpt-4o-mini",
	"openai":      "gpt-4o-mini",
	"anthropic":   "claude-3-5-haiku-20241022",
	"google":      "gemini-1.5-flash",
	"groq":        "llama-3.1-8b-instant",
	"together":    "meta-llama/Llama-3.1-8B-Instruct",
	"huggingface": "meta-llama/Llama-3.1-8B-Instruct",
}

// PingModelForProvider returns the model id used to validate a provider key on save.
func PingModelForProvider(provider string) string {
	p := normalizeProvider(provider)
	if p == "" {
		return ""
	}
	if m := pingModelByProvider[p]; m != "" {
		return m
	}
	suggestions := DefaultModelsForProvider(p)
	if len(suggestions) == 0 {
		return ""
	}
	return suggestions[0].ModelID
}
