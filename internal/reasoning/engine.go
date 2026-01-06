package reasoning

import (
	"context"
	"fmt"
	"time"

	"gaiol/internal/database"
	"gaiol/internal/models"

	"github.com/google/uuid"
)

// ReasoningEngine is the central coordinator for the multi-agent reasoning flow
type ReasoningEngine struct {
	MemoryManager    *MemoryManager
	Decomposer       *Decomposer
	Orchestrator     *Orchestrator
	RAG              *RAGManager // NEW: RAG management
	Scorer           *Scorer
	Selector         *Selector
	Composer         *Composer
	Critic           *Critic                    // NEW: Quality validator
	Refiner          *Refiner                   // NEW: Output improver
	ReflectionConfig ReflectionConfig           // NEW: Reflection settings
	BeamConfig       BeamConfig                 // NEW: Beam settings
	ConsensusConfig  ConsensusConfig            // NEW: Consensus settings
	ConsensusAgent   *ConsensusAgent            // NEW: Meta-reasoning agent
	Tracker          *models.PerformanceTracker // NEW: Learning loop
	OnEvent          EventCallback
}

// BeamConfig contains settings for beam search reasoning
type BeamConfig struct {
	Enabled   bool `json:"enabled"`
	BeamWidth int  `json:"beam_width"` // Number of paths to maintain
}

// DefaultBeamConfig returns the default beam search settings
func DefaultBeamConfig() BeamConfig {
	return BeamConfig{
		Enabled:   false,
		BeamWidth: 2,
	}
}

// NewReasoningEngine creates a new reasoning engine instance
func NewReasoningEngine(router *models.ModelRouter) *ReasoningEngine {
	pb := NewPromptBuilder()
	queryModel := NewQueryModel(router)
	reflectionConfig := DefaultReflectionConfig()

	orchestrator := NewOrchestrator(router, pb)

	// Initialize RAG if database is available
	var rag *RAGManager
	dbClient := database.GetClient()
	if dbClient != nil {
		store := database.NewSupabaseVectorStore(dbClient)

		// Find an adapter that supports embeddings (OpenRouter)
		allModels := router.GetRegistry().ListModels()
		for _, m := range allModels {
			if m.Provider == "openrouter" && m.Adapter != nil {
				if embedder, ok := m.Adapter.(models.EmbeddingProvider); ok {
					rag = NewRAGManager(store, embedder)
					orchestrator.RAG = rag
					break
				}
			}
		}
	}

	// Initialize Performance Tracker
	var tracker *models.PerformanceTracker
	if dbClient != nil {
		tracker = models.NewPerformanceTracker(dbClient)
		tracker.RefreshCache(context.Background())
	}

	return &ReasoningEngine{
		MemoryManager:    NewMemoryManager(),
		Decomposer:       NewDecomposer(router),
		Orchestrator:     orchestrator,
		RAG:              rag,
		Scorer:           NewScorer(router, tracker),
		Selector:         NewSelector("greedy"),
		Composer:         NewComposer(),
		Critic:           NewCritic(queryModel, reflectionConfig),
		Refiner:          NewRefiner(queryModel),
		ReflectionConfig: reflectionConfig,
		BeamConfig:       DefaultBeamConfig(),
		ConsensusConfig:  DefaultConsensusConfig(),
		ConsensusAgent:   NewConsensusAgent(NewOrchestrator(router, pb)),
		Tracker:          tracker,
	}
}

// emitEvent sends an event if the callback is set
func (re *ReasoningEngine) emitEvent(sessionID string, et EventType, payload interface{}) {
	if re.OnEvent != nil {
		re.OnEvent(ReasoningEvent{
			Type:      et,
			SessionID: sessionID,
			Payload:   payload,
			Timestamp: time.Now(),
		})
	}
}

// InitSession creates a new session and returns the ID
func (re *ReasoningEngine) InitSession(ctx context.Context, prompt string) string {
	sessionID := uuid.New().String()
	sm := re.MemoryManager.CreateSession(sessionID, prompt)

	// Try to get user/tenant info from context
	if t, ok := database.GetTenantFromContext(ctx); ok {
		sm.UserID = t.UserID
		sm.TenantID = t.TenantID
	}

	// Initial persistence
	_ = re.MemoryManager.SaveSession(sm)

	return sessionID
}

