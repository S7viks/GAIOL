package httpserver

import (
	"context"
	"testing"
)

func TestResolveOrchestrationPrefs_NoDBUsesEnv(t *testing.T) {
	t.Setenv("GAIOL_BEAM_WIDTH", "4")
	t.Setenv("GAIOL_CONSENSUS_MODE", "uniform")
	t.Setenv("GAIOL_DOMAIN", "reasoning")
	t.Setenv("GAIOL_EXPLORE_PATHS", "1")

	got := (&Deps{}).resolveOrchestrationPrefs(context.Background(), "tenant-1")
	if got.BeamWidth != 4 {
		t.Fatalf("beam_width = %d want 4", got.BeamWidth)
	}
	if got.ConsensusMode != "uniform" {
		t.Fatalf("consensus = %q want uniform", got.ConsensusMode)
	}
	if got.Domain != "reasoning" {
		t.Fatalf("domain = %q want reasoning", got.Domain)
	}
	if !got.ExplorePaths {
		t.Fatal("explore_paths want true")
	}
}

func TestPreferencesResponseFromSettings_UsesEnvWhenUnset(t *testing.T) {
	t.Setenv("GAIOL_BEAM_WIDTH", "3")
	t.Setenv("GAIOL_CONSENSUS_MODE", "static")
	t.Setenv("GAIOL_DOMAIN", "general")
	t.Setenv("GAIOL_EXPLORE_PATHS", "1")

	out := preferencesResponseFromSettings(nil)
	if out["beam_width"] != 3 {
		t.Fatalf("beam_width = %v want 3", out["beam_width"])
	}
	if out["consensus_mode"] != "static" {
		t.Fatalf("consensus_mode = %v want static", out["consensus_mode"])
	}
	if out["explore_paths"] != true {
		t.Fatalf("explore_paths = %v want true", out["explore_paths"])
	}
}

func TestNormalizeConsensusMode(t *testing.T) {
	if normalizeConsensusMode("ABTC") != "abtc" {
		t.Fatal("expected abtc")
	}
	if normalizeConsensusMode("nope") != "" {
		t.Fatal("expected empty for invalid mode")
	}
}
