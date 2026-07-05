package orchestration

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"gaiol/internal/orchestration/llm"
)

// Pipeline runs the orchestration loop.
type Pipeline struct {
	Trust    TrustRepository
	Traces   TraceRepository
	Registry []ModelRegistryEntry
	Adapters map[string]llm.ProviderAdapter
	Config   OrchestratorConfig
}

// Run executes orchestration for one request.
func (p *Pipeline) Run(ctx context.Context, req OrchestrationRequest, cfgOverride *OrchestratorConfig) (*OrchestrationResult, error) {
	cfg := p.Config
	if cfgOverride != nil {
		cfg = mergeConfig(cfg, *cfgOverride)
	}
	if cfg.BeamWidth < 1 {
		cfg.BeamWidth = 2
	}
	if cfg.MaxParallelCalls < 1 {
		cfg.MaxParallelCalls = 3
	}

	startedAt := nowRFC3339()
	decomposition := HeuristicDecompose(req)

	subtaskTraces := make([]SubtaskExecutionTrace, 0, len(decomposition.Subtasks))
	modelIDs := make([]string, 0, len(p.Registry))
	for _, e := range p.Registry {
		modelIDs = append(modelIDs, e.ModelID)
	}

	spentUsd := 0.0
	var trustUpdates []TrustUpdateEvent
	totalRetries := 0

	for _, sub := range decomposition.Subtasks {
		trustState, err := buildTrustMap(ctx, p.Trust, req.Domain, modelIDs)
		if err != nil {
			return nil, err
		}
		trustByModel := make(map[string]BetaTrust, len(trustState))
		for id, v := range trustState {
			if v.record != nil {
				trustByModel[id] = v.record.Distribution
			} else {
				trustByModel[id] = UniformPrior
			}
		}

		explore := false
		if req.ExplorePaths != nil {
			explore = *req.ExplorePaths
		}
		beamW := cfg.BeamWidth
		if explore && req.BeamWidth != nil && *req.BeamWidth > beamW {
			beamW = *req.BeamWidth
		}
		if !explore {
			beamW = 1
		}

		maxPar := cfg.MaxParallelCalls
		if req.Constraints != nil && req.Constraints.MaxParallelCalls != nil {
			if *req.Constraints.MaxParallelCalls < maxPar {
				maxPar = *req.Constraints.MaxParallelCalls
			}
		}

		rctx := RoutingContext{
			Domain:       req.Domain,
			TaskKind:     sub.TaskKind,
			Subtask:      sub,
			Registry:     p.Registry,
			TrustByModel: trustByModel,
		}
		plan := PlanSubtaskRouting(rctx, explore, beamW, maxPar)
		selected := append([]string(nil), plan.CandidateModelIDs...)
		if cfg.MaxCostUsdPerRequest != nil && spentUsd >= *cfg.MaxCostUsdPerRequest && len(selected) > 0 {
			selected = selected[:1]
		}

		calls, retries, err := p.invokeModels(ctx, req, sub.Description, selected, cfg)
		totalRetries += retries
		if err != nil {
			return nil, err
		}
		for _, c := range calls {
			if c.Usage != nil && c.Usage.CostUsd != nil {
				spentUsd += *c.Usage.CostUsd
			}
		}

		scoredPaths := ScorePaths(sub.Description, calls)
		kept, discarded := PruneBeam(scoredPaths, beamW)
		keptCalls := make([]ModelCallResult, len(kept))
		for i, path := range kept {
			keptCalls[i] = path.Result
		}

		scores := make(map[string]float64, len(scoredPaths))
		for _, sp := range scoredPaths {
			scores[sp.ModelID] = sp.Score
		}

		trustMeans := make(map[string]float64, len(trustState))
		trustRecords := make(map[string]BetaTrust, len(trustState))
		for id, v := range trustState {
			trustMeans[id] = v.mean
			dist := UniformPrior
			if v.record != nil {
				dist = v.record.Distribution
			}
			trustRecords[id] = dist
		}

		consensusScores := make(map[string]float64, len(kept))
		for _, path := range kept {
			consensusScores[path.ModelID] = path.Score
		}

		consensus := RunConsensus(ConsensusInput{
			Query:                 sub.Description,
			Mode:                  cfg.ConsensusMode,
			Domain:                req.Domain,
			Candidates:            keptCalls,
			Scores:                consensusScores,
			StaticWeights:         cfg.StaticWeights,
			TrustMeans:            trustMeansForMode(cfg.ConsensusMode, trustMeans),
			TrustRecords:          trustRecordsForMode(cfg.ConsensusMode, trustRecords),
			ABTCConsensusExponent: cfg.ABTC.ConsensusTrustExponent,
			Lambda:                1 - cfg.ABTC.Decay,
		})

		trustRound, events, err := p.processTrustRound(ctx, processTrustArgs{
			traceID:         req.TraceID,
			sessionHint:     req.SessionHint,
			domain:          req.Domain,
			subtaskID:       sub.ID,
			consensusMode:   cfg.ConsensusMode,
			winnerModelID:   consensus.ChosenModelID,
			keptCalls:       keptCalls,
			consensusScores: consensusScores,
			cfg:             cfg,
		})
		if err != nil {
			return nil, err
		}
		trustUpdates = append(trustUpdates, events...)

		winningPathID := PathIDForModel(consensus.ChosenModelID)
		candidates := make([]PathCandidateTrace, len(scoredPaths))
		keptSet := make(map[string]struct{}, len(kept))
		for _, k := range kept {
			keptSet[k.PathID] = struct{}{}
		}
		for i, sp := range scoredPaths {
			preview := sp.Result.Text
			if len(preview) > 200 {
				preview = preview[:200]
			}
			_, isKept := keptSet[sp.PathID]
			candidates[i] = PathCandidateTrace{
				PathID: sp.PathID, ModelID: sp.ModelID, ProviderID: sp.ProviderID,
				Score: sp.Score, Kept: isKept, TextPreview: preview,
			}
		}

		discardedIDs := make([]string, len(discarded))
		for i, d := range discarded {
			discardedIDs[i] = d.PathID
		}
		keptIDs := make([]string, len(kept))
		for i, k := range kept {
			keptIDs[i] = k.PathID
		}

		subtaskTraces = append(subtaskTraces, SubtaskExecutionTrace{
			SubtaskID:      sub.ID,
			RoutedModelIDs: selected,
			Calls:          calls,
			Scores:         scores,
			ChosenModelID:  consensus.ChosenModelID,
			ConsensusText:  consensus.Text,
			TrustRound:     trustRound,
			RoutingExplanation: &RoutingExplanationTrace{
				DiversityRationale: plan.DiversityExplanation,
				CandidatePoolSize:  plan.CandidatePoolSize,
				BeamWidth:          beamW,
				ModelRankSnapshot:  plan.ModelRankSnapshot,
			},
			PathExploration: &PathExplorationTrace{
				Candidates: candidates,
				Pruning: BeamPruneTrace{
					BeamWidth: beamW, KeptPathIDs: keptIDs, DiscardedPathIDs: discardedIDs,
				},
				WinningPathID: winningPathID,
			},
		})
	}

	answerParts := make([]string, 0, len(subtaskTraces))
	for _, st := range subtaskTraces {
		answerParts = append(answerParts, st.ConsensusText)
	}
	answer := strings.Join(answerParts, "\n\n")
	finishedAt := nowRFC3339()

	trace := OrchestrationTrace{
		TraceID:       req.TraceID,
		Domain:        req.Domain,
		Decomposition: decomposition,
		Subtasks:      subtaskTraces,
		StartedAt:     startedAt,
		FinishedAt:    finishedAt,
	}
	if p.Traces != nil {
		if err := p.Traces.Append(trace); err != nil {
			log.Printf("orchestration: trace append failed trace_id=%s: %v", req.TraceID, err)
		}
	}

	return &OrchestrationResult{
		Trace:        trace,
		Answer:       answer,
		TrustUpdates: trustUpdates,
		TotalRetries: totalRetries,
	}, nil
}

