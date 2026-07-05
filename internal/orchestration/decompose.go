package orchestration

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// sentenceBoundary matches end-of-sentence punctuation followed by whitespace.
var sentenceBoundary = regexp.MustCompile(`[.!?]\s+`)

func splitSentences(text string) []string {
	loc := sentenceBoundary.FindAllStringIndex(text, -1)
	if len(loc) == 0 {
		t := strings.TrimSpace(text)
		if t == "" {
			return nil
		}
		return []string{t}
	}
	out := make([]string, 0, len(loc)+1)
	start := 0
	for _, idx := range loc {
		// Include the sentence-ending punctuation in the chunk.
		end := idx[0] + 1
		if chunk := strings.TrimSpace(text[start:end]); chunk != "" {
			out = append(out, chunk)
		}
		start = idx[1]
	}
	if tail := strings.TrimSpace(text[start:]); tail != "" {
		out = append(out, tail)
	}
	return out
}

// HeuristicDecompose is a fast baseline decomposer (no extra LLM call).
func HeuristicDecompose(req OrchestrationRequest) DecompositionResult {
	text := strings.TrimSpace(req.Objective)
	sentences := splitSentences(text)
	if len(sentences) == 0 {
		fallbackObj := req.Objective
		if fallbackObj == "" {
			var parts []string
			for _, m := range req.Messages {
				parts = append(parts, m.Content)
			}
			fallbackObj = strings.Join(parts, "\n")
		}
		return DecompositionResult{
			Subtasks:  fallbackDecompose(fallbackObj, req.TaskKind),
			Rationale: "fallback-7step: empty objective",
		}
	}

	if len(sentences) == 1 {
		return DecompositionResult{
			Subtasks: []SubtaskSpec{{
				ID:          uuid.NewString(),
				Title:       "main",
				Description: text,
				TaskKind:    req.TaskKind,
			}},
			Rationale: "single-step",
		}
	}

	subtasks := make([]SubtaskSpec, len(sentences))
	for i, s := range sentences {
		subtasks[i] = SubtaskSpec{
			ID:                   uuid.NewString(),
			Title:                "step-" + strconv.Itoa(i+1),
			Description:          s,
			TaskKind:             req.TaskKind,
			RequiredCapabilities: guessCaps(req.TaskKind),
		}
	}
	return DecompositionResult{
		Subtasks:  subtasks,
		Rationale: "sentence-split",
	}
}

var fallbackSteps = []struct {
	title string
	desc  func(string) string
}{
	{"problem-statement", func(q string) string { return "Extract and restate the core problem: " + q }},
	{"constraint-identification", func(q string) string { return "Identify constraints, boundary conditions, and requirements for: " + q }},
	{"approach-selection", func(q string) string { return "Select the most appropriate reasoning approach or algorithm for: " + q }},
	{"step-by-step-execution", func(q string) string { return "Execute the chosen approach step by step for: " + q }},
	{"intermediate-verification", func(q string) string { return "Verify intermediate results and check for logical consistency for: " + q }},
	{"synthesis", func(q string) string { return "Synthesize all intermediate results into a cohesive answer for: " + q }},
	{"confidence-assessment", func(q string) string { return "Assess confidence in the final answer and flag any remaining uncertainties for: " + q }},
}

func fallbackDecompose(objective, kind string) []SubtaskSpec {
	out := make([]SubtaskSpec, len(fallbackSteps))
	for i, step := range fallbackSteps {
		out[i] = SubtaskSpec{
			ID:                   uuid.NewString(),
			Title:                step.title,
			Description:          step.desc(objective),
			TaskKind:             kind,
			RequiredCapabilities: guessCaps(kind),
		}
	}
	return out
}

func guessCaps(kind string) []string {
	switch kind {
	case "code":
		return []string{"code"}
	case "reasoning":
		return []string{"reasoning"}
	default:
		return nil
	}
}
