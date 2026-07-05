package orchestration

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// ConsensusInput is input to RunConsensus.
type ConsensusInput struct {
	Query                  string
	Mode                   string
	Domain                 string
	Candidates             []ModelCallResult
	Scores                 map[string]float64
	StaticWeights          map[string]float64
	TrustMeans             map[string]float64
	TrustRecords           map[string]BetaTrust
	ABTCConsensusExponent  float64
	Lambda                 float64
}

// ConsensusOutput is the chosen answer and metadata.
type ConsensusOutput struct {
	Text          string
	ChosenModelID string
	Weights       map[string]float64
	Agreement     float64
	Confidence    float64
	Notes         string
}

type scoredCandidate struct {
	candidate ModelCallResult
	score     float64
	alpha     float64
	beta      float64
	tauHat    float64
}

// RunConsensus aggregates parallel model outputs (orchestrator/src/consensus/engine.ts).
func RunConsensus(input ConsensusInput) ConsensusOutput {
	successful := filterSuccessful(input.Candidates)
	byID := make(map[string]ModelCallResult, len(input.Candidates))
	for _, c := range input.Candidates {
		byID[c.ModelID] = c
	}
	ids := make([]string, 0, len(successful))
	for _, c := range successful {
		ids = append(ids, c.ModelID)
	}

	if len(ids) == 0 {
		chosenID := "none"
		text := "All model calls failed. Check your provider API key in Settings — it may be invalid, expired, or missing inference permissions."
		if len(input.Candidates) > 0 {
			chosenID = input.Candidates[0].ModelID
			if input.Candidates[0].Error != "" {
				text = input.Candidates[0].Error + ". Re-save your provider key in Settings if this persists."
			}
		}
		return ConsensusOutput{
			Text:          text,
			ChosenModelID: chosenID,
			Weights:       map[string]float64{},
			Agreement:     0,
			Notes:         "no successful candidates",
		}
	}

	if input.Mode == "abtc" {
		return runABTCConsensus(input, successful, ids)
	}

	weights := buildModeWeights(input, ids)
	weights = normalizeWeights(weights)

	bestID := ids[0]
	best := -1.0
	for _, id := range ids {
		comb := (weights[id]) * (input.Scores[id])
		if comb > best {
			best = comb
			bestID = id
		}
	}

	chosenText := byID[bestID].Text
	otherTexts := make([]string, 0, len(ids)-1)
	for _, id := range ids {
		if id != bestID {
			otherTexts = append(otherTexts, byID[id].Text)
		}
	}
	agreement := crossModelAgreement(chosenText, otherTexts)

	text := chosenText
	if input.Mode == "uniform" && len(ids) > 1 {
		texts := make([]string, len(ids))
		for i, id := range ids {
			texts[i] = byID[id].Text
		}
		text = weightedBlend(texts, ids, weights)
	}

	return ConsensusOutput{
		Text:          text,
		ChosenModelID: bestID,
		Weights:       weights,
		Agreement:     agreement,
		Confidence:    0,
	}
}

