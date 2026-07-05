package orchestration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// goldenExpectations documents structural invariants for the mock beam pipeline.
type goldenExpectations struct {
	MinSubtasks        int      `json:"min_subtasks"`
	MinPathCandidates  int      `json:"min_path_candidates"`
	MinRoutedModels    int      `json:"min_routed_models"`
	MinTrustUpdates    int      `json:"min_trust_updates"`
	ConsensusMode      string   `json:"consensus_mode"`
	RequiredModelIDs   []string `json:"required_model_ids_subset"`
}

func TestGolden_BeamMockPipeline(t *testing.T) {
	root := moduleRoot(t)
	fixture := filepath.Join(root, "internal", "orchestration", "testdata", "golden_beam_expectations.json")
	raw, err := os.ReadFile(fixture)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var exp goldenExpectations
	if err := json.Unmarshal(raw, &exp); err != nil {
		t.Fatal(err)
	}

	p := mockPipeline()
	explore := true
	req := OrchestrationRequest{
		TraceID:      "golden-trace-beam",
		Domain:       "general",
		TaskKind:     "qa",
		Objective:    "integration alpha beta gamma objective for beam routing",
		ExplorePaths: &explore,
		Messages:     []ChatMessage{{Role: "user", Content: reqObjective}},
	}
	result, err := p.Run(context.Background(), req, nil)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Trace.Subtasks) < exp.MinSubtasks {
		t.Fatalf("subtasks=%d want>=%d", len(result.Trace.Subtasks), exp.MinSubtasks)
	}
	if len(result.TrustUpdates) < exp.MinTrustUpdates {
		t.Fatalf("trust updates=%d want>=%d", len(result.TrustUpdates), exp.MinTrustUpdates)
	}

	st := result.Trace.Subtasks[0]
	if st.PathExploration == nil {
		t.Fatal("missing path_exploration")
	}
	if len(st.PathExploration.Candidates) < exp.MinPathCandidates {
		t.Fatalf("candidates=%d want>=%d", len(st.PathExploration.Candidates), exp.MinPathCandidates)
	}
	if len(st.RoutedModelIDs) < exp.MinRoutedModels {
		t.Fatalf("routed=%v", st.RoutedModelIDs)
	}

	routedSet := make(map[string]struct{}, len(st.RoutedModelIDs))
	for _, id := range st.RoutedModelIDs {
		routedSet[id] = struct{}{}
	}
	for _, id := range exp.RequiredModelIDs {
		if _, ok := routedSet[id]; !ok {
			t.Fatalf("missing routed model %q in %v", id, st.RoutedModelIDs)
		}
	}

	v1 := ToOrchestrateResponseV1(result, "")
	if v1.SchemaVersion != "1.0" || v1.TraceID != req.TraceID {
		t.Fatalf("wire mapping: %+v", v1)
	}
}

const reqObjective = "integration alpha beta gamma objective for beam routing"

func moduleRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
