export type ModelSuggestion = {
  model_id: string
  display_name: string
}

export type FrontierTier = 'frontier' | 'standard' | 'free'

export type FrontierModel = {
  /** Model id as used with the provider (OpenRouter slug or native id). */
  model_id: string
  display_name: string
  /** Provider key for Settings / tenant_models (openrouter, openai, …). */
  provider_key: string
  /** Display vendor (Anthropic, OpenAI, …). */
  vendor: string
  tier: FrontierTier
  tags: string[]
  context?: string
}

/** Curated frontier + standard catalog — keep in sync with internal/keys/provider_catalog.go */
export const FRONTIER_MODELS: FrontierModel[] = [
  // —— Anthropic ——
  {
    model_id: 'claude-opus-4-20250514',
    display_name: 'Claude Opus 4',
    provider_key: 'anthropic',
    vendor: 'Anthropic',
    tier: 'frontier',
    tags: ['reasoning', 'long-context', 'premium'],
    context: '200K',
  },
  {
    model_id: 'claude-sonnet-4-20250514',
    display_name: 'Claude Sonnet 4',
    provider_key: 'anthropic',
    vendor: 'Anthropic',
    tier: 'frontier',
    tags: ['balanced', 'long-context', 'premium'],
    context: '200K',
  },
  {
    model_id: 'claude-3-7-sonnet-20250219',
    display_name: 'Claude 3.7 Sonnet',
    provider_key: 'anthropic',
    vendor: 'Anthropic',
    tier: 'frontier',
    tags: ['reasoning', 'long-context'],
    context: '200K',
  },
  {
    model_id: 'claude-3-5-sonnet-20241022',
    display_name: 'Claude 3.5 Sonnet',
    provider_key: 'anthropic',
    vendor: 'Anthropic',
    tier: 'standard',
    tags: ['balanced', 'long-context'],
    context: '200K',
  },
  {
    model_id: 'claude-3-5-haiku-20241022',
    display_name: 'Claude 3.5 Haiku',
    provider_key: 'anthropic',
    vendor: 'Anthropic',
    tier: 'standard',
    tags: ['fast', 'cheap'],
    context: '200K',
  },
  {
    model_id: 'anthropic/claude-sonnet-4',
    display_name: 'Claude Sonnet 4 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Anthropic',
    tier: 'frontier',
    tags: ['balanced', 'openrouter'],
    context: '200K',
  },
  {
    model_id: 'anthropic/claude-opus-4',
    display_name: 'Claude Opus 4 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Anthropic',
    tier: 'frontier',
    tags: ['reasoning', 'openrouter'],
    context: '200K',
  },
  {
    model_id: 'anthropic/claude-3.5-sonnet',
    display_name: 'Claude 3.5 Sonnet (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Anthropic',
    tier: 'standard',
    tags: ['openrouter'],
    context: '200K',
  },

  // —— OpenAI ——
  {
    model_id: 'gpt-4.1',
    display_name: 'GPT-4.1',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['flagship', 'multimodal'],
    context: '1M',
  },
  {
    model_id: 'gpt-4.1-mini',
    display_name: 'GPT-4.1 mini',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'standard',
    tags: ['fast', 'cheap'],
    context: '1M',
  },
  {
    model_id: 'gpt-4o',
    display_name: 'GPT-4o',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['multimodal', 'balanced'],
    context: '128K',
  },
  {
    model_id: 'gpt-4o-mini',
    display_name: 'GPT-4o mini',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'standard',
    tags: ['fast', 'cheap'],
    context: '128K',
  },
  {
    model_id: 'o1',
    display_name: 'OpenAI o1',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['reasoning', 'slow'],
    context: '200K',
  },
  {
    model_id: 'o3-mini',
    display_name: 'OpenAI o3-mini',
    provider_key: 'openai',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['reasoning', 'fast'],
    context: '200K',
  },
  {
    model_id: 'openai/gpt-4.1',
    display_name: 'GPT-4.1 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '1M',
  },
  {
    model_id: 'openai/gpt-4o',
    display_name: 'GPT-4o (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '128K',
  },
  {
    model_id: 'openai/o1',
    display_name: 'o1 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'OpenAI',
    tier: 'frontier',
    tags: ['reasoning', 'openrouter'],
    context: '200K',
  },

  // —— Google ——
  {
    model_id: 'gemini-2.5-pro',
    display_name: 'Gemini 2.5 Pro',
    provider_key: 'google',
    vendor: 'Google',
    tier: 'frontier',
    tags: ['reasoning', 'multimodal', 'long-context'],
    context: '1M',
  },
  {
    model_id: 'gemini-2.5-flash',
    display_name: 'Gemini 2.5 Flash',
    provider_key: 'google',
    vendor: 'Google',
    tier: 'frontier',
    tags: ['fast', 'multimodal'],
    context: '1M',
  },
  {
    model_id: 'gemini-2.0-flash',
    display_name: 'Gemini 2.0 Flash',
    provider_key: 'google',
    vendor: 'Google',
    tier: 'standard',
    tags: ['fast'],
    context: '1M',
  },
  {
    model_id: 'google/gemini-2.5-pro-preview',
    display_name: 'Gemini 2.5 Pro (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Google',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '1M',
  },
  {
    model_id: 'google/gemini-2.5-flash-preview',
    display_name: 'Gemini 2.5 Flash (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Google',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '1M',
  },

  // —— Meta ——
  {
    model_id: 'meta-llama/llama-4-maverick',
    display_name: 'Llama 4 Maverick (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Meta',
    tier: 'frontier',
    tags: ['open-source', 'openrouter'],
    context: '256K',
  },
  {
    model_id: 'meta-llama/llama-4-scout',
    display_name: 'Llama 4 Scout (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Meta',
    tier: 'frontier',
    tags: ['open-source', 'openrouter'],
    context: '512K',
  },
  {
    model_id: 'meta-llama/llama-3.3-70b-instruct',
    display_name: 'Llama 3.3 70B (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Meta',
    tier: 'standard',
    tags: ['open-source', 'openrouter'],
    context: '128K',
  },
  {
    model_id: 'llama-3.3-70b-versatile',
    display_name: 'Llama 3.3 70B',
    provider_key: 'groq',
    vendor: 'Meta',
    tier: 'standard',
    tags: ['fast', 'groq'],
    context: '128K',
  },

  // —— DeepSeek ——
  {
    model_id: 'deepseek-chat',
    display_name: 'DeepSeek V3',
    provider_key: 'openrouter',
    vendor: 'DeepSeek',
    tier: 'frontier',
    tags: ['reasoning', 'cheap'],
    context: '128K',
  },
  {
    model_id: 'deepseek/deepseek-chat',
    display_name: 'DeepSeek V3 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'DeepSeek',
    tier: 'frontier',
    tags: ['reasoning', 'openrouter'],
    context: '128K',
  },
  {
    model_id: 'deepseek/deepseek-r1',
    display_name: 'DeepSeek R1 (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'DeepSeek',
    tier: 'frontier',
    tags: ['reasoning', 'openrouter'],
    context: '128K',
  },
  {
    model_id: 'deepseek/deepseek-r1:free',
    display_name: 'DeepSeek R1 (Free)',
    provider_key: 'openrouter',
    vendor: 'DeepSeek',
    tier: 'free',
    tags: ['free', 'reasoning'],
    context: '128K',
  },

  // —— xAI ——
  {
    model_id: 'x-ai/grok-3',
    display_name: 'Grok 3',
    provider_key: 'openrouter',
    vendor: 'xAI',
    tier: 'frontier',
    tags: ['reasoning', 'openrouter'],
    context: '131K',
  },
  {
    model_id: 'x-ai/grok-2',
    display_name: 'Grok 2',
    provider_key: 'openrouter',
    vendor: 'xAI',
    tier: 'standard',
    tags: ['openrouter'],
    context: '131K',
  },

  // —— Mistral ——
  {
    model_id: 'mistral-large-latest',
    display_name: 'Mistral Large',
    provider_key: 'openrouter',
    vendor: 'Mistral',
    tier: 'frontier',
    tags: ['european'],
    context: '128K',
  },
  {
    model_id: 'mistralai/mistral-large-latest',
    display_name: 'Mistral Large (OpenRouter)',
    provider_key: 'openrouter',
    vendor: 'Mistral',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '128K',
  },
  {
    model_id: 'mistralai/pixtral-large-latest',
    display_name: 'Pixtral Large',
    provider_key: 'openrouter',
    vendor: 'Mistral',
    tier: 'frontier',
    tags: ['multimodal', 'openrouter'],
    context: '128K',
  },

  // —— Qwen ——
  {
    model_id: 'qwen/qwen-2.5-72b-instruct',
    display_name: 'Qwen 2.5 72B',
    provider_key: 'openrouter',
    vendor: 'Qwen',
    tier: 'standard',
    tags: ['openrouter'],
    context: '128K',
  },
  {
    model_id: 'qwen/qwen-max',
    display_name: 'Qwen Max',
    provider_key: 'openrouter',
    vendor: 'Qwen',
    tier: 'frontier',
    tags: ['openrouter'],
    context: '128K',
  },

  // —— Together / HF / Groq / Ollama ——
  {
    model_id: 'meta-llama/Llama-3.3-70B-Instruct-Turbo',
    display_name: 'Llama 3.3 70B Turbo',
    provider_key: 'together',
    vendor: 'Meta',
    tier: 'standard',
    tags: ['together'],
    context: '128K',
  },
  {
    model_id: 'meta-llama/Llama-3.1-8B-Instruct',
    display_name: 'Llama 3.1 8B',
    provider_key: 'huggingface',
    vendor: 'Meta',
    tier: 'free',
    tags: ['free', 'small'],
    context: '128K',
  },
  {
    model_id: 'llama3.3',
    display_name: 'Llama 3.3 (local)',
    provider_key: 'ollama',
    vendor: 'Meta',
    tier: 'free',
    tags: ['local', 'free'],
    context: '128K',
  },
  {
    model_id: 'qwen2.5:72b',
    display_name: 'Qwen 2.5 72B (local)',
    provider_key: 'ollama',
    vendor: 'Qwen',
    tier: 'free',
    tags: ['local', 'free'],
    context: '128K',
  },
]

export const FRONTIER_VENDORS = [
  'All',
  'Anthropic',
  'OpenAI',
  'Google',
  'Meta',
  'DeepSeek',
  'xAI',
  'Mistral',
  'Qwen',
] as const

export type FrontierVendor = (typeof FRONTIER_VENDORS)[number]

/** Build MODEL_SUGGESTIONS map for Settings quick-add chips. */
export function modelSuggestionsFromFrontier(): Record<string, ModelSuggestion[]> {
  const out: Record<string, ModelSuggestion[]> = {}
  for (const m of FRONTIER_MODELS) {
    if (!out[m.provider_key]) out[m.provider_key] = []
    const exists = out[m.provider_key]!.some((x) => x.model_id === m.model_id)
    if (!exists) {
      out[m.provider_key]!.push({ model_id: m.model_id, display_name: m.display_name })
    }
  }
  return out
}

export function frontierModelsForVendor(vendor: FrontierVendor, tier?: FrontierTier | 'all'): FrontierModel[] {
  let list = vendor === 'All' ? FRONTIER_MODELS : FRONTIER_MODELS.filter((m) => m.vendor === vendor)
  if (tier && tier !== 'all') {
    list = list.filter((m) => m.tier === tier)
  }
  return list
}
