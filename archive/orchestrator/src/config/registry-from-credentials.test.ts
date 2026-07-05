import { describe, expect, it } from "vitest";
import { buildRuntimeFromCredentials } from "./registry-from-credentials.js";
import type { RequestCredentialsV1 } from "../contract/v1/wire-types.js";

describe("buildRuntimeFromCredentials", () => {
  it("builds registry and adapters from tenant credentials", () => {
    const creds: RequestCredentialsV1 = {
      schema_version: "1",
      providers: [
        {
          id: "openai",
          kind: "openai_compatible",
          api_key: "sk-test",
          base_url: "https://api.openai.com/v1",
        },
        {
          id: "google",
          kind: "gemini",
          api_key: "gemini-test",
        },
      ],
      models: [
        { provider: "openai", model_id: "gpt-4o-mini", display_name: "GPT-4o mini" },
        { provider: "google", model_id: "gemini-2.0-flash", display_name: "Gemini 2.0 Flash" },
      ],
      default_model_id: "openai:gpt-4o-mini",
    };

    const runtime = buildRuntimeFromCredentials(creds);
    expect(runtime.adapters.has("openai")).toBe(true);
    expect(runtime.adapters.has("google")).toBe(true);
    expect(runtime.registry.length).toBe(2);
    expect(runtime.registry.some((e) => e.modelId === "openai:gpt-4o-mini")).toBe(true);
    expect(runtime.defaultModelId).toBe("openai:gpt-4o-mini");
  });

  it("ignores models whose provider is not in the credentials pool", () => {
    const creds: RequestCredentialsV1 = {
      schema_version: "1",
      providers: [
        { id: "groq", kind: "openai_compatible", api_key: "gsk_test", base_url: "https://api.groq.com/openai/v1" },
      ],
      models: [
        { provider: "groq", model_id: "llama-3.3-70b-versatile" },
        { provider: "openai", model_id: "gpt-4o", display_name: "orphan" },
      ],
    };

    const runtime = buildRuntimeFromCredentials(creds);
    expect(runtime.registry.length).toBe(1);
    expect(runtime.registry[0]?.modelId).toBe("groq:llama-3.3-70b-versatile");
  });

  it("wires ollama without an api key", () => {
    const creds: RequestCredentialsV1 = {
      schema_version: "1",
      providers: [{ id: "ollama", kind: "ollama", base_url: "http://localhost:11434" }],
      models: [{ provider: "ollama", model_id: "llama3.2" }],
    };

    const runtime = buildRuntimeFromCredentials(creds);
    expect(runtime.adapters.has("ollama")).toBe(true);
    expect(runtime.registry[0]?.remoteName).toBe("llama3.2");
  });
});
