package orchestration

func computePosteriorMean(alpha, beta float64) float64 {
	if alpha+beta <= 0 {
		return 0.5
	}
	return alpha / (alpha + beta)
}

func updateTrust(alpha, beta float64, isWinner bool, lambda float64) BetaTrust {
	reward := 0.0
	penalty := 0.0
	if isWinner {
		reward = 1.0
	} else {
		penalty = 1.0
	}
	return BetaTrust{
		Alpha: lambda*alpha + reward,
		Beta:  lambda*beta + penalty,
	}
}

func computeCompositeScore(quality, agreement, trustMean float64) float64 {
	return WQuality*quality + WAgreement*agreement + WTrust*trustMean
}

func computeConfidence(scores []float64) float64 {
	total := 0.0
	for _, s := range scores {
		total += s
	}
	if total <= 0 || len(scores) == 0 {
		return 0
	}
	top := scores[0]
	for _, s := range scores {
		if s > top {
			top = s
		}
	}
	return top / total
}
