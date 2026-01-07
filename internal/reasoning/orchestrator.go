package reasoning

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gaiol/internal/models"
	"gaiol/internal/models/adapters"
	"gaiol/internal/uaip"
)

// Orchestrator handles parallel execution of multiple LLMs
type Orchestrator struct {
	Router        *models.ModelRouter
	PromptBuilder *PromptBuilder
	RAG           *RAGManager
	SessionID     string        // NEW: Store current session ID for events
	OnEvent       EventCallback // NEW: Callback for live updates
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(router *models.ModelRouter, pb *PromptBuilder) *Orchestrator {
	return &Orchestrator{
		Router:        router,
		PromptBuilder: pb,
	}
}

// ExecuteStep runs parallel models for a given step and handles routing
func (o *Orchestrator) ExecuteStep(ctx context.Context, step ReasoningStep, sharedContext string, modelIDs []string, config SessionConfig) ([]ModelOutput, error) {
	var wg sync.WaitGroup
	// Handle dynamic model selection if "auto" is requested or no models provided
	effectiveModelIDs := modelIDs
	if len(modelIDs) == 0 || (len(modelIDs) == 1 && modelIDs[0] == "auto") {
		strategy := models.StrategyHighestQuality

		// Map priority profile to routing strategy
		switch config.PriorityProfile {
		case "speed":
			strategy = models.StrategyLowestCost // Assuming lower cost models are faster or we want to save budget for speed
		case "balanced":
			strategy = models.StrategyBalanced
		}

		routeConfig := models.RoutingConfig{
			Strategy: strategy,
			Task:     step.TaskType,
			MaxCost:  config.BudgetLimit, // Use budget as cost constraint
		}
		// Default to logic if not specified
		if routeConfig.Task == "" {
			routeConfig.Task = models.TaskAnalyze
		}

		model, err := o.Router.Route(routeConfig)
		if err != nil {
			fmt.Printf("⚠️ Dynamic routing failed for task %s, falling back to default: %v\n", step.TaskType, err)
			effectiveModelIDs = []string{"anthropic/claude-3-5-sonnet"} // Safe fallback
		} else {
			effectiveModelIDs = []string{string(model.ID)}
			fmt.Printf("🎯 Dynamic routing selected %s for task %s\n", model.ID, step.TaskType)
		}
	}

	outputChan := make(chan ModelOutput, len(effectiveModelIDs))

	// Wrap the objective with shared context
	prompt := step.Objective
	if o.RAG != nil {
		if augmented, docs, err := o.RAG.AugmentPrompt(ctx, prompt); err == nil {
			prompt = augmented
			if len(docs) > 0 {
				o.emitEvent(ctx, EventRAG, docs)
			}
		}
	}
	wrappedPrompt := o.PromptBuilder.WrapWithContext(prompt, sharedContext)

	for _, modelID := range effectiveModelIDs {
		wg.Add(1)
		go func(mid string) {
			defer wg.Done()

			// Add timeout per model query (reduced to 20s for faster feedback)
			mctx, cancel := context.WithTimeout(ctx, 20*time.Second)
			defer cancel()

			output, err := o.executeModelWithRetry(mctx, mid, wrappedPrompt, 2)
			if err != nil {
				// Don't send error outputs - let the fallback logic handle completely failed steps
				fmt.Printf("⚠️  Model %s failed: %v\n", mid, err)
				return
			}

			outputChan <- output
		}(modelID)
	}

	// Wait for all models to finish or context to be cancelled
	wg.Wait()
	close(outputChan)

	var results []ModelOutput
	for out := range outputChan {
		results = append(results, out)
	}

	// If NO successful results, try fallbacks
	if len(results) == 0 && len(effectiveModelIDs) > 0 {
		fmt.Println("🚨 All models failed for step. Attempting fallback to guardian model...")
		fallbackModel := "anthropic/claude-3-5-sonnet"
		output, err := o.executeModelWithRetry(ctx, fallbackModel, wrappedPrompt, 1)
		if err == nil {
			output.ModelName += " (Fallback)"
			return []ModelOutput{output}, nil
		}
		fmt.Printf("❌ Fallback guardian model also failed: %v\n", err)

		// Try Ollama as local fallback with EXTENDED timeout
		fmt.Println("🔄 Trying local Ollama fallback...")

		// Create NEW context with longer timeout for local models (don't inherit 20s timeout!)
		ollamaCtx, ollamaCancel := context.WithTimeout(context.Background(), 120*time.Second)
		defer ollamaCancel()

		ollamaOutput, ollamaErr := o.tryOllamaFallback(ollamaCtx, wrappedPrompt)
		if ollamaErr == nil {
			fmt.Println("✅ Ollama fallback succeeded!")
			return []ModelOutput{ollamaOutput}, nil
		}
		fmt.Printf("❌ Ollama fallback also failed: %v\n", ollamaErr)
	}

	// EMERGENCY FALLBACK: If still no results, create placeholder
	if len(results) == 0 {
		fmt.Println("⚠️  No results from any model. Creating emergency fallback response.")
		results = append(results, ModelOutput{
			ModelID:   "emergency-fallback",
			ModelName: "System Fallback",
			Response:  fmt.Sprintf("[All AI models unavailable]\n\nStep: %s\nObjective: %s\n\nThis step could not be completed because all AI models are currently unavailable due to API rate limits or service issues. Please wait and try again.", step.Title, step.Objective),
			Scores:    MetricScores{Overall: 0.1},
			Timestamp: time.Now(),
		})
	}

	return results, nil
}

// tryOllamaFallback attempts to use local Ollama as a last resort
func (o *Orchestrator) tryOllamaFallback(ctx context.Context, prompt string) (ModelOutput, error) {
	// Try to import and use Ollama adapter
	ollamaAdapter := adapters.NewOllamaAdapter("")

	// Check if Ollama is available
	models, err := ollamaAdapter.CheckAvailability(ctx)
	if err != nil || len(models) == 0 {
		return ModelOutput{}, fmt.Errorf("ollama not available: %w", err)
	}

	// Use first available model
	modelName := models[0]
	fmt.Printf("🦙 Using local Ollama model: %s\n", modelName)

	// Create minimal UAIP request
	uaipReq := &uaip.UAIPRequest{
		Payload: uaip.Payload{
			Input: uaip.PayloadInput{
				Data:   prompt,
				Format: "text",
			},
			OutputRequirements: uaip.OutputRequirements{
				MaxTokens:   1000,
				Temperature: 0.7,
			},
		},
	}

	resp, err := ollamaAdapter.GenerateText(ctx, modelName, uaipReq)
	if err != nil {
		return ModelOutput{}, err
	}

	if !resp.Status.Success {
		return ModelOutput{}, fmt.Errorf("ollama error: %s", resp.Status.Message)
	}

	return ModelOutput{
		ModelID:    "ollama:" + modelName,
		ModelName:  "Ollama " + modelName + " (Local)",
		Response:   resp.Result.Data,
		TokensUsed: resp.Result.TokensUsed,
		Cost:       0.0, // Local is free!
		Timestamp:  time.Now(),
		Scores:     MetricScores{Overall: 0.7}, // Decent quality
	}, nil
}

// emitEvent sends an event to the callback
func (o *Orchestrator) emitEvent(ctx context.Context, et EventType, payload interface{}) {
	if o.OnEvent != nil {
		o.OnEvent(ReasoningEvent{
			Type:      et,
			SessionID: o.SessionID,
			Payload:   payload,
			Timestamp: time.Now(),
		})
	}
}

// executeModelWithRetry handles a single model query with retries
func (o *Orchestrator) executeModelWithRetry(ctx context.Context, modelID, prompt string, maxRetries int) (ModelOutput, error) {
	var lastErr error
	var lastResponse QueryResponse
	for i := 0; i <= maxRetries; i++ {
		startTime := time.Now()
		// Use reasoning's QueryModel wrapper
		qm := NewQueryModel(o.Router)

		resp, err := qm.QueryFull(ctx, modelID, prompt)
		latency := time.Since(startTime).Milliseconds()

		if err == nil {
			return ModelOutput{
				ModelID:    modelID,
				ModelName:  modelID,
				Response:   resp.Response,
				TokensUsed: resp.Usage.TotalTokens,
				Cost:       resp.EstimatedCost,
				LatencyMs:  latency,
				Timestamp:  time.Now(),
			}, nil
		}

		// If error but response has data (error message), store it
		if resp.Response != "" {
			lastResponse = resp
		}

		lastErr = err
		// Exponential backoff or simple sleep could be added here
		time.Sleep(time.Duration(i*100) * time.Millisecond)
	}

	// If we have a response with error data, return it as ModelOutput instead of error
	// This allows error messages to be displayed to users
	if lastResponse.Response != "" {
		return ModelOutput{
			ModelID:    modelID,
			ModelName:  modelID + " (Error)",
			Response:   lastResponse.Response,
			TokensUsed: lastResponse.Usage.TotalTokens,
			Cost:       lastResponse.EstimatedCost,
			LatencyMs:  0,
			Timestamp:  time.Now(),
			Scores:     MetricScores{Overall: 0.0}, // Mark as low quality
		}, nil
	}

	return ModelOutput{}, lastErr
}

// Query is a convenience method for a single model query
func (o *Orchestrator) Query(ctx context.Context, modelID, prompt string) (string, error) {
	qm := NewQueryModel(o.Router)
	return qm.Query(ctx, modelID, prompt)
}
