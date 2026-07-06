package orchestration

import (
	"context"
	"strings"
)

// expandCallsOnProviderFailure retries other providers when every selected model failed
// with a billing/auth/rate-limit style error (e.g. OpenRouter 402 insufficient credits).
func (p *Pipeline) expandCallsOnProviderFailure(
	ctx context.Context,
	req OrchestrationRequest,
	subObjective string,
	cfg OrchestratorConfig,
	plan RoutingPlan,
	calls []ModelCallResult,
) ([]ModelCallResult, int) {
	if hasSuccessfulCall(calls) || !allCallsProviderFailoverEligible(calls) {
		return calls, 0
	}

	triedModel := make(map[string]bool, len(calls))
	triedProvider := make(map[string]bool, len(calls))
	for _, c := range calls {
		triedModel[c.ModelID] = true
		if c.ProviderID != "" {
			triedProvider[c.ProviderID] = true
		}
	}

	var fallback []string
	for _, r := range plan.Ranked {
		if triedModel[r.ModelID] || triedProvider[r.ProviderID] {
			continue
		}
		if entryByID(p.Registry, r.ModelID) == nil || p.Adapters[r.ProviderID] == nil {
			continue
		}
		fallback = append(fallback, r.ModelID)
		triedProvider[r.ProviderID] = true
		if len(fallback) >= cfg.MaxParallelCalls {
			break
		}
	}
	if len(fallback) == 0 {
		return calls, 0
	}

	more, retries, _ := p.invokeModels(ctx, req, subObjective, fallback, cfg)
	return append(calls, more...), retries
}

func hasSuccessfulCall(calls []ModelCallResult) bool {
	for _, c := range calls {
		if c.Error == "" && strings.TrimSpace(c.Text) != "" {
			return true
		}
	}
	return false
}

func allCallsProviderFailoverEligible(calls []ModelCallResult) bool {
	if len(calls) == 0 {
		return false
	}
	for _, c := range calls {
		if c.Error == "" {
			return false
		}
		if !isProviderFailoverError(c.Error) {
			return false
		}
	}
	return true
}

func isProviderFailoverError(msg string) bool {
	e := strings.ToLower(msg)
	return strings.Contains(e, "402") ||
		strings.Contains(e, "insufficient credits") ||
		strings.Contains(e, "payment required") ||
		strings.Contains(e, "401") ||
		strings.Contains(e, "403") ||
		strings.Contains(e, "429") ||
		strings.Contains(e, "invalid or revoked") ||
		strings.Contains(e, "provider rejected key") ||
		strings.Contains(e, "503") ||
		strings.Contains(e, "rate limit")
}
