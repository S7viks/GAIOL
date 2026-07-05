package orchestration

import orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"

// ToOrchestrateResponseV1 maps pipeline result to wire response.
func ToOrchestrateResponseV1(result *OrchestrationResult, sessionID string) *orchestratorv1.OrchestrateResponseV1 {
	traceV1 := TraceToV1(result.Trace)
	updates := make([]orchestratorv1.TrustUpdateEventV1, len(result.TrustUpdates))
	for i, e := range result.TrustUpdates {
		updates[i] = trustUpdateToV1(e)
	}
	return &orchestratorv1.OrchestrateResponseV1{
		SchemaVersion: "1.0",
		TraceID:       result.Trace.TraceID,
		SessionID:     sessionID,
		Answer:        result.Answer,
		Trace:         traceV1,
		TrustUpdates:  updates,
	}
}

// TraceToV1 converts domain trace to wire trace.
func TraceToV1(t OrchestrationTrace) orchestratorv1.OrchestrationTraceV1 {
	subtasks := make([]orchestratorv1.SubtaskTraceV1, len(t.Subtasks))
	for i, s := range t.Subtasks {
		subtasks[i] = subtaskTraceToV1(s)
	}
	specs := make([]orchestratorv1.SubtaskSpecV1, len(t.Decomposition.Subtasks))
	for i, sp := range t.Decomposition.Subtasks {
		specs[i] = orchestratorv1.SubtaskSpecV1{
			ID: sp.ID, ParentID: sp.ParentID, Title: sp.Title, Description: sp.Description,
			TaskKind: sp.TaskKind, RequiredCapabilities: sp.RequiredCapabilities,
		}
	}
	return orchestratorv1.OrchestrationTraceV1{
		TraceID: t.TraceID, Domain: t.Domain, StartedAt: t.StartedAt, FinishedAt: t.FinishedAt,
		Decomposition: orchestratorv1.DecompositionV1{Subtasks: specs, Rationale: t.Decomposition.Rationale},
		Subtasks:      subtasks,
	}
}