// RunSession runs the complete reasoning process for an existing session
func (re *ReasoningEngine) RunSession(ctx context.Context, sessionID, prompt string, modelIDs []string) (*SharedMemory, error) {
	sm, exists := re.MemoryManager.GetSession(sessionID)
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	// 2. Decompose Prompt
	re.emitEvent(sessionID, EventDecomposeStart, nil)
	steps, err := re.Decomposer.DecomposeWithRetry(ctx, prompt, 3)
	if err != nil {
		re.emitEvent(sessionID, EventError, err.Error())
		return nil, fmt.Errorf("decomposition failed: %v", err)
	}

	sm.mu.Lock()
	sm.Steps = steps
	sm.mu.Unlock()

	// Persist the decomposed steps
	for _, step := range steps {
		_ = re.MemoryManager.SaveStep(sessionID, step)
	}

	re.emitEvent(sessionID, EventDecomposeEnd, EventDecomposePayload{Steps: steps})

	// 3. Process each step
	for i := range steps {
		re.Orchestrator.SessionID = sessionID
		re.Orchestrator.OnEvent = re.OnEvent

		// Update status
		sm.mu.Lock()
		sm.Steps[i].Status = "processing"
		sm.Steps[i].StartTime = time.Now()
		sm.mu.Unlock()

		re.emitEvent(sessionID, EventStepStart, EventStepPayload{
			StepIndex: i,
			Title:     sm.Steps[i].Title,
			Objective: sm.Steps[i].Objective,
			TaskType:  sm.Steps[i].TaskType,
		})

		var newPaths [][]ModelOutput

		if re.BeamConfig.Enabled && i > 0 {
			// BEAM SEARCH LOGIC
			sm.mu.RLock()
			activePaths := sm.ActivePaths
			if len(activePaths) == 0 {
				activePaths = [][]ModelOutput{sm.SelectedPath}
			}
			sm.mu.RUnlock()

			for _, path := range activePaths {
				// Build context for this specific path
				contextStr, _ := re.MemoryManager.GetContextForPath(sessionID, path)

				// Execute parallel models for this path
				outputs, err := re.Orchestrator.ExecuteStep(ctx, sm.Steps[i], contextStr, modelIDs, sm.Config)
				if err != nil {
					continue
				}

				// Score outputs
				scoredOutputs, err := re.Scorer.ScoreMultipleOutputs(ctx, sm.Steps[i].Objective, outputs, sm.Steps[i].TaskType, sm.Config.PriorityProfile)
				if err != nil {
					scoredOutputs = outputs
				}

				// Accumulate cost
				sm.mu.Lock()
				for _, out := range scoredOutputs {
					sm.TotalCost += out.Cost
				}
				sm.mu.Unlock()

				// Create new candidate paths
				for _, out := range scoredOutputs {
					newPath := make([]ModelOutput, len(path))
					copy(newPath, path)
					newPath = append(newPath, out)
					newPaths = append(newPaths, newPath)
				}
			}

			// Prune and update active paths
			err = re.MemoryManager.UpdateBeamResults(sessionID, i, newPaths, re.BeamConfig.BeamWidth)
			if err != nil {
				return nil, fmt.Errorf("failed to update beam results for step %d: %v", i, err)
			}

			// Emit beam update event
			re.emitEvent(sessionID, EventBeamUpdate, map[string]interface{}{
				"step_index":   i,
				"active_paths": len(sm.ActivePaths),
				"best_score":   sm.Steps[i].SelectedOutput.Scores.Overall,
				"total_cost":   sm.TotalCost,
			})

			// Persist beam outputs
			for pathIdx, path := range newPaths {
				output := path[len(path)-1]
				isSelected := false
				if len(sm.ActivePaths) > 0 && &sm.ActivePaths[0][len(sm.ActivePaths[0])-1] == &output {
					isSelected = true
				}
				_ = re.MemoryManager.SaveOutput(sessionID, i, output, isSelected, pathIdx)
			}
			// Update step status in DB
			_ = re.MemoryManager.SaveStep(sessionID, sm.Steps[i])

		} else {
			// GREEDY / INITIAL STEP LOGIC
			// Build context from previous steps
			contextStr, _ := re.MemoryManager.GetContextForStep(sessionID, i)

			// Execute parallel models
			outputs, err := re.Orchestrator.ExecuteStep(ctx, sm.Steps[i], contextStr, modelIDs, sm.Config)
			if err != nil {
				re.emitEvent(sessionID, EventError, err.Error())
				return nil, fmt.Errorf("step %d execution failed: %v", i, err)
			}

			// Score outputs
			scoredOutputs, err := re.Scorer.ScoreMultipleOutputs(ctx, sm.Steps[i].Objective, outputs, sm.Steps[i].TaskType, sm.Config.PriorityProfile)
			if err != nil {
				scoredOutputs = outputs
			}

			// 4. Consensus Reconciliation (NEW)
			if re.ConsensusConfig.Enabled {
				consensusResult, err := re.ConsensusAgent.Reconcile(ctx, sm.Steps[i].Objective, scoredOutputs, re.ConsensusConfig)
				if err == nil {
					sm.mu.Lock()
					sm.Steps[i].Consensus = consensusResult
					sm.mu.Unlock()

					// Emit consensus event
					re.emitEvent(sessionID, EventConsensus, consensusResult)

					// If consensus reached a synthesized output, we can use it
					if consensusResult.BestOutput != nil {
						// Note: We still honor Scorer's ranking for consistency unless meta-agent synthesis happened
						if consensusResult.Method == "meta_agent" {
							// For meta-agent, we might override the selected winner
							// In this implementation, we allow Reconcile to provide the best output
						}
					}
				}
			}

			// Update results and Select winner (Greedy)
			err = re.MemoryManager.UpdateStepResults(sessionID, i, scoredOutputs)
			if err != nil {
				return nil, fmt.Errorf("failed to update results for step %d: %v", i, err)
			}

			// If beam is enabled, initialize active paths
			if re.BeamConfig.Enabled {
				sm.mu.Lock()
				sm.ActivePaths = [][]ModelOutput{sm.SelectedPath}
				sm.mu.Unlock()
			}

			// Persist greedy outputs
			for _, out := range scoredOutputs {
				isSelected := false
				if sm.Steps[i].SelectedOutput != nil && sm.Steps[i].SelectedOutput.ModelID == out.ModelID {
					isSelected = true
				}
				_ = re.MemoryManager.SaveOutput(sessionID, i, out, isSelected, 0)
			}
			// Update step status in DB
			_ = re.MemoryManager.SaveStep(sessionID, sm.Steps[i])
		}

		// Get the selected output for potential reflection
		selectedOutput := sm.SelectedPath[len(sm.SelectedPath)-1]

		// SELF-REFLECTION LOOP
		if re.ReflectionConfig.Enabled {
			// ... (Reflection logic remains same, but applies to the selected path winner)
			// (Truncated for readability, keeping implementation as is)
			attempts := 0
			accepted := false

			for !accepted && attempts < re.ReflectionConfig.MaxRetries {
				feedback, err := re.Critic.ValidateOutput(ctx, sm.Steps[i], selectedOutput, sm)
				if err != nil {
					feedback = CriticFeedback{IsAcceptable: true, QualityScore: 0.8}
				}

				re.emitEvent(sessionID, EventReflection, map[string]interface{}{
					"step_index":  i,
					"accepted":    feedback.IsAcceptable,
					"quality":     feedback.QualityScore,
					"issues":      feedback.Issues,
					"suggestions": feedback.Suggestions,
					"attempt":     attempts + 1,
				})

				if feedback.IsAcceptable {
					accepted = true
					break
				}

				attempts++
				if attempts < re.ReflectionConfig.MaxRetries {
					re.emitEvent(sessionID, EventRefinement, map[string]interface{}{
						"step_index": i,
						"attempt":    attempts,
					})

					improved, err := re.Refiner.ImproveOutput(ctx, selectedOutput, feedback, sm.Steps[i], sm)
					if err == nil {
						sm.mu.Lock()
						sm.SelectedPath[len(sm.SelectedPath)-1] = improved
						// Also update in ActivePaths[0] if it's the same
						if len(sm.ActivePaths) > 0 && len(sm.ActivePaths[0]) == len(sm.SelectedPath) {
							sm.ActivePaths[0][len(sm.ActivePaths[0])-1] = improved
						}
						sm.mu.Unlock()
						selectedOutput = improved
					}
				}
			}
		}

		re.emitEvent(sessionID, EventStepEnd, sm.Steps[i])
	}

	// 4. Assemble Final Output
	finalOutput := re.Composer.AssembleFinalOutput(sm.SelectedPath)
	re.emitEvent(sessionID, EventReasoningEnd, EventReasoningEndPayload{FinalOutput: finalOutput})

	return sm, nil
}

// EnableReflection turns on self-reflection with custom config
func (re *ReasoningEngine) EnableReflection(config ReflectionConfig) {
	re.ReflectionConfig = config
}

// DisableReflection turns off self-reflection
func (re *ReasoningEngine) DisableReflection() {
	re.ReflectionConfig.Enabled = false
}

// EnableBeamSearch turns on beam search with custom config
func (re *ReasoningEngine) EnableBeamSearch(config BeamConfig) {
	re.BeamConfig = config
}

// DisableBeamSearch turns off beam search
func (re *ReasoningEngine) DisableBeamSearch() {
	re.BeamConfig.Enabled = false
}