func runABTCConsensus(input ConsensusInput, successful []ModelCallResult, ids []string) ConsensusOutput {
	trustStore := make(map[string]BetaTrust)
	for k, v := range input.TrustRecords {
		trustStore[k] = v
	}
	query := input.Query
	lambda := input.Lambda
	if lambda <= 0 {
		lambda = Lambda
	}

	scored := make([]scoredCandidate, 0, len(successful))
	for _, candidate := range successful {
		prior := trustStore[candidate.ModelID]
		alpha := prior.Alpha
		beta := prior.Beta
		if alpha == 0 && beta == 0 {
			alpha, beta = AlphaInit, BetaInit
		}
		tauHat := computePosteriorMean(alpha, beta)
		quality := computeEvaluateQuality(candidate.Text, query)
		others := make([]string, 0, len(successful)-1)
		for _, other := range successful {
			if other.ModelID != candidate.ModelID {
				others = append(others, other.Text)
			}
		}
		agreement := crossModelAgreement(candidate.Text, others)
		score := computeCompositeScore(quality, agreement, tauHat)
		scored = append(scored, scoredCandidate{candidate: candidate, score: score, alpha: alpha, beta: beta, tauHat: tauHat})
	}

	sort.Slice(scored, func(i, j int) bool { return scored[i].score > scored[j].score })

	allScores := make([]float64, len(scored))
	for i, s := range scored {
		allScores[i] = s.score
	}
	sigma := computeConfidence(allScores)

	baseWinner := scored[0]
	winner := baseWinner.candidate
	winnerText := winner.Text
	if sigma < ThetaMin && len(scored) >= 2 {
		topN := minInt(3, len(scored))
		parts := make([]string, 0, topN)
		for i := 0; i < topN; i++ {
			if isUsableResponseText(scored[i].candidate.Text) {
				parts = append(parts, scored[i].candidate.Text)
			}
		}
		switch len(parts) {
		case 0:
			// keep base winner text
		case 1:
			winnerText = parts[0]
		default:
			winnerText = strings.Join(parts, "\n\n")
		}
	}

	weightMap := make(map[string]float64, len(scored))
	for _, entry := range scored {
		weightMap[entry.candidate.ModelID] = maxFloat(0, entry.score)
	}
	weights := normalizeWeights(weightMap)

	otherTexts := make([]string, 0, len(successful)-1)
	for _, c := range successful {
		if c.ModelID != baseWinner.candidate.ModelID {
			otherTexts = append(otherTexts, c.Text)
		}
	}
	winnerAgreement := crossModelAgreement(winnerText, otherTexts)

	notes := ""
	if sigma < ThetaMin && len(scored) >= 2 {
		notes = "synthesized top candidates"
	}

	return ConsensusOutput{
		Text:          winnerText,
		ChosenModelID: baseWinner.candidate.ModelID,
		Weights:       weights,
		Agreement:     winnerAgreement,
		Confidence:    sigma,
		Notes:         notes,
	}
}

func isUsableResponseText(text string) bool {
	t := strings.TrimSpace(text)
	if t == "" {
		return false
	}
	if strings.HasPrefix(t, "[Empty response - Finish reason:") {
		return false
	}
	return true
}

func filterSuccessful(candidates []ModelCallResult) []ModelCallResult {
	out := make([]ModelCallResult, 0, len(candidates))
	for _, c := range candidates {
		if c.Error == "" && isUsableResponseText(c.Text) {
			out = append(out, c)
		}
	}
	return out
}

func buildModeWeights(input ConsensusInput, ids []string) map[string]float64 {
	weights := make(map[string]float64, len(ids))
	switch input.Mode {
	case "uniform":
		for _, id := range ids {
			weights[id] = 1
		}
	case "static":
		for _, id := range ids {
			if w, ok := input.StaticWeights[id]; ok {
				weights[id] = w
			} else {
				weights[id] = 1
			}
		}
	default:
		exp := input.ABTCConsensusExponent
		if exp == 0 {
			exp = 1
		}
		for _, id := range ids {
			m, ok := input.TrustMeans[id]
			base := 0.5
			if ok && m > 0 {
				base = m
			} else if s, ok := input.Scores[id]; ok {
				base = betaFallbackFromScore(s)
			}
			if exp == 1 {
				weights[id] = base
			} else {
				weights[id] = powFloat(maxFloat(1e-9, base), exp)
			}
		}
	}
	return weights
}

func normalizeWeights(raw map[string]float64) map[string]float64 {
	sum := 0.0
	for _, v := range raw {
		sum += v
	}
	out := make(map[string]float64, len(raw))
	if sum <= 0 {
		u := 0.0
		if len(raw) > 0 {
			u = 1 / float64(len(raw))
		}
		for k := range raw {
			out[k] = u
		}
		return out
	}
	for k, v := range raw {
		out[k] = v / sum
	}
	return out
}

func betaFallbackFromScore(score float64) float64 {
	if score < 0 || score > 1 || math.IsNaN(score) {
		return 0.5
	}
	return clamp01(math.Max(0.05, score))
}

func weightedBlend(texts []string, ids []string, weights map[string]float64) string {
	var parts []string
	for i, id := range ids {
		w := weights[id]
		if w <= 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("[%s w=%.2f]\n%s", id, w, texts[i]))
	}
	return strings.Join(parts, "\n\n---\n\n")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func powFloat(base, exp float64) float64 {
	return math.Pow(base, exp)
}