type trustStateEntry struct {
	record *TrustRecord
	mean   float64
}

func buildTrustMap(ctx context.Context, repo TrustRepository, domain string, modelIDs []string) (map[string]trustStateEntry, error) {
	_ = ctx
	out := make(map[string]trustStateEntry, len(modelIDs))
	if repo == nil {
		for _, id := range modelIDs {
			out[id] = trustStateEntry{mean: betaMean(UniformPrior)}
		}
		return out, nil
	}
	for _, id := range modelIDs {
		rec, err := repo.GetTrust(id, domain)
		if err != nil {
			return nil, err
		}
		dist := UniformPrior
		if rec != nil {
			dist = rec.Distribution
		}
		out[id] = trustStateEntry{record: rec, mean: betaMean(dist)}
	}
	return out, nil
}

type processTrustArgs struct {
	traceID, sessionHint, domain, subtaskID, consensusMode, winnerModelID string
	keptCalls                                                           []ModelCallResult
	consensusScores                                                     map[string]float64
	cfg                                                                 OrchestratorConfig
}

func (p *Pipeline) processTrustRound(ctx context.Context, args processTrustArgs) (*TrustRoundTrace, []TrustUpdateEvent, error) {
	decay := args.cfg.ABTC.Decay
	strengthWinner := args.cfg.ABTC.Strength
	strengthParticipant := args.cfg.ABTC.ParticipantStrength
	if strengthParticipant == 0 {
		strengthParticipant = strengthWinner * 0.6
	}
	persist := args.consensusMode == "abtc"
	var events []TrustUpdateEvent
	var entries []TrustRoundEntryTrace

	lambda := clamp01(1 - decay)

	for _, c := range args.keptCalls {
		if c.Error != "" {
			continue
		}
		role := "participant"
		isWinner := args.winnerModelID != "none" && c.ModelID == args.winnerModelID
		if isWinner {
			role = "winner"
		}

		var stored BetaTrust = UniformPrior
		if p.Trust != nil {
			existing, err := p.Trust.GetTrust(c.ModelID, args.domain)
			if err != nil {
				return nil, nil, err
			}
			if existing != nil {
				stored = existing.Distribution
			}
		}
		posterior := updateTrust(stored.Alpha, stored.Beta, isWinner, lambda)
		afterDecay := BetaTrust{Alpha: lambda * stored.Alpha, Beta: lambda * stored.Beta}
		priorMean, posteriorMean := betaMeanPair(stored, posterior)
		signal := 0.0
		if isWinner {
			signal = 1.0
		}
		explanation := fmt.Sprintf("binary: lambda=%.4f isWinner=%v alpha: %.4f->%.4f beta: %.4f->%.4f",
			lambda, isWinner, stored.Alpha, posterior.Alpha, stored.Beta, posterior.Beta)
		strength := strengthParticipant
		if isWinner {
			strength = strengthWinner
		}

		entries = append(entries, TrustRoundEntryTrace{
			ModelID: c.ModelID, ProviderID: c.ProviderID, Domain: args.domain, Role: role,
			Prior: stored, AfterDecay: afterDecay, Posterior: posterior,
			PriorMean: priorMean, PosteriorMean: posteriorMean, Decay: decay, Strength: strength,
			Signal: signal, Explanation: explanation, Persisted: persist,
		})

		if persist && p.Trust != nil {
			updatedAt := trustNow()
			rec := TrustRecord{ModelID: c.ModelID, Domain: args.domain, Distribution: posterior, UpdatedAt: updatedAt}
			if err := p.Trust.UpsertTrust(rec); err != nil {
				return nil, nil, err
			}
			events = append(events, TrustUpdateEvent{
				TraceID: args.traceID, SessionHint: args.sessionHint, Domain: args.domain,
				ModelID: c.ModelID, ProviderID: c.ProviderID, Distribution: posterior, UpdatedAt: updatedAt,
				SubtaskID: args.subtaskID, PriorDistribution: stored, AfterDecayDistribution: afterDecay,
				PriorMean: priorMean, PosteriorMean: posteriorMean, Decay: decay, Strength: strength,
				Signal: signal, Role: role, Explanation: explanation,
			})
		}
	}

	exp := args.cfg.ABTC.ConsensusTrustExponent
	tr := &TrustRoundTrace{
		ConsensusMode:          args.consensusMode,
		WinnerModelID:          args.winnerModelID,
		SubtaskID:              args.subtaskID,
		Decay:                  decay,
		StrengthWinner:         strengthWinner,
		StrengthParticipant:    strengthParticipant,
		ConsensusTrustExponent: &exp,
		Entries:                entries,
	}
	return tr, events, nil
}

