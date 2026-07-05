package orchestration

import "testing"

func TestEvaluateAgainstContains_CaseInsensitive(t *testing.T) {
	ex := EvalExample{
		Objective:        "Greet the user",
		ExpectedContains: []string{"hello"},
	}
	got := EvaluateAgainstContains(ex, "Hello there!")
	if !got.Pass {
		t.Fatalf("expected pass, got %+v", got)
	}
}

func TestEvaluateAgainstContains_MissingSubstring(t *testing.T) {
	ex := EvalExample{
		Objective:        "Greet the user",
		ExpectedContains: []string{"hello", "hi"},
	}
	got := EvaluateAgainstContains(ex, "Hello there!")
	if got.Pass {
		t.Fatalf("expected fail (hi not in answer), got %+v", got)
	}
	if got.Notes != "missing: hi" {
		t.Fatalf("notes = %q", got.Notes)
	}
}
