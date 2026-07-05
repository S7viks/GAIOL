package httpserver

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	"gaiol/internal/apijson"
	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
	"gaiol/internal/orchestration"
)

func (d *Deps) orchestratorReady() bool {
	return d.Orchestrator != nil
}

// handleV1Orchestrate serves POST /v1/orchestrate (Go in-process orchestrator).
func (d *Deps) handleV1Orchestrate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !d.orchestratorReady() {
		apijson.WriteError(w, http.StatusServiceUnavailable, "Orchestrator is not configured", "orchestrator_unavailable")
		return
	}
	raw, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		apijson.WriteError(w, http.StatusBadRequest, "read body", "bad_request")
		return
	}
	var req orchestratorv1.OrchestrateRequestV1
	if err := json.Unmarshal(raw, &req); err != nil {
		apijson.WriteError(w, http.StatusBadRequest, "Invalid JSON", "invalid_json")
		return
	}
	res, err := d.Orchestrator.Orchestrate(r.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "no_models_for_credentials") {
			apijson.WriteError(w, http.StatusBadRequest, "Credentials payload resolved to zero routable models.", "no_models_for_credentials")
			return
		}
		log.Printf("orchestrate error: %v", err)
		apijson.WriteError(w, http.StatusInternalServerError, err.Error(), "orchestrate_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(res)
}

func (d *Deps) handleOrchestrationTraceProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !d.orchestratorReady() {
		apijson.WriteError(w, http.StatusServiceUnavailable, "Orchestrator is not configured", "orchestrator_disabled")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/orchestration/traces/")
	id = strings.TrimPrefix(id, "/v1/traces/")
	id = strings.Trim(id, "/")
	if id == "" || strings.Contains(id, "/") {
		http.Error(w, "trace id required", http.StatusBadRequest)
		return
	}
	bundle, ok := d.Orchestrator.GetTraceBundle(id)
	if !ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"error": "not_found", "trace_id": id})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bundle)
}

func (d *Deps) handleOrchestrationTrustProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !d.orchestratorReady() {
		apijson.WriteError(w, http.StatusServiceUnavailable, "Orchestrator is not configured", "orchestrator_disabled")
		return
	}
	domain := strings.TrimSpace(r.URL.Query().Get("domain"))
	records, err := d.Orchestrator.ListTrust(domain)
	if err != nil {
		apijson.WriteError(w, http.StatusInternalServerError, err.Error(), "trust_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"records": records,
		"count":   len(records),
		"domain":  nilIfEmpty(domain),
	})
}

func (d *Deps) handleOrchestrationTraceIDsProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !d.orchestratorReady() {
		apijson.WriteError(w, http.StatusServiceUnavailable, "Orchestrator is not configured", "orchestrator_disabled")
		return
	}
	limit := 50
	if s := strings.TrimSpace(r.URL.Query().Get("limit")); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			limit = n
		}
	}
	ids, err := d.Orchestrator.ListTraceIDs(limit)
	if err != nil {
		apijson.WriteError(w, http.StatusInternalServerError, err.Error(), "trace_index_error")
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"trace_ids": ids, "count": len(ids)})
}

func (d *Deps) handleOrchestrationEvalContainsProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if !d.orchestratorReady() {
		apijson.WriteError(w, http.StatusServiceUnavailable, "Orchestrator is not configured", "orchestrator_disabled")
		return
	}
	var body struct {
		Examples   []orchestration.EvalExample `json:"examples"`
		AnswerText string                      `json:"answerText"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil || len(body.Examples) == 0 {
		apijson.WriteError(w, http.StatusBadRequest, "examples_required", "validation_error")
		return
	}
	results := make([]orchestration.EvalResult, len(body.Examples))
	passAll := true
	for i, ex := range body.Examples {
		results[i] = orchestration.EvaluateAgainstContains(ex, body.AnswerText)
		if !results[i].Pass {
			passAll = false
		}
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"pass":    passAll,
		"results": results,
	})
}

func nilIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
