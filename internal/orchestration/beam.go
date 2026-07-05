package orchestration

import "sort"

// ScoredPath is a beam path with score.
type ScoredPath struct {
	PathID     string
	ModelID    string
	ProviderID string
	Result     ModelCallResult
	Score      float64
}

// PathIDForModel returns deterministic path id for a single-hop path.
func PathIDForModel(modelID string) string {
	return "path:" + modelID
}

// ScorePaths attaches heuristic scores to each model result.
func ScorePaths(subtaskDescription string, results []ModelCallResult) []ScoredPath {
	paths := make([]ScoredPath, 0, len(results))
	for _, r := range results {
		score := 0.0
		if r.Error == "" {
			score = ScoreAnswer(subtaskDescription, r.Text)
		}
		paths = append(paths, ScoredPath{
			PathID:     PathIDForModel(r.ModelID),
			ModelID:    r.ModelID,
			ProviderID: r.ProviderID,
			Result:     r,
			Score:      score,
		})
	}
	sort.Slice(paths, func(i, j int) bool {
		if paths[i].Score != paths[j].Score {
			return paths[i].Score > paths[j].Score
		}
		return paths[i].PathID < paths[j].PathID
	})
	return paths
}

// PruneBeam keeps top beamWidth paths by score.
func PruneBeam(paths []ScoredPath, beamWidth int) (kept, discarded []ScoredPath) {
	sorted := append([]ScoredPath(nil), paths...)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Score != sorted[j].Score {
			return sorted[i].Score > sorted[j].Score
		}
		return sorted[i].PathID < sorted[j].PathID
	})
	w := beamWidth
	if w < 1 {
		w = 1
	}
	limit := w
	if limit > len(sorted) {
		limit = len(sorted)
	}
	kept = sorted[:limit]
	keptIDs := make(map[string]struct{}, len(kept))
	for _, p := range kept {
		keptIDs[p.PathID] = struct{}{}
	}
	for _, p := range sorted {
		if _, ok := keptIDs[p.PathID]; !ok {
			discarded = append(discarded, p)
		}
	}
	return kept, discarded
}
