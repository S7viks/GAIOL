package orchestration

import (
	"context"
	"testing"

	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
	"gaiol/internal/orchestration/llm"
)

func mockPipeline() *Pipeline {
	reg := sampleRegistry()
	return &Pipeline{
		Trust:    NewMemoryTrustStore(),
		Traces:   NewMemoryTraceStore(),
		Registry: reg,
		Adapters: mockAdapters(),
		Config:   DefaultConfig(),
	}
}

func TestPipeline_BeamMockRun(t *testing.T) {
	p := mockPipeline()
	explore := true
	req := OrchestrationRequest{
		TraceID:      "golden-trace-001",
		Domain:       "general",
		TaskKind:     "qa",
		Objective:    "integration alpha beta gamma objective for beam routing",
		ExplorePaths: &explore,
		Messages:     []ChatMessage{{Role: "user", Content: "integration alpha beta gamma objective for beam routing"}},
	}
	result, err := p.Run(context.Background(), req, nil)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil || len(result.Trace.Subtasks) == 0 {
		t.Fatal("expected subtasks")
	}
	st := result.Trace.Subtasks[0]
	if st.PathExploration == nil {
		t.Fatal("expected path exploration")
	}
	if len(st.PathExploration.Candidates) < 2 {
		t.Fatalf("candidates=%d", len(st.PathExploration.Candidates))
	}
	if len(st.RoutedModelIDs) < 2 {
		t.Fatalf("routed=%v", st.RoutedModelIDs)
	}
	if len(result.TrustUpdates) == 0 {
		t.Fatal("expected trust updates in abtc mode")
	}
}

func TestRunConsensus_NoSuccessfulCandidates(t *testing.T) {
	t.Parallel()
	out := RunConsensus(ConsensusInput{
		Mode: "uniform",
		Candidates: []ModelCallResult{
			{ModelID: "m1", ProviderID: "p", Error: "fail"},
		},
		Scores: map[string]float64{"m1": 0},
	})
	if out.ChosenModelID != "m1" {
		t.Fatalf("chosen=%s", out.ChosenModelID)
	}
}

func TestService_OrchestrateMock(t *testing.T) {
	svc := NewService()
	explore := true
	bw := 2
	req := &orchestratorv1.OrchestrateRequestV1{
		SchemaVersion: "1.0",
		TraceID:       "svc-trace-001",
		Domain:        "general",
		TaskKind:      "qa",
		Objective:     "short prompt for diversity",
		Messages:      []orchestratorv1.ChatMessageV1{{Role: "user", Content: "short prompt for diversity"}},
		ExplorePaths:  &explore,
		BeamWidth:     &bw,
		ConsensusMode: "abtc",
	}
	res, err := svc.Orchestrate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if err := orchestratorv1.ValidateOrchestrateResponseV1(res); err != nil {
		t.Fatal(err)
	}
	if res.Answer == "" {
		t.Fatal("empty answer")
	}
}

func TestMockAdapterGenerate(t *testing.T) {
	t.Parallel()
	m := llm.NewMockAdapter("mock")
	res, err := m.Generate(context.Background(), llm.GenerateParams{
		Model:    "mock-fast",
		Messages: []llm.ChatMessage{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Text == "" {
		t.Fatal("empty text")
	}
}
