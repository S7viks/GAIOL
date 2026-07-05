package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// PingProvider verifies credentials before saving in Settings.
// OpenRouter uses GET /api/v1/auth/key (no model spend). Other providers send a tiny completion.
func PingProvider(ctx context.Context, providerID, kind, apiKey, baseURL, modelID string) error {
	providerID = strings.TrimSpace(providerID)
	kind = strings.TrimSpace(kind)
	apiKey = strings.TrimSpace(apiKey)
	modelID = strings.TrimSpace(modelID)
	if providerID == "ollama" || kind == "ollama" {
		return nil // local; optional health check is separate
	}
	if apiKey == "" {
		return fmt.Errorf("api key is empty")
	}
	if providerID == "openrouter" {
		return pingOpenRouterKey(ctx, apiKey)
	}
	if modelID == "" {
		return fmt.Errorf("no default model configured for provider %q", providerID)
	}
	a, err := AdapterForCredential(providerID, kind, apiKey, baseURL)
	if err != nil {
		return err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()
	maxTok := 8
	temp := 0.0
	_, err = a.Generate(pingCtx, GenerateParams{
		Model:           modelID,
		Messages:        []ChatMessage{{Role: "user", Content: "Reply with exactly: OK"}},
		MaxOutputTokens: &maxTok,
		Temperature:     &temp,
	})
	if err != nil {
		return fmt.Errorf("provider rejected key: %w", err)
	}
	return nil
}

func pingOpenRouterKey(ctx context.Context, apiKey string) error {
	pingCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(pingCtx, http.MethodGet, "https://openrouter.ai/api/v1/auth/key", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("HTTP-Referer", "https://gaiol.app")
	req.Header.Set("X-Title", "GAIOL")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not reach OpenRouter: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized, http.StatusForbidden:
		if msg := openRouterErrorMessage(body); msg != "" {
			return fmt.Errorf("invalid or revoked OpenRouter API key: %s", msg)
		}
		return fmt.Errorf("invalid or revoked OpenRouter API key")
	default:
		if msg := openRouterErrorMessage(body); msg != "" {
			return fmt.Errorf("OpenRouter auth check failed (HTTP %d): %s", resp.StatusCode, msg)
		}
		return fmt.Errorf("OpenRouter auth check failed (HTTP %d)", resp.StatusCode)
	}
}

func openRouterErrorMessage(body []byte) string {
	var parsed struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &parsed); err == nil {
		if msg := strings.TrimSpace(parsed.Error.Message); msg != "" {
			return msg
		}
	}
	raw := strings.TrimSpace(string(body))
	if raw == "" {
		return ""
	}
	return compactAPIError(raw)
}

func compactAPIError(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) <= 160 {
		return raw
	}
	return raw[:157] + "..."
}
