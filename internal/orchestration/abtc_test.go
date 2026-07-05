package orchestration

import (
	"math"
	"testing"
)

func TestComputePosteriorMean(t *testing.T) {
	t.Parallel()
	if got := computePosteriorMean(1, 1); math.Abs(got-0.5) > 1e-9 {
		t.Fatalf("uniform prior mean: got %v want 0.5", got)
	}
	if got := computePosteriorMean(3, 1); math.Abs(got-0.75) > 1e-9 {
		t.Fatalf("skewed mean: got %v want 0.75", got)
	}
}

func TestUpdateTrustWinnerIncreasesAlpha(t *testing.T) {
	t.Parallel()
	prior := BetaTrust{Alpha: 1, Beta: 1}
	next := updateTrust(prior.Alpha, prior.Beta, true, Lambda)
	if next.Alpha <= prior.Alpha {
		t.Fatalf("winner should increase alpha: %+v -> %+v", prior, next)
	}
}

func TestUpdateTrustLoserIncreasesBeta(t *testing.T) {
	t.Parallel()
	prior := BetaTrust{Alpha: 1, Beta: 1}
	next := updateTrust(prior.Alpha, prior.Beta, false, Lambda)
	if next.Beta <= prior.Beta {
		t.Fatalf("loser should increase beta: %+v -> %+v", prior, next)
	}
}

func TestComputeCompositeScoreWeights(t *testing.T) {
	t.Parallel()
	score := computeCompositeScore(1, 0, 0)
	if math.Abs(score-WQuality) > 1e-9 {
		t.Fatalf("quality-only score: got %v want %v", score, WQuality)
	}
}

func TestComputeConfidence(t *testing.T) {
	t.Parallel()
	conf := computeConfidence([]float64{0.8, 0.2})
	if math.Abs(conf-0.8) > 1e-9 {
		t.Fatalf("confidence: got %v want 0.8", conf)
	}
}
