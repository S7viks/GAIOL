package orchestration

import orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"

// ConfigOverrideFromV1 builds per-request config overrides from the wire envelope.
// Mirrors orchestrator/src/contract/v1/map.ts (consensusModeV1ToConfigPartial, abtcDecayV1ToConfigPartial)
// and orchestrator/src/orchestration/config-merge.ts.
func ConfigOverrideFromV1(req *orchestratorv1.OrchestrateRequestV1) *OrchestratorConfig {
	if req == nil {
		return nil
	}
	var partial OrchestratorConfig
	has := false

	if req.ConsensusMode != "" {
		partial.ConsensusMode = req.ConsensusMode
		has = true
	}
	if req.BeamWidth != nil && *req.BeamWidth > 0 {
		partial.BeamWidth = *req.BeamWidth
		has = true
	}
	if req.Constraints != nil {
		if req.Constraints.MaxParallelCalls != nil && *req.Constraints.MaxParallelCalls > 0 {
			partial.MaxParallelCalls = *req.Constraints.MaxParallelCalls
			has = true
		}
		if req.Constraints.MaxCostUsd != nil {
			partial.MaxCostUsdPerRequest = req.Constraints.MaxCostUsd
			has = true
		}
	}
	if req.AbtcDecay != nil && isFiniteFloat(*req.AbtcDecay) {
		clamped := clamp01(*req.AbtcDecay)
		if clamped > 0.99 {
			clamped = 0.99
		}
		partial.ABTC = ABTCConfig{
			Decay:               clamped,
			Strength:            1.5,
			ParticipantStrength: 0.9,
		}
		has = true
	}
	if !has {
		return nil
	}
	return &partial
}

func isFiniteFloat(v float64) bool {
	return !((v != v) || v > 1e308 || v < -1e308)
}