func (p *Pipeline) invokeModels(ctx context.Context, req OrchestrationRequest, subObjective string, selected []string, cfg OrchestratorConfig) ([]ModelCallResult, int, error) {
	messages := append([]ChatMessage(nil), req.Messages...)
	messages = append(messages, ChatMessage{Role: "user", Content: "Subtask: " + subObjective})

	type result struct {
		call    ModelCallResult
		retries int
	}
	results := make([]result, len(selected))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, modelID := range selected {
		wg.Add(1)
		go func(idx int, modelID string) {
			defer wg.Done()
			call, retries := p.invokeOne(ctx, req, messages, modelID, cfg)
			mu.Lock()
			results[idx] = result{call: call, retries: retries}
			mu.Unlock()
		}(i, modelID)
	}
	wg.Wait()

	out := make([]ModelCallResult, len(results))
	retries := 0
	for i, r := range results {
		out[i] = r.call
		retries += r.retries
	}
	return out, retries, nil
}

func (p *Pipeline) invokeOne(ctx context.Context, req OrchestrationRequest, messages []ChatMessage, modelID string, cfg OrchestratorConfig) (ModelCallResult, int) {
	entry := entryByID(p.Registry, modelID)
	if entry == nil {
		return ModelCallResult{ModelID: modelID, ProviderID: "unknown", Error: "unknown_model"}, 0
	}
	adapter := p.Adapters[entry.ProviderID]
	if adapter == nil {
		return ModelCallResult{ModelID: modelID, ProviderID: entry.ProviderID, Error: "no_adapter"}, 0
	}

	llmMsgs := make([]llm.ChatMessage, len(messages))
	for i, m := range messages {
		llmMsgs[i] = llm.ChatMessage{Role: m.Role, Content: m.Content}
	}
	params := llm.GenerateParams{TraceID: req.TraceID, Model: entry.RemoteName, Messages: llmMsgs}
	if req.Constraints != nil {
		params.Temperature = req.Constraints.Temperature
		params.MaxOutputTokens = req.Constraints.MaxOutputTokens
	}

	retries := 0
	var gen llm.GenerateResult
	err := withRetry(ctx, func() error {
		var e error
		gen, e = adapter.Generate(ctx, params)
		return e
	}, RetryOptions{
		Retries:     cfg.Retry.Retries,
		BaseDelayMs: cfg.Retry.BaseDelayMs,
		OnRetry: func(attempt, maxAttempts, delayMs int, err error) {
			retries++
		},
	})
	if err != nil {
		return ModelCallResult{
			ModelID: modelID, ProviderID: entry.ProviderID, Text: "", LatencyMs: 0,
			Error: fmt.Sprintf("%s: %s", entry.ProviderID, err.Error()),
		}, retries
	}

	var usage *ModelCallUsage
	if gen.Usage != nil {
		cost := gen.Usage.CostUsd
		pt, ct := gen.Usage.PromptTokens, gen.Usage.CompletionTokens
		usage = &ModelCallUsage{PromptTokens: &pt, CompletionTokens: &ct, CostUsd: &cost}
	}
	return ModelCallResult{
		ModelID: modelID, ProviderID: entry.ProviderID, Text: gen.Text, LatencyMs: gen.LatencyMs, Usage: usage,
	}, retries
}

