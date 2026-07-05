package orchestration

import "testing"

func TestPathIDForModelStable(t *testing.T) {
	t.Parallel()
	if got := PathIDForModel("m1"); got != "path:m1" {
		t.Fatalf("got %q", got)
	}
}

func TestScorePathsSortsByScoreDesc(t *testing.T) {
	t.Parallel()
	calls := []ModelCallResult{
		{ModelID: "z", ProviderID: "p", Text: "objective match"},
		{ModelID: "a", ProviderID: "p", Text: "other"},
	}
	scored := ScorePaths("objective text", calls)
	if len(scored) != 2 {
		t.Fatalf("len=%d", len(scored))
	}
	if scored[0].ModelID != "z" {
		t.Fatalf("expected z first, got %s", scored[0].ModelID)
	}
}

func TestPruneBeamKeepsTopK(t *testing.T) {
	t.Parallel()
	calls := []ModelCallResult{
		{ModelID: "m1", ProviderID: "p", Text: "obj"},
		{ModelID: "m2", ProviderID: "p", Text: "obj long"},
		{ModelID: "m3", ProviderID: "p", Text: "x"},
	}
	paths := ScorePaths("obj", calls)
	kept, discarded := PruneBeam(paths, 2)
	if len(kept) != 2 || len(discarded) != 1 {
		t.Fatalf("kept=%d discarded=%d", len(kept), len(discarded))
	}
	keptIDs := make(map[string]struct{}, len(kept))
	for _, k := range kept {
		keptIDs[k.PathID] = struct{}{}
	}
	for _, d := range discarded {
		if _, ok := keptIDs[d.PathID]; ok {
			t.Fatalf("discarded path %s also kept", d.PathID)
		}
	}
}
