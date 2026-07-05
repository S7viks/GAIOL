package orchestration

import (
	"context"
	"fmt"

	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
)

// Service is the in-process Go orchestrator (replaces TS runtime).
type Service struct {
	Trust      *MemoryTrustStore
	Traces     *MemoryTraceStore
	EnvRuntime *Runtime
	Config     OrchestratorConfig
}

// NewService creates a service with in-memory stores and env-backed fallback runtime.
func NewService() *Service {
	return &Service{
		Trust:      NewMemoryTrustStore(),
		Traces:     NewMemoryTraceStore(),
		EnvRuntime: RuntimeFromEnv(),
		Config:     DefaultConfig(),
	}
}

// Orchestrate runs the v1 contract entrypoint.
func (s *Service) Orchestrate(ctx context.Context, req *orchestratorv1.OrchestrateRequestV1) (*orchestratorv1.OrchestrateResponseV1, error) {
	if err := orchestratorv1.ValidateOrchestrateRequestV1(req); err != nil {
		return nil, err
	}

	runtime := s.EnvRuntime
	if req.Credentials != nil {
		built, err := RuntimeFromCredentials(req.Credentials)
		if err != nil {
			return nil, err
		}
		if built == nil || len(built.Registry) == 0 {
			return nil, fmt.Errorf("no_models_for_credentials")
		}
		runtime = built
	}
	if runtime == nil || len(runtime.Registry) == 0 {
		return nil, fmt.Errorf("no models available in orchestrator registry")
	}

	domainReq := RequestFromV1(req)
	cfgOverride := ConfigOverrideFromV1(req)

	pipeline := &Pipeline{
		Trust:    s.Trust,
		Traces:   s.Traces,
		Registry: runtime.Registry,
		Adapters: runtime.Adapters,
		Config:   s.Config,
	}
	result, err := pipeline.Run(ctx, domainReq, cfgOverride)
	if err != nil {
		return nil, err
	}

	out := ToOrchestrateResponseV1(result, req.SessionID)
	if err := orchestratorv1.ValidateOrchestrateResponseV1(out); err != nil {
		return nil, fmt.Errorf("invalid orchestrate response: %w", err)
	}
	return out, nil
}

// GetTraceBundle returns trace + metrics for GET /v1/traces/:id.
func (s *Service) GetTraceBundle(traceID string) (map[string]interface{}, bool) {
	trace, err := s.Traces.Get(traceID)
	if err != nil || trace == nil {
		return nil, false
	}
	v1 := TraceToV1(*trace)
	metrics := orchestratorv1.SummarizeOrchestrationTrace(&v1, 0)
	return map[string]interface{}{
		"trace":           v1,
		"metrics_summary": metrics,
	}, true
}

// ListTrust returns trust records for GET /v1/trust.
func (s *Service) ListTrust(domain string) ([]TrustRecord, error) {
	if domain != "" {
		return s.Trust.ListByDomain(domain)
	}
	return s.Trust.ListAll()
}

// ListTraceIDs returns recent trace ids.
func (s *Service) ListTraceIDs(limit int) ([]string, error) {
	return s.Traces.ListTraceIDs(limit)
}
