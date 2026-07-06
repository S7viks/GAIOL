package orchestration

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"gaiol/internal/orchestration/llm"
)

type errorAdapter struct {
	provider string
	errText  string
}

func (e *errorAdapter) ProviderID() string { return e.provider }

func (e *errorAdapter) Generate(context.Context, llm.GenerateParams) (llm.GenerateResult, error) {
	return llm.GenerateResult{}, fmt.Errorf("%s", e.errText)
}

func TestExpandCallsOnProviderFailure_FallsBackToOtherProvider(t *testing.T) {
	acc := 0.7
	reg := []ModelRegistryEntry{
		{ModelID: "openrouter:claude-opus-4", ProviderID: "openrouter", RemoteName: "anthropic/claude-opus-4", Capabilities: []string{"general"}, CostIndex: 0.5, LatencyPriorMs: 800, AccuracyPrior: &acc, Available: true},
		{ModelID: "google:gemini-2.5-pro", ProviderID: "google", RemoteName: "gemini-2.5-pro", Capabilities: []string{"general"}, CostIndex: 0.3, LatencyPriorMs: 700, AccuracyPrior: &acc, Available: true},
		{ModelID: "groq:llama-3.3-70b", ProviderID: "groq", RemoteName: "llama-3.3-70b-versatile", Capabilities: []string{"general"}, CostIndex: 0.2, LatencyPriorMs: 400, AccuracyPrior: &acc, Available: true},
	}
	p := &Pipeline{
		Registry: reg,
		Adapters: map[string]llm.ProviderAdapter{
			"openrouter": &errorAdapter{provider: "openrouter", errText: "openrouter: provider error (402): Insufficient credits"},
			"google":     llm.NewMockAdapter("google"),
			"groq":       llm.NewMockAdapter("groq"),
		},
		Config: DefaultConfig(),
	}

	plan := RoutingPlan{
		CandidateModelIDs: []string{"openrouter:claude-opus-4"},
		Ranked: []RankedModel{
			{ModelID: "openrouter:claude-opus-4", ProviderID: "openrouter", Score: 0.9},
			{ModelID: "google:gemini-2.5-pro", ProviderID: "google", Score: 0.8},
			{ModelID: "groq:llama-3.3-70b", ProviderID: "groq", Score: 0.7},
		},
	}
	initial := []ModelCallResult{{
		ModelID: "openrouter:claude-opus-4", ProviderID: "openrouter",
		Error: "openrouter: provider error (402): Insufficient credits",
	}}

	expanded, extraRetries := p.expandCallsOnProviderFailure(context.Background(), OrchestrationRequest{}, "test", p.Config, plan, initial)
	if extraRetries != 0 {
		t.Fatalf("extraRetries=%d", extraRetries)
	}
	if !hasSuccessfulCall(expanded) {
		t.Fatalf("expected fallback success, got %+v", expanded)
	}
	for _, c := range expanded {
		if c.Error == "" && c.ProviderID == "openrouter" {
			t.Fatalf("should not succeed on openrouter: %+v", c)
		}
	}
}

func TestRunConsensus_UsesSuccessfulProviderAfterFailover(t *testing.T) {
	out := RunConsensus(ConsensusInput{
		Mode: "uniform",
		Candidates: []ModelCallResult{
			{ModelID: "openrouter:claude-opus-4", ProviderID: "openrouter", Error: "openrouter: provider error (402): Insufficient credits"},
			{ModelID: "google:gemini-2.5-pro", ProviderID: "google", Text: "A Dyson sphere is a hypothetical megastructure."},
		},
		Scores: map[string]float64{
			"openrouter:claude-opus-4": 0.9,
			"google:gemini-2.5-pro":    0.8,
		},
	})
	if !strings.Contains(out.Text, "Dyson sphere") {
		t.Fatalf("answer=%q", out.Text)
	}
	if out.ChosenModelID != "google:gemini-2.5-pro" {
		t.Fatalf("chosen=%s", out.ChosenModelID)
	}
}