func mergeConfig(base, override OrchestratorConfig) OrchestratorConfig {
	out := base
	if override.ConsensusMode != "" {
		out.ConsensusMode = override.ConsensusMode
	}
	if override.BeamWidth > 0 {
		out.BeamWidth = override.BeamWidth
	}
	if override.MaxParallelCalls > 0 {
		out.MaxParallelCalls = override.MaxParallelCalls
	}
	if override.MaxCostUsdPerRequest != nil {
		out.MaxCostUsdPerRequest = override.MaxCostUsdPerRequest
	}
	if override.ABTC.Decay > 0 || override.ABTC.Strength > 0 || override.ABTC.ParticipantStrength > 0 || override.ABTC.ConsensusTrustExponent > 0 {
		abtc := out.ABTC
		if override.ABTC.Decay > 0 {
			abtc.Decay = override.ABTC.Decay
		}
		if override.ABTC.Strength > 0 {
			abtc.Strength = override.ABTC.Strength
		}
		if override.ABTC.ParticipantStrength > 0 {
			abtc.ParticipantStrength = override.ABTC.ParticipantStrength
		}
		if override.ABTC.ConsensusTrustExponent > 0 {
			abtc.ConsensusTrustExponent = override.ABTC.ConsensusTrustExponent
		}
		out.ABTC = abtc
	}
	if override.Retry.Retries > 0 || override.Retry.BaseDelayMs > 0 {
		out.Retry = override.Retry
	}
	if len(override.StaticWeights) > 0 {
		out.StaticWeights = override.StaticWeights
	}
	return out
}

func trustMeansForMode(mode string, means map[string]float64) map[string]float64 {
	if mode != "abtc" {
		return nil
	}
	return means
}

func trustRecordsForMode(mode string, records map[string]BetaTrust) map[string]BetaTrust {
	if mode != "abtc" {
		return nil
	}
	return records
}
