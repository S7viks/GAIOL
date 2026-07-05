package orchestration

import "strings"

// ScoreAnswer is a cheap deterministic quality proxy (orchestrator/src/evaluation/scorer.ts).
func ScoreAnswer(objective, answer string) float64 {
	j := tokenJaccard(objective, answer)
	lenScore := answerLenScore(answer)
	return clamp01(0.6*j + 0.4*lenScore)
}

func answerLenScore(answer string) float64 {
	if len(answer) >= 400 {
		return 1
	}
	return float64(len(answer)) / 400
}

func computeLexicalCoverage(candidate, query string) float64 {
	queryTokens := make(map[string]struct{})
	for _, w := range strings.Fields(strings.ToLower(query)) {
		if len(w) > 3 {
			queryTokens[w] = struct{}{}
		}
	}
	candidateTokens := tokenSet(candidate)
	if len(queryTokens) == 0 {
		return 1
	}
	overlap := 0
	for t := range queryTokens {
		if _, ok := candidateTokens[t]; ok {
			overlap++
		}
	}
	recall := float64(overlap) / float64(len(queryTokens))
	precision := float64(overlap) / float64(maxInt(1, len(candidateTokens)))
	if precision+recall == 0 {
		return 0
	}
	return (2 * precision * recall) / (precision + recall)
}

func computeStructuralCompleteness(candidate string, queryLen int) float64 {
	score := 0.5
	if len(candidate) > queryLen/2 {
		score += 0.2
	}
	if len(candidate) > queryLen*3/2 {
		score += 0.1
	}
	lower := strings.ToLower(candidate)
	for _, m := range []string{"conclusion", "summary", "therefore", "in conclusion", "result", "thus"} {
		if strings.Contains(lower, m) {
			score += 0.2
			break
		}
	}
	return clamp01(score)
}

func computeEvaluateQuality(candidate, query string) float64 {
	semantic := tokenJaccard(candidate, query)
	lexical := computeLexicalCoverage(candidate, query)
	structural := computeStructuralCompleteness(candidate, len(query))
	return clamp01(0.4*semantic + 0.3*lexical + 0.3*structural)
}

func crossModelAgreement(candidate string, others []string) float64 {
	if len(others) == 0 {
		return 1
	}
	sum := 0.0
	for _, o := range others {
		sum += tokenJaccard(candidate, o)
	}
	return sum / float64(len(others))
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
