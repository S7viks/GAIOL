package orchestration

import (
	"testing"

	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
)

func TestConfigOverrideFromV1_Nil(t *testing.T) {
	t.Parallel()
	if ConfigOverrideFromV1(nil) != nil {
		t.Fatal("expected nil")
	}
}

func TestConfigOverrideFromV1_ConsensusMode(t *testing.T) {
	t.Parallel()
	mode := "uniform"
	got := ConfigOverrideFromV1(&orchestratorv1.OrchestrateRequestV1{ConsensusMode: mode})
	if got == nil || got.ConsensusMode != mode {
		t.Fatalf("got %+v", got)
	}
}

func TestConfigOverrideFromV1_BeamWidth(t *testing.T) {
	t.Parallel()
	bw := 4
	got := ConfigOverrideFromV1(&orchestratorv1.OrchestrateRequestV1{BeamWidth: &bw})
	if got == nil || got.BeamWidth != 4 {
		t.Fatalf("got %+v", got)
	}
}

func TestConfigOverrideFromV1_AbtcDecay(t *testing.T) {
	t.Parallel()
	decay := 0.02
	got := ConfigOverrideFromV1(&orchestratorv1.OrchestrateRequestV1{AbtcDecay: &decay})
	if got == nil || got.ABTC.Decay != 0.02 {
		t.Fatalf("got %+v", got)
	}
	if got.ABTC.Strength != 1.5 {
		t.Fatalf("strength=%v", got.ABTC.Strength)
	}
}

func TestMergeConfigPartialABTC(t *testing.T) {
	t.Parallel()
	base := DefaultConfig()
	override := OrchestratorConfig{ABTC: ABTCConfig{Decay: 0.02}}
	merged := mergeConfig(base, override)
	if merged.ABTC.Decay != 0.02 {
		t.Fatalf("decay=%v", merged.ABTC.Decay)
	}
	if merged.ABTC.Strength != base.ABTC.Strength {
		t.Fatalf("strength overwritten: %v", merged.ABTC.Strength)
	}
}
