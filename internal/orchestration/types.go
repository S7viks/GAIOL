package orchestration

import (
	"time"

	orchestratorv1 "gaiol/internal/gaiol/orchestratorcontract/v1"
)

// ChatMessage is one turn in an orchestration request.
type ChatMessage struct {
	Role    string
	Content string
	Name    string
}

// TaskConstraints mirror TS OrchestrationRequest.constraints.
type TaskConstraints struct {
	MaxCostUsd          *float64
	MaxLatencyMsPerCall *int
	MaxParallelCalls    *int
	Temperature         *float64
	MaxOutputTokens     *int
}

// OrchestrationRequest is the domain request passed to the pipeline.
type OrchestrationRequest struct {
	TraceID      string
	SessionHint  string
	Domain       string
	TaskKind     string
	Objective    string
	Messages     []ChatMessage
	Constraints  *TaskConstraints
	ExplorePaths *bool
	BeamWidth    *int
}

// SubtaskSpec is one decomposed subtask.
type SubtaskSpec struct {
	ID                   string
	ParentID             string
	Title                string
	Description          string
	TaskKind             string
	RequiredCapabilities []string
}

// DecompositionResult is the output of decompose().
type DecompositionResult struct {
	Subtasks  []SubtaskSpec
	Rationale string
}

// ModelRegistryEntry describes a routable model in the pool.
type ModelRegistryEntry struct {
	ModelID       string
	ProviderID    string
	RemoteName    string
	Capabilities  []string
	CostIndex     float64
	LatencyPriorMs int
	AccuracyPrior  *float64
	Available     bool
}

// ModelCallUsage holds token/cost usage for one call.
type ModelCallUsage struct {
	PromptTokens     *int
	CompletionTokens *int
	CostUsd          *float64
}

// ModelCallResult is one provider generation result.
type ModelCallResult struct {
	ModelID    string
	ProviderID string
	Text       string
	LatencyMs  int64
	Usage      *ModelCallUsage
	Error      string
}

// TrustRoundEntryTrace is one trust update row in a subtask trace.
type TrustRoundEntryTrace struct {
	ModelID       string
	ProviderID    string
	Domain        string
	Role          string
	Prior         BetaTrust
	AfterDecay    BetaTrust
	Posterior     BetaTrust
	PriorMean     float64
	PosteriorMean float64
	Decay         float64
	Strength      float64
	Signal        float64
	Explanation   string
	Persisted     bool
}

// TrustRoundTrace captures trust movement for one subtask.
type TrustRoundTrace struct {
	ConsensusMode          string
	WinnerModelID          string
	SubtaskID              string
	Decay                  float64
	StrengthWinner         float64
	StrengthParticipant    float64
	ConsensusTrustExponent *float64
	Entries                []TrustRoundEntryTrace
}

// RoutingExplanationTrace explains model selection.
type RoutingExplanationTrace struct {
	DiversityRationale string
	CandidatePoolSize  int
	BeamWidth          int
	ModelRankSnapshot  []RankedModel
}

// PathCandidateTrace is one beam path candidate.
type PathCandidateTrace struct {
	PathID      string
	ModelID     string
	ProviderID  string
	Score       float64
	Kept        bool
	TextPreview string
}

// BeamPruneTrace records beam pruning.
type BeamPruneTrace struct {
	BeamWidth        int
	KeptPathIDs      []string
	DiscardedPathIDs []string
}

// PathExplorationTrace is beam search metadata.
type PathExplorationTrace struct {
	Candidates    []PathCandidateTrace
	Pruning       BeamPruneTrace
	WinningPathID string
}

// SubtaskExecutionTrace is the per-subtask execution record.
type SubtaskExecutionTrace struct {
	SubtaskID          string
	RoutedModelIDs     []string
	Calls              []ModelCallResult
	Scores             map[string]float64
	ChosenModelID      string
	ConsensusText      string
	TrustRound         *TrustRoundTrace
	RoutingExplanation *RoutingExplanationTrace
	PathExploration    *PathExplorationTrace
}

// OrchestrationTrace is the full run trace.
type OrchestrationTrace struct {
	TraceID       string
	Domain        string
	Decomposition DecompositionResult
	Subtasks      []SubtaskExecutionTrace
	StartedAt     string
	FinishedAt    string
}

// TrustUpdateEvent is emitted when trust is persisted (ABTC mode).
type TrustUpdateEvent struct {
	TraceID                string
	SessionHint            string
	Domain                 string
	ModelID                string
	ProviderID             string
	Distribution           BetaTrust
	UpdatedAt              string
	SubtaskID              string
	PriorDistribution      BetaTrust
	AfterDecayDistribution BetaTrust
	PriorMean              float64
	PosteriorMean          float64
	Decay                  float64
	Strength               float64
	Signal                 float64
	Role                   string
	Explanation            string
}

// OrchestrationResult is the pipeline output.
type OrchestrationResult struct {
	Trace        OrchestrationTrace
	Answer       string
	TrustUpdates []TrustUpdateEvent
	TotalRetries int
}

// OrchestratorConfig is default + per-request overrides.
type OrchestratorConfig struct {
	ConsensusMode        string
	BeamWidth            int
	MaxParallelCalls     int
	MaxCostUsdPerRequest *float64
	ABTC                 ABTCConfig
	Retry                RetryConfig
	StaticWeights        map[string]float64
}

// ABTCConfig holds ABTC hyperparameters.
type ABTCConfig struct {
	Decay                  float64
	Strength               float64
	ParticipantStrength    float64
	ConsensusTrustExponent float64
}

// RetryConfig controls provider call retries.
type RetryConfig struct {
	Retries      int
	BaseDelayMs  int
}

// DefaultConfig matches orchestrator/src/api/server.ts defaults.
func DefaultConfig() OrchestratorConfig {
	return OrchestratorConfig{
		ConsensusMode:    "abtc",
		BeamWidth:        2,
		MaxParallelCalls: 3,
		MaxCostUsdPerRequest: float64Ptr(5),
		ABTC: ABTCConfig{
			Decay:                  0.15,
			Strength:               2,
			ParticipantStrength:    1.2,
			ConsensusTrustExponent: 1.5,
		},
		Retry: RetryConfig{Retries: 2, BaseDelayMs: 50},
	}
}

func float64Ptr(v float64) *float64 { return &v }

// RequestFromV1 maps wire request to domain request.
func RequestFromV1(v *orchestratorv1.OrchestrateRequestV1) OrchestrationRequest {
	req := OrchestrationRequest{
		TraceID:     v.TraceID,
		SessionHint: v.SessionID,
		Domain:      v.Domain,
		TaskKind:    v.TaskKind,
		Objective:   v.Objective,
		ExplorePaths: v.ExplorePaths,
		BeamWidth:    v.BeamWidth,
	}
	for _, m := range v.Messages {
		req.Messages = append(req.Messages, ChatMessage{Role: m.Role, Content: m.Content, Name: m.Name})
	}
	if v.Constraints != nil {
		req.Constraints = &TaskConstraints{
			MaxCostUsd:          v.Constraints.MaxCostUsd,
			MaxLatencyMsPerCall: v.Constraints.MaxLatencyMsPerCall,
			MaxParallelCalls:    v.Constraints.MaxParallelCalls,
			Temperature:         v.Constraints.Temperature,
			MaxOutputTokens:     v.Constraints.MaxOutputTokens,
		}
	}
	return req
}

func nowRFC3339() string { return time.Now().UTC().Format(time.RFC3339Nano) }
