package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"gaiol/internal/apijson"
	"gaiol/internal/database"
	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
	"gaiol/internal/keys"
	"strings"

	"github.com/google/uuid"
)

// orchestrateTimeout bounds one user orchestrate call (decomposition + beams across providers).
const orchestrateTimeout = 120 * time.Second

// userChatRequest is the normalized user prompt accepted by every chat surface
// (dashboard /api/query/smart, /v1/chat, CLI later).
type userChatRequest struct {
	Prompt      string
	Task        string
	Strategy    string
	MaxTokens   int
	Temperature float64
}

// contractCredentials maps resolved tenant credentials to the orchestrate v1 payload.
func contractCredentials(tc *keys.TenantCredentials) *orchestratorv1.RequestCredentialsV1 {
	if tc == nil || len(tc.Providers) == 0 {
		return nil
	}
	out := &orchestratorv1.RequestCredentialsV1{
		SchemaVersion: "1",
	}
	for _, p := range tc.Providers {
		out.Providers = append(out.Providers, orchestratorv1.ProviderCredentialV1{
			ID:      p.ID,
			Kind:    p.Kind,
			APIKey:  p.APIKey,
			BaseURL: p.BaseURL,
		})
	}
	for _, m := range tc.Models {
		out.Models = append(out.Models, orchestratorv1.CredentialModelV1{
			Provider:    m.Provider,
			ModelID:     m.ModelID,
			DisplayName: m.DisplayName,
		})
	}
	return out
}

// orchestrateUserRequest is the single inference path for all user chat.
// It resolves tenant credentials (auth mode) and runs the in-process Go orchestrator.
func (d *Deps) orchestrateUserRequest(
	w http.ResponseWriter,
	r *http.Request,
	in userChatRequest,
	tenantCtx database.TenantContext,
	gaiolKeyID string,
) {
	if d.Orchestrator == nil {
		apijson.WriteError(w, http.StatusServiceUnavailable,
			"Orchestrator is not configured.",
			"orchestrator_unavailable")
		return
	}

	// Production: tenant DB keys only. Local no-auth mode sends no credentials and the
	// orchestrator falls back to its own env keys (dev-only exception).
	var creds *orchestratorv1.RequestCredentialsV1
	if !d.AuthDisabled {
		tc, err := keys.ResolveTenantCredentials(r.Context(), d.DB, tenantCtx.TenantID)
		if err != nil {
			log.Printf("ERROR: resolve tenant credentials tenant_id=%s: %v", tenantCtx.TenantID, err)
			apijson.WriteError(w, http.StatusInternalServerError, "Failed to load provider credentials", "credentials_error")
			return
		}
		creds = contractCredentials(tc)
		if creds == nil {
			apijson.WriteError(w, http.StatusBadRequest,
				"No provider connected. Add a provider API key in Settings before chatting.",
				"no_provider_keys")
			return
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), orchestrateTimeout)
	defer cancel()

	traceID := uuid.New().String()
	orch := d.resolveOrchestrationPrefs(r.Context(), tenantCtx.TenantID)
	explore := orch.ExplorePaths
	switch strings.ToLower(strings.TrimSpace(in.Strategy)) {
	case "beam":
		explore = true
	}
	bw := orch.BeamWidth
	consensus := orch.ConsensusMode
	maxTok := in.MaxTokens
	tempPtr := in.Temperature

	reqV1 := &orchestratorv1.OrchestrateRequestV1{
		SchemaVersion: "1.0",
		TraceID:       traceID,
		Domain:        orch.Domain,
		TaskKind:      mapTaskKindV1(in.Task),
		Objective:     in.Prompt,
		Messages: []orchestratorv1.ChatMessageV1{
			{Role: "user", Content: in.Prompt},
		},
		Constraints: &orchestratorv1.TaskConstraintsV1{
			Temperature:     &tempPtr,
			MaxOutputTokens: &maxTok,
		},
		ExplorePaths:  &explore,
		BeamWidth:     &bw,
		ConsensusMode: consensus,
		Credentials:   creds,
	}
	if tenantCtx.TenantID != "" {
		reqV1.SessionID = tenantCtx.TenantID
	}

	res, err := d.Orchestrator.Orchestrate(ctx, reqV1)
	if err != nil {
		writeOrchestrateError(w, ctx, err)
		return
	}

	metrics := orchestratorv1.SummarizeOrchestrationTrace(&res.Trace, 0)
	totalCost := 0.0
	totalTokens := 0
	processingMs := int64(0)
	usageModelID := "go-orchestrator"
	if metrics != nil {
		totalCost = metrics.CostUSD.Total
		totalTokens = metrics.TotalTokens
		processingMs = metrics.DurationMs
	}
	if len(res.Trace.Subtasks) > 0 {
		if chosen := strings.TrimSpace(res.Trace.Subtasks[0].ChosenModelID); chosen != "" {
			usageModelID = chosen
		}
	}

	if d.DB != nil && tenantCtx.TenantID != "" {
		if err := logUsageToAPIQueries(d.DB, tenantCtx, usageModelID, totalTokens, totalCost, int(processingMs), true, "", gaiolKeyID); err != nil {
			log.Printf("usage log failed tenant_id=%s: %v", tenantCtx.TenantID, err)
		}
	}

	response := map[string]interface{}{
		"uaip": true,
		"status": map[string]interface{}{
			"success": true,
		},
		"result": map[string]interface{}{
			"data":          res.Answer,
			"tokens_used":   totalTokens,
			"model_used":    usageModelID,
			"processing_ms": processingMs,
			"quality":       1.0,
		},
		"metadata": map[string]interface{}{
			"cost_info": map[string]interface{}{
				"total_cost": totalCost,
			},
			"session_id":     traceID,
			"steps_executed": len(res.Trace.Subtasks),
			"engine":         "go_orchestrator",
			"trace_id":       res.TraceID,
		},
		"model_id":    "go-orchestrator",
		"model_name":  "GAIOL Go Orchestrator",
		"response":    res.Answer,
		"session_id":  traceID,
		"tokens_used": totalTokens,
		"cost":        totalCost,
		"latency_ms":  processingMs,
		"quality":     1.0,
		"strategy":    "go_orchestrator",
		"orchestration": map[string]interface{}{
			"schema_version":      res.SchemaVersion,
			"trace_id":            res.TraceID,
			"trust_updates_count": len(res.TrustUpdates),
			"consensus_mode":      consensus,
			"explore_paths":       explore,
			"beam_width":          bw,
		},
		"orchestration_trace":         res.Trace,
		"orchestration_trust_updates": res.TrustUpdates,
		"orchestration_metrics":       metrics,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("encode orchestrate response: %v", err)
	}
}

