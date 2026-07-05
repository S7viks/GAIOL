package orchestration

import "strings"

// EvalExample is one contains-based eval case.
type EvalExample struct {
	Objective        string   `json:"objective"`
	ExpectedContains []string `json:"expectedContains"`
}

// EvalResult is one example result.
type EvalResult struct {
	Objective string  `json:"objective"`
	Pass      bool    `json:"pass"`
	Score     float64 `json:"score"`
	Notes     string  `json:"notes,omitempty"`
}

// EvaluateAgainstContains scores answer text against expected substrings (case-insensitive).
func EvaluateAgainstContains(ex EvalExample, answerText string) EvalResult {
	base := ScoreAnswer(ex.Objective, answerText)
	if len(ex.ExpectedContains) == 0 {
		return EvalResult{Objective: ex.Objective, Pass: base >= 0.2, Score: base}
	}
	answerLower := strings.ToLower(answerText)
	var missing []string
	for _, s := range ex.ExpectedContains {
		needle := strings.ToLower(strings.TrimSpace(s))
		if needle == "" {
			continue
		}
		if !strings.Contains(answerLower, needle) {
			missing = append(missing, s)
		}
	}
	ok := len(missing) == 0
	score := base * 0.5
	if ok {
		score = minFloat(1, base+0.2)
	}
	notes := ""
	if len(missing) > 0 {
		notes = "missing: " + strings.Join(missing, ", ")
	}
	return EvalResult{Objective: ex.Objective, Pass: ok, Score: score, Notes: notes}
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
