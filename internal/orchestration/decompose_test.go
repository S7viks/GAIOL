package orchestration

import (
	"strings"
	"testing"
)

func TestHeuristicDecomposeSingleSentence(t *testing.T) {
	t.Parallel()
	r := HeuristicDecompose(OrchestrationRequest{
		TraceID:  "t1",
		Domain:   "general",
		TaskKind: "qa",
		Objective: "What is 2+2?",
	})
	if len(r.Subtasks) != 1 {
		t.Fatalf("subtasks=%d", len(r.Subtasks))
	}
	if !strings.Contains(r.Subtasks[0].Description, "2+2") {
		t.Fatalf("description=%q", r.Subtasks[0].Description)
	}
}

func TestHeuristicDecomposeMultiSentence(t *testing.T) {
	t.Parallel()
	r := HeuristicDecompose(OrchestrationRequest{
		TraceID:  "t2",
		Domain:   "general",
		TaskKind: "reasoning",
		Objective: "First step. Second step! Third step?",
	})
	if len(r.Subtasks) < 2 {
		t.Fatalf("expected >=2 subtasks, got %d", len(r.Subtasks))
	}
}

func TestSplitSentences(t *testing.T) {
	t.Parallel()
	got := splitSentences("Alpha. Beta! Gamma?")
	if len(got) != 3 {
		t.Fatalf("sentences=%v", got)
	}
}