// writeOrchestrateError maps orchestrate call failures: timeout -> 504, network
// unreachable -> 503 (no fallback), upstream non-200 -> 502.
func writeOrchestrateError(w http.ResponseWriter, ctx context.Context, err error) {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		log.Printf("ERROR: orchestrator timeout: %v", err)
		apijson.WriteError(w, http.StatusGatewayTimeout, "Orchestrator timeout", "orchestrator_timeout")
		return
	}
	log.Printf("ERROR: orchestrator error: %v", err)
	apijson.WriteError(w, http.StatusBadGateway, "Orchestrator error: "+err.Error(), "orchestrator_error")
}

// handleSetupStatus reports single-path readiness: orchestrator reachability and
// whether the tenant has at least one connected provider.
func (d *Deps) handleSetupStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	orchestratorConfigured := d.Orchestrator != nil
	orchestratorReachable := d.Orchestrator != nil

	tenantReady := false
	providersConnected := 0
	gaiolKeysCount := 0
	if d.AuthDisabled {
		// Local dev: orchestrator env keys are the pool; tenant setup is not required.
		tenantReady = true
	} else if tc, err := database.EnsureTenantContext(r.Context()); err == nil {
		creds, err := keys.ResolveTenantCredentials(r.Context(), d.DB, tc.TenantID)
		if err == nil && creds != nil {
			providersConnected = len(creds.Providers)
			tenantReady = providersConnected > 0
		}
		if d.DB != nil {
			if list, err := keys.ListGAIOLKeys(r.Context(), d.DB, tc.TenantID); err == nil {
				gaiolKeysCount = len(list)
			}
		}
	}

	setupComplete := d.AuthDisabled || (tenantReady && gaiolKeysCount > 0)

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"inference_mode":          "orchestrator_only",
		"orchestrator_configured": orchestratorConfigured,
		"orchestrator_reachable":  orchestratorReachable,
		"tenant_ready":            tenantReady,
		"providers_connected":     providersConnected,
		"gaiol_keys_count":        gaiolKeysCount,
		"setup_complete":          setupComplete,
	})
}
