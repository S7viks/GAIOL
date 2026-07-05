package httpserver

import (
	"os"
	"strconv"
	"strings"
)

func envBool(key string) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	return v == "1" || v == "true" || v == "yes" || v == "y" || v == "on"
}

func explorePathsDefaultOn() bool {
	v := strings.TrimSpace(os.Getenv("GAIOL_TS_EXPLORE_PATHS"))
	if v == "" {
		v = strings.TrimSpace(os.Getenv("GAIOL_EXPLORE_PATHS"))
	}
	if v == "" {
		return false
	}
	lower := strings.ToLower(v)
	return lower == "1" || lower == "true" || lower == "yes" || lower == "y" || lower == "on"
}

func beamWidthFromEnv() int {
	for _, key := range []string{"GAIOL_TS_BEAM_WIDTH", "GAIOL_BEAM_WIDTH"} {
		if n, err := strconv.Atoi(strings.TrimSpace(os.Getenv(key))); err == nil && n >= 1 {
			return n
		}
	}
	return 2
}

func consensusModeFromEnv() string {
	for _, key := range []string{"GAIOL_TS_CONSENSUS_MODE", "GAIOL_CONSENSUS_MODE"} {
		m := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
		if m == "uniform" || m == "static" || m == "abtc" {
			return m
		}
	}
	return "abtc"
}

func orchestratorDomainFromEnv() string {
	for _, key := range []string{"GAIOL_TS_DOMAIN", "GAIOL_DOMAIN"} {
		if d := strings.TrimSpace(os.Getenv(key)); d != "" {
			return d
		}
	}
	return "general"
}

func mapTaskKindV1(task string) string {
	switch strings.ToLower(strings.TrimSpace(task)) {
	case "code":
		return "code"
	case "summarization", "summarize":
		return "summarization"
	case "reasoning":
		return "reasoning"
	case "creative":
		return "creative"
	case "tool_use", "tool":
		return "tool_use"
	case "unknown":
		return "unknown"
	default:
		return "qa"
	}
}
