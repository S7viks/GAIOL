package orchestration

// BetaTrust is a Beta distribution for model trust.
type BetaTrust struct {
	Alpha float64 `json:"alpha"`
	Beta  float64 `json:"beta"`
}

// UniformPrior is Beta(1,1).
var UniformPrior = BetaTrust{Alpha: 1, Beta: 1}

// TrustRecord is a persisted trust row.
type TrustRecord struct {
	ModelID      string    `json:"model_id"`
	Domain       string    `json:"domain"`
	Distribution BetaTrust `json:"distribution"`
	UpdatedAt    string    `json:"updated_at"`
}

func betaMean(t BetaTrust) float64 {
	if t.Alpha+t.Beta <= 0 {
		return 0.5
	}
	return t.Alpha / (t.Alpha + t.Beta)
}

func betaMeanPair(prior, posterior BetaTrust) (priorMean, posteriorMean float64) {
	return betaMean(prior), betaMean(posterior)
}
