package httpserver

import (
	"context"
	"strings"

	"gaiol/internal/database"
)

type orchestrationPrefs struct {
	BeamWidth     int
	ConsensusMode string
	Domain        string
	ExplorePaths  bool
}

func orchestrationPrefsFromEnv() orchestrationPrefs {
	return orchestrationPrefs{
		BeamWidth:     beamWidthFromEnv(),
		ConsensusMode: consensusModeFromEnv(),
		Domain:        orchestratorDomainFromEnv(),
		ExplorePaths:  explorePathsDefaultOn(),
	}
}

func (d *Deps) resolveOrchestrationPrefs(ctx context.Context, tenantID string) orchestrationPrefs {
	prefs := orchestrationPrefsFromEnv()
	if d == nil || d.DB == nil || tenantID == "" {
		return prefs
	}
	s, err := d.DB.GetTenantSettings(ctx, tenantID)
	if err != nil || s == nil {
		return prefs
	}
	if s.BeamWidth != nil && *s.BeamWidth >= 1 {
		prefs.BeamWidth = *s.BeamWidth
	}
	if s.ConsensusMode != nil {
		if m := normalizeConsensusMode(*s.ConsensusMode); m != "" {
			prefs.ConsensusMode = m
		}
	}
	if s.Domain != nil {
		if dom := strings.TrimSpace(*s.Domain); dom != "" {
			prefs.Domain = dom
		}
	}
	if s.ExplorePaths != nil {
		prefs.ExplorePaths = *s.ExplorePaths
	}
	return prefs
}

func normalizeConsensusMode(m string) string {
	switch strings.ToLower(strings.TrimSpace(m)) {
	case "uniform", "static", "abtc":
		return strings.ToLower(strings.TrimSpace(m))
	default:
		return ""
	}
}

func preferencesResponseFromSettings(s *database.TenantSettings) map[string]interface{} {
	env := orchestrationPrefsFromEnv()
	out := map[string]interface{}{
		"budget_limit":  nil,
		"strategy":      "balanced",
		"beam_width":    env.BeamWidth,
		"consensus_mode": env.ConsensusMode,
		"domain":        env.Domain,
		"explore_paths": env.ExplorePaths,
	}
	if s == nil {
		return out
	}
	out["budget_limit"] = s.BudgetLimit
	if s.Strategy != "" {
		out["strategy"] = s.Strategy
	}
	if s.BeamWidth != nil {
		out["beam_width"] = *s.BeamWidth
	}
	if s.ConsensusMode != nil && normalizeConsensusMode(*s.ConsensusMode) != "" {
		out["consensus_mode"] = normalizeConsensusMode(*s.ConsensusMode)
	}
	if s.Domain != nil && strings.TrimSpace(*s.Domain) != "" {
		out["domain"] = strings.TrimSpace(*s.Domain)
	}
	if s.ExplorePaths != nil {
		out["explore_paths"] = *s.ExplorePaths
	}
	return out
}
