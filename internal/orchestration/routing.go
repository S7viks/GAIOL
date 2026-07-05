package orchestration

import "sort"

// RankedModel is a registry entry with routing score.
type RankedModel struct {
	ModelID    string
	ProviderID string
	Score      float64
	Breakdown  RoutingBreakdown
}

// RoutingBreakdown is the fitness decomposition.
type RoutingBreakdown struct {
	Accuracy        float64
	Latency         float64
	Cost            float64
	Availability    float64
	CapabilityMatch float64
}

// RoutingContext is input to routing planners.
type RoutingContext struct {
	Domain       string
	TaskKind     string
	Subtask      SubtaskSpec
	Registry     []ModelRegistryEntry
	TrustByModel map[string]BetaTrust
}

// RoutingPlan is the output of PlanSubtaskRouting.
type RoutingPlan struct {
	CandidateModelIDs   []string
	Ranked              []RankedModel
	DiversityExplanation string
	CandidatePoolSize   int
	BeamWidth           int
	ModelRankSnapshot   []RankedModel
}

func invNorm(x, max float64) float64 {
	if max <= 0 {
		return 1
	}
	return clamp01(1 - x/max)
}

func capabilityMatch(entry ModelRegistryEntry, sub SubtaskSpec) float64 {
	req := sub.RequiredCapabilities
	if len(req) == 0 {
		return 1
	}
	caps := make(map[string]struct{}, len(entry.Capabilities))
	for _, c := range entry.Capabilities {
		caps[c] = struct{}{}
	}
	hit := 0
	for _, r := range req {
		if _, ok := caps[r]; ok {
			hit++
		}
	}
	return float64(hit) / float64(len(req))
}

func scoreModel(ctx RoutingContext, entry ModelRegistryEntry) RankedModel {
	trust := ctx.TrustByModel[entry.ModelID]
	if trust.Alpha == 0 && trust.Beta == 0 {
		trust = UniformPrior
	}
	histAcc := 0.5
	if entry.AccuracyPrior != nil {
		histAcc = clamp01(*entry.AccuracyPrior)
	} else {
		histAcc = clamp01(betaMean(trust))
	}

	maxCost := 1e-6
	maxLat := 1.0
	for _, e := range ctx.Registry {
		if e.CostIndex > maxCost {
			maxCost = e.CostIndex
		}
		if float64(e.LatencyPriorMs) > maxLat {
			maxLat = float64(e.LatencyPriorMs)
		}
	}

	costEff := invNorm(entry.CostIndex, maxCost)
	lat := invNorm(float64(entry.LatencyPriorMs), maxLat)
	avail := 0.0
	if entry.Available {
		avail = 1
	}
	cap := capabilityMatch(entry, ctx.Subtask)

	fitness := WCap*cap + WHistAcc*histAcc + WCostEff*costEff
	latencyModifier := 1 + 0.05*lat
	score := clamp01(fitness) * latencyModifier * avail

	return RankedModel{
		ModelID:    entry.ModelID,
		ProviderID: entry.ProviderID,
		Score:      score,
		Breakdown: RoutingBreakdown{
			Accuracy:        histAcc,
			Latency:         lat,
			Cost:            costEff,
			Availability:    avail,
			CapabilityMatch: cap,
		},
	}
}

func rankModels(ctx RoutingContext) []RankedModel {
	ranked := make([]RankedModel, 0, len(ctx.Registry))
	for _, e := range ctx.Registry {
		if !e.Available {
			continue
		}
		ranked = append(ranked, scoreModel(ctx, e))
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].Score > ranked[j].Score })
	return ranked
}

func selectDiverseRankedModels(ranked []RankedModel, k int) (selected []RankedModel, explanation string) {
	kk := k
	if kk < 0 {
		kk = 0
	}
	if len(ranked) == 0 || kk == 0 {
		return nil, "empty_ranking_or_zero_k"
	}
	cap := kk
	if cap > len(ranked) {
		cap = len(ranked)
	}

	groups := make(map[string][]RankedModel)
	for _, r := range ranked {
		groups[r.ProviderID] = append(groups[r.ProviderID], r)
	}

	type provScore struct {
		id    string
		score float64
	}
	var order []provScore
	for pid, list := range groups {
		top := 0.0
		if len(list) > 0 {
			top = list[0].Score
		}
		order = append(order, provScore{id: pid, score: top})
	}
	sort.Slice(order, func(i, j int) bool { return order[i].score > order[j].score })

	providerOrder := make([]string, len(order))
	for i, p := range order {
		providerOrder[i] = p.id
	}

	selected = make([]RankedModel, 0, cap)
	round := 0
	for len(selected) < cap {
		progressed := false
		for _, pid := range providerOrder {
			g := groups[pid]
			if round >= len(g) {
				continue
			}
			selected = append(selected, g[round])
			progressed = true
			if len(selected) >= cap {
				break
			}
		}
		if !progressed {
			break
		}
		round++
	}

	ids := make([]string, len(selected))
	for i, s := range selected {
		ids[i] = s.ModelID
	}
	explanation = "diverse_round_robin_by_provider order=" + joinStrings(providerOrder, ">") + " picked=" + joinStrings(ids, ",")
	return selected, explanation
}

func joinStrings(parts []string, sep string) string {
	if len(parts) == 0 {
		return ""
	}
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += sep + parts[i]
	}
	return out
}

// PlanSubtaskRouting selects candidate models for a subtask.
func PlanSubtaskRouting(ctx RoutingContext, explorePaths bool, beamWidth, maxParallelCalls int) RoutingPlan {
	ranked := rankModels(ctx)
	bw := beamWidth
	if bw < 1 {
		bw = 1
	}

	poolSize := 1
	if explorePaths {
		want := bw * 2
		if want < bw {
			want = bw
		}
		poolSize = want
		if maxParallelCalls < poolSize {
			poolSize = maxParallelCalls
		}
		if len(ranked) < poolSize {
			poolSize = len(ranked)
		}
	}

	selected, explanation := selectDiverseRankedModels(ranked, poolSize)
	ids := make([]string, len(selected))
	for i, s := range selected {
		ids[i] = s.ModelID
	}

	return RoutingPlan{
		CandidateModelIDs:    ids,
		Ranked:               ranked,
		DiversityExplanation: explanation,
		CandidatePoolSize:    poolSize,
		BeamWidth:            bw,
		ModelRankSnapshot:    ranked,
	}
}
