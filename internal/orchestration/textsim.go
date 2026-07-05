package orchestration

import "strings"

func tokenJaccard(a, b string) float64 {
	setA := tokenSet(a)
	setB := tokenSet(b)
	if len(setA) == 0 && len(setB) == 0 {
		return 1
	}
	inter := 0
	union := make(map[string]struct{}, len(setA)+len(setB))
	for t := range setA {
		union[t] = struct{}{}
		if _, ok := setB[t]; ok {
			inter++
		}
	}
	for t := range setB {
		union[t] = struct{}{}
	}
	if len(union) == 0 {
		return 0
	}
	return float64(inter) / float64(len(union))
}

func tokenSet(s string) map[string]struct{} {
	out := make(map[string]struct{})
	for _, t := range strings.Fields(strings.ToLower(s)) {
		if t != "" {
			out[t] = struct{}{}
		}
	}
	return out
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
