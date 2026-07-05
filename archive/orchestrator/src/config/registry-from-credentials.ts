import { HeuristicDecomposer, LlmDecomposer, type LlmInfer } from "../decomposition/engine.js";
import type { Decomposer } from "../decomposition/types.js";
import type { ModelRegistryEntry } from "../domain/registry.js";
import { newTraceId } from "../observability/trace.js";
import { AnthropicMessagesAdapter } from "../providers/anthropic-adapter.js";
import type { LLMProviderAdapter } from "../providers/contract.js";
import { GeminiGenerativeAdapter } from "../providers/gemini-adapter.js";
import { OpenAICompatibleAdapter } from "../providers/openai-adapter.js";
import type { ProviderCredentialV1, RequestCredentialsV1 } from "../contract/v1/wire-types.js";

/**
 * Per-request runtime built from the tenant credentials payload. When a request
 * carries credentials this is the only key source: env provider keys are ignored.
 */
export interface CredentialRuntime {
  registry: ModelRegistryEntry[];
  adapters: Map<string, LLMProviderAdapter>;
  defaultModelId?: string;
}

function adapterForCredential(p: ProviderCredentialV1): LLMProviderAdapter | undefined {
  switch (p.kind) {
    case "openai_compatible": {
      if (!p.api_key) return undefined;
      return new OpenAICompatibleAdapter({
        apiKey: p.api_key,
        baseUrl: p.base_url,
        providerId: p.id,
        ...(p.id === "openrouter"
          ? {
              extraHeaders: {
                "HTTP-Referer": (process.env.OPENROUTER_HTTP_REFERER ?? "https://gaiol.local").trim(),
                "X-Title": (process.env.OPENROUTER_APP_TITLE ?? "GAIOL").trim(),
              },
            }
          : {}),
      });
    }
    case "anthropic_messages": {
      if (!p.api_key) return undefined;
      return new AnthropicMessagesAdapter({
        apiKey: p.api_key,
        ...(p.base_url ? { baseUrl: p.base_url } : {}),
      });
    }
    case "gemini": {
      if (!p.api_key) return undefined;
      return new GeminiGenerativeAdapter({
        apiKey: p.api_key,
        ...(p.base_url ? { baseUrl: p.base_url } : {}),
      });
    }
    case "ollama": {
      // Ollama exposes an OpenAI-compatible API under /v1; the bearer token is unused.
      const base = (p.base_url ?? "http://localhost:11434").replace(/\/$/, "");
      return new OpenAICompatibleAdapter({
        apiKey: p.api_key || "ollama",
        baseUrl: base.endsWith("/v1") ? base : `${base}/v1`,
        providerId: p.id,
      });
    }
    default:
      return undefined;
  }
}

/**
 * Builds the per-request registry + adapter map from tenant credentials.
 * Registry model ids use the `provider:model` format shared with the Go shell
 * (e.g. tenant_settings.default_model_id = "openai:gpt-4o-mini").
 */
export function buildRuntimeFromCredentials(creds: RequestCredentialsV1): CredentialRuntime {
  const adapters = new Map<string, LLMProviderAdapter>();
  for (const p of creds.providers) {
    const id = p.id.trim();
    if (!id || adapters.has(id)) continue;
    const adapter = adapterForCredential(p);
    if (adapter) {
      adapters.set(id, adapter);
    }
  }

  const registry: ModelRegistryEntry[] = [];
  const seen = new Set<string>();
  for (const m of creds.models ?? []) {
    const provider = m.provider.trim();
    const remoteName = m.model_id.trim();
    if (!provider || !remoteName || !adapters.has(provider)) continue;
    const modelId = `${provider}:${remoteName}`;
    if (seen.has(modelId)) continue;
    seen.add(modelId);
    registry.push({
      modelId,
      providerId: provider,
      remoteName,
      capabilities: ["general", "reasoning", "code"],
      costIndex: 0.3,
      latencyPriorMs: 800,
      accuracyPrior: 0.7,
      available: true,
    });
  }

  const runtime: CredentialRuntime = { registry, adapters };
  const def = creds.default_model_id?.trim();
  if (def && seen.has(def)) {
    runtime.defaultModelId = def;
  }
  return runtime;
}

/**
 * LLM decomposer over the credential pool (default model first). Falls back to the
 * heuristic decomposer when the pool is empty or GAIOL_LLM_DECOMPOSER is off.
 */
export function buildDecomposerFromCredentials(
  runtime: CredentialRuntime,
  env: NodeJS.ProcessEnv = process.env,
): Decomposer {
  const flag = (env.GAIOL_LLM_DECOMPOSER ?? "").trim().toLowerCase();
  if (flag === "0" || flag === "false" || flag === "off") {
    return new HeuristicDecomposer();
  }

  const entry =
    (runtime.defaultModelId
      ? runtime.registry.find((e) => e.modelId === runtime.defaultModelId && e.available)
      : undefined) ?? runtime.registry.find((e) => e.available);
  if (!entry) {
    return new HeuristicDecomposer();
  }
  const adapter = runtime.adapters.get(entry.providerId);
  if (!adapter) {
    return new HeuristicDecomposer();
  }

  const infer: LlmInfer = async (prompt: string) => {
    const result = await adapter.generate({
      traceId: newTraceId(),
      model: entry.remoteName,
      messages: [{ role: "user", content: prompt }],
      temperature: 0,
      maxOutputTokens: 1024,
    });
    return result.text;
  };

  return new LlmDecomposer(infer);
}
