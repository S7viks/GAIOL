import { modelSuggestionsFromFrontier, type ModelSuggestion } from './frontier-models'

export type { ModelSuggestion }

export type ProviderOption = {
  value: string
  label: string
  helpUrl: string
  requiresApiKey: boolean
  optionalBaseUrl?: boolean
  defaultBaseUrl?: string
  placeholderKeyHint: string
}

/** Canonical provider list — keep in sync with internal/keys/provider_catalog.go */
export const PROVIDER_OPTIONS: ProviderOption[] = [
  {
    value: 'openrouter',
    label: 'OpenRouter',
    helpUrl: 'https://openrouter.ai/keys',
    requiresApiKey: true,
    placeholderKeyHint: 'sk-or-…',
  },
  {
    value: 'openai',
    label: 'OpenAI',
    helpUrl: 'https://platform.openai.com/api-keys',
    requiresApiKey: true,
    placeholderKeyHint: 'sk-…',
  },
  {
    value: 'anthropic',
    label: 'Anthropic (Claude)',
    helpUrl: 'https://console.anthropic.com/settings/keys',
    requiresApiKey: true,
    placeholderKeyHint: 'sk-ant-…',
  },
  {
    value: 'google',
    label: 'Google (Gemini)',
    helpUrl: 'https://aistudio.google.com/apikey',
    requiresApiKey: true,
    placeholderKeyHint: 'AI…',
  },
  {
    value: 'groq',
    label: 'Groq',
    helpUrl: 'https://console.groq.com/keys',
    requiresApiKey: true,
    placeholderKeyHint: 'gsk_…',
  },
  {
    value: 'together',
    label: 'Together AI',
    helpUrl: 'https://api.together.xyz/settings/api-keys',
    requiresApiKey: true,
    placeholderKeyHint: '…',
  },
  {
    value: 'huggingface',
    label: 'Hugging Face',
    helpUrl: 'https://huggingface.co/settings/tokens',
    requiresApiKey: true,
    placeholderKeyHint: 'hf_…',
  },
  {
    value: 'ollama',
    label: 'Ollama (local)',
    helpUrl: 'https://ollama.com',
    requiresApiKey: false,
    optionalBaseUrl: true,
    defaultBaseUrl: 'http://localhost:11434',
    placeholderKeyHint: 'not required',
  },
]

/** Quick-add chips in Settings — derived from frontier catalog. */
export const MODEL_SUGGESTIONS: Record<string, ModelSuggestion[]> = modelSuggestionsFromFrontier()

export function providerMeta(id: string): ProviderOption | undefined {
  return PROVIDER_OPTIONS.find((p) => p.value === id)
}

export function isLocalProvider(id: string): boolean {
  return id === 'ollama'
}