func subtaskTraceToV1(s SubtaskExecutionTrace) orchestratorv1.SubtaskTraceV1 {
	calls := make([]orchestratorv1.ModelCallV1, len(s.Calls))
	for i, c := range s.Calls {
		calls[i] = modelCallToV1(c)
	}
	out := orchestratorv1.SubtaskTraceV1{
		SubtaskID: s.SubtaskID, RoutedModelIDs: s.RoutedModelIDs, Calls: calls,
		Scores: s.Scores, ChosenModelID: s.ChosenModelID, ConsensusText: s.ConsensusText,
	}
	if s.RoutingExplanation != nil {
		re := s.RoutingExplanation
		snap := make([]orchestratorv1.RoutingRankRowV1, len(re.ModelRankSnapshot))
		for i, row := range re.ModelRankSnapshot {
			snap[i] = orchestratorv1.RoutingRankRowV1{
				ModelID: row.ModelID, ProviderID: row.ProviderID, RoutingScore: row.Score,
				Breakdown: orchestratorv1.RoutingRankBreakdownV1{
					Accuracy: row.Breakdown.Accuracy, Latency: row.Breakdown.Latency,
					Cost: row.Breakdown.Cost, Availability: row.Breakdown.Availability,
					CapabilityMatch: row.Breakdown.CapabilityMatch,
				},
			}
		}
		out.RoutingExplanation = &orchestratorv1.RoutingExplanationTraceV1{
			DiversityRationale: re.DiversityRationale, CandidatePoolSize: re.CandidatePoolSize,
			BeamWidth: re.BeamWidth, ModelRankSnapshot: snap,
		}
	}
	if s.PathExploration != nil {
		pe := s.PathExploration
		cands := make([]orchestratorv1.PathCandidateTraceV1, len(pe.Candidates))
		for i, c := range pe.Candidates {
			cands[i] = orchestratorv1.PathCandidateTraceV1{
				PathID: c.PathID, ModelID: c.ModelID, ProviderID: c.ProviderID,
				Score: c.Score, Kept: c.Kept, TextPreview: c.TextPreview,
			}
		}
		out.PathExploration = &orchestratorv1.PathExplorationTraceV1{
			Candidates: cands,
			Pruning: orchestratorv1.BeamPruneTraceV1{
				BeamWidth: pe.Pruning.BeamWidth, KeptPathIDs: pe.Pruning.KeptPathIDs,
				DiscardedPathIDs: pe.Pruning.DiscardedPathIDs,
			},
			WinningPathID: pe.WinningPathID,
		}
	}
	if s.TrustRound != nil {
		tr := s.TrustRound
		entries := make([]orchestratorv1.TrustRoundEntryTraceV1, len(tr.Entries))
		for i, e := range tr.Entries {
			entries[i] = orchestratorv1.TrustRoundEntryTraceV1{
				ModelID: e.ModelID, ProviderID: e.ProviderID, Domain: e.Domain, Role: e.Role,
				Prior: betaToV1(e.Prior), AfterDecay: betaToV1(e.AfterDecay), Posterior: betaToV1(e.Posterior),
				PriorMean: e.PriorMean, PosteriorMean: e.PosteriorMean, Decay: e.Decay,
				Strength: e.Strength, Signal: e.Signal, Explanation: e.Explanation, Persisted: e.Persisted,
			}
		}
		out.TrustRound = &orchestratorv1.TrustRoundTraceV1{
			ConsensusMode: tr.ConsensusMode, WinnerModelID: tr.WinnerModelID, SubtaskID: tr.SubtaskID,
			Decay: tr.Decay, StrengthWinner: tr.StrengthWinner, StrengthParticipant: tr.StrengthParticipant,
			ConsensusTrustExponent: tr.ConsensusTrustExponent, Entries: entries,
		}
	}
	return out
}

func modelCallToV1(c ModelCallResult) orchestratorv1.ModelCallV1 {
	out := orchestratorv1.ModelCallV1{
		ModelID: c.ModelID, ProviderID: c.ProviderID, Text: c.Text, LatencyMs: c.LatencyMs, Error: c.Error,
	}
	if c.Usage != nil {
		out.Usage = &orchestratorv1.ModelCallUsageV1{
			PromptTokens: c.Usage.PromptTokens, CompletionTokens: c.Usage.CompletionTokens, CostUsd: c.Usage.CostUsd,
		}
	}
	return out
}

func betaToV1(b BetaTrust) orchestratorv1.BetaDistributionV1 {
	return orchestratorv1.BetaDistributionV1{Alpha: b.Alpha, Beta: b.Beta}
}

func trustUpdateToV1(e TrustUpdateEvent) orchestratorv1.TrustUpdateEventV1 {
	out := orchestratorv1.TrustUpdateEventV1{
		SchemaVersion: "1.0", Event: "trust_updated", TraceID: e.TraceID, SessionID: e.SessionHint,
		Domain: e.Domain, ModelID: e.ModelID, ProviderID: e.ProviderID,
		Distribution: betaToV1(e.Distribution), UpdatedAt: e.UpdatedAt, SubtaskID: e.SubtaskID,
		PriorDistribution:      ptrBetaToV1(e.PriorDistribution),
		AfterDecayDistribution: ptrBetaToV1(e.AfterDecayDistribution),
		PriorMean:              &e.PriorMean,
		PosteriorMean:          &e.PosteriorMean,
		Decay:                  &e.Decay,
		Strength:               &e.Strength,
		Signal:                 &e.Signal,
		Role:                   e.Role,
		Explanation:            e.Explanation,
	}
	return out
}

func ptrBetaToV1(b BetaTrust) *orchestratorv1.BetaDistributionV1 {
	v := betaToV1(b)
	return &v
}
