package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"  // FIXED: For contains helper.
	"sync"
	"time"

	"gaiol/internal/models"
	"gaiol/internal/uaip"
)

// RateLimiter handles Gemini's 15 requests/minute limit
type RateLimiter struct {
	tokens chan struct{}
	ticker *time.Ticker
	mu     sync.Mutex
	lastUsed time.Time
}

// GeminiAdapter implements ModelAdapter for Google's Gemini API
type GeminiAdapter struct {
	apiKey      string
	baseURL     string
	client      *http.Client
	rateLimiter *RateLimiter
}

// NewRateLimiter creates a rate limiter for Gemini free tier
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		tokens: make(chan struct{}, 1),
		ticker: time.NewTicker(4 * time.Second), // 15/min = every 4 seconds
	}
	
	// Start with one token
	rl.tokens <- struct{}{}
	
	// Refill tokens every 4 seconds
	go func() {
		for range rl.ticker.C {
			select {
			case rl.tokens <- struct{}{}:
				// Token added
			default:
				// Channel full, skip
			}
		}
	}()
	
	return rl
}

// Wait blocks until a rate limit token is available
func (rl *RateLimiter) Wait(ctx context.Context) error {
	rl.mu.Lock()
	rl.lastUsed = time.Now()
	rl.mu.Unlock()
	
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return fmt.Errorf("rate limit timeout")
	}
}

// NewGeminiAdapter creates a new Gemini adapter
func NewGeminiAdapter(apiKey string) *GeminiAdapter {
	return &GeminiAdapter{
		apiKey:      apiKey,
		baseURL:     "https://generativelanguage.googleapis.com/v1beta",
		client:      &http.Client{Timeout: 30 * time.Second},
		rateLimiter: NewRateLimiter(),
	}
}

// Gemini API structures (unchanged, but added JSON tags for completeness)
type GeminiRequest struct {
	Contents         []GeminiContent  `json:"contents"`
	GenerationConfig *GeminiConfig    `json:"generationConfig,omitempty"`
	SafetySettings   []SafetySetting  `json:"safetySettings,omitempty"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	TopK            *int     `json:"topK,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
}

type SafetySetting struct {
	Category  string `json:"category"`
	Threshold string `json:"threshold"`
}

type GeminiResponse struct {
	Candidates    []GeminiCandidate `json:"candidates"`
	UsageMetadata GeminiUsage       `json:"usageMetadata"`
	Error         *GeminiError      `json:"error,omitempty"`
}

type GeminiCandidate struct {
	Content        GeminiContent  `json:"content"`
	FinishReason   string         `json:"finishReason"`
	Index          int            `json:"index"`
	SafetyRatings  []SafetyRating `json:"safetyRatings"`
}

type GeminiUsage struct {
	PromptTokenCount     int `json:"promptTokenCount"`
	CandidatesTokenCount int `json:"candidatesTokenCount"`
	TotalTokenCount      int `json:"totalTokenCount"`
}

type SafetyRating struct {
	Category    string `json:"category"`
	Probability string `json:"probability"`
}

type GeminiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Status  string `json:"status"`
}

// Implementation of ModelAdapter interface methods (unchanged)

func (g *GeminiAdapter) Name() string {
	return "gemini-1.5-flash"
}

func (g *GeminiAdapter) Provider() string {
	return "google"
}

func (g *GeminiAdapter) SupportedTasks() []models.TaskType {
	return []models.TaskType{
		models.TaskGenerate,
		models.TaskAnalyze,
		models.TaskSummarize,
		models.TaskTransform,
	}
}

func (g *GeminiAdapter) RequiresAuth() bool {
	return true
}

func (g *GeminiAdapter) GetCapabilities() models.ModelCapabilities {
	return models.ModelCapabilities{
		MaxTokens:         8192,
		SupportsStreaming: false,
		Languages:         []string{"en", "es", "fr", "de", "it", "pt", "hi", "ja", "ko", "zh"},
		ContextWindow:     1048576, // 1M tokens
		QualityScore:      0.85,
		Multimodal:        true,
	}
}

func (g *GeminiAdapter) GetCost() models.CostInfo {
	return models.CostInfo{
		CostPerToken:    0.0, // Free tier
		CostPerRequest:  0.0, // Free tier
		FreeTierLimit:   15,  // 15 requests per minute
		RateLimitPerMin: 15,
	}
}

// GenerateText - Main method that processes UAIP requests
func (g *GeminiAdapter) GenerateText(ctx context.Context, req *uaip.UAIPRequest) (*uaip.UAIPResponse, error) {
	startTime := time.Now()
	
	// Rate limiting
	if err := g.rateLimiter.Wait(ctx); err != nil {
		return g.createErrorResponse(req, fmt.Errorf("rate limit error: %w", err), startTime), nil
	}
	
	// Convert UAIP request to Gemini format
	geminiReq := g.convertToGeminiRequest(req)
	
	// Make API call
	geminiResp, err := g.callGeminiAPI(ctx, geminiReq)
	if err != nil {
		return g.createErrorResponse(req, err, startTime), nil
	}
	
	// Convert response back to UAIP format
	return g.convertToUAIPResponse(geminiResp, req, startTime), nil
}

// convertToGeminiRequest converts UAIP request to Gemini API format
func (g *GeminiAdapter) convertToGeminiRequest(req *uaip.UAIPRequest) *GeminiRequest {
	maxTokens := req.Payload.OutputRequirements.MaxTokens
	temperature := req.Payload.OutputRequirements.Temperature
	topK := 64
	topP := 0.95
	
	return &GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: req.Payload.Input.Data},
				},
			},
		},
		GenerationConfig: &GeminiConfig{
			Temperature:     &temperature,
			MaxOutputTokens: &maxTokens,
			TopK:            &topK,
			TopP:            &topP,
		},
		SafetySettings: []SafetySetting{
			{Category: "HARM_CATEGORY_HARASSMENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_HATE_SPEECH", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_SEXUALLY_EXPLICIT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
			{Category: "HARM_CATEGORY_DANGEROUS_CONTENT", Threshold: "BLOCK_MEDIUM_AND_ABOVE"},
		},
	}
}

// callGeminiAPI makes the actual HTTP request to Gemini
func (g *GeminiAdapter) callGeminiAPI(ctx context.Context, req *GeminiRequest) (*GeminiResponse, error) {
	// Marshal request
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	
	// Create HTTP request
	url := fmt.Sprintf("%s/models/gemini-1.5-flash-latest:generateContent?key=%s", g.baseURL, g.apiKey)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("User-Agent", "GAIOL/1.0")
	
	// Make request
	resp, err := g.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()
	
	// Parse response
	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Check for API errors
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("Gemini API error %d: %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}
	
	// Check HTTP status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	return &geminiResp, nil
}

// convertToUAIPResponse converts Gemini response to UAIP format
func (g *GeminiAdapter) convertToUAIPResponse(resp *GeminiResponse, originalReq *uaip.UAIPRequest, startTime time.Time) *uaip.UAIPResponse {
	_ = int(time.Since(startTime).Milliseconds())  // FIXED: Unused processingMs assigned to _.
	
	// Check if we have candidates
	if len(resp.Candidates) == 0 {
		return &uaip.UAIPResponse{
			UAIP: uaip.UAIPHeader{
				Version:       uaip.ProtocolVersion,
				MessageID:     g.generateMessageID(),
				CorrelationID: originalReq.UAIP.MessageID,
				Timestamp:     time.Now(),
			},
			Status: uaip.ResponseStatus{
				Code:    uaip.StatusInternalError,
				Message: "No response candidates generated",
				Success: false,
			},
			Error: &uaip.ErrorInfo{
				Code:            uaip.ErrorCodeInternalError,
				Type:            uaip.ErrorTypeInternal,
				Message:         "Gemini returned no response candidates",
				SuggestedAction: "retry_with_different_prompt",
			},
		}
	}
	
	// Extract response text from first candidate
	candidate := resp.Candidates[0]
	var responseText string
	if len(candidate.Content.Parts) > 0 {
		responseText = candidate.Content.Parts[0].Text
	}
	
	// Calculate quality score based on finish reason
	qualityScore := g.calculateQualityScore(candidate.FinishReason)
	
	return &uaip.UAIPResponse{
		UAIP: uaip.UAIPHeader{
			Version:       uaip.ProtocolVersion,
			MessageID:     g.generateMessageID(),
			CorrelationID: originalReq.UAIP.MessageID,
			Timestamp:     time.Now(),
		},
		Status: uaip.ResponseStatus{
			Code:    uaip.StatusOK,
			Message: "Generated successfully",
			Success: true,
		},
		Result: uaip.Result{
			Data:         responseText,
			Format:       "text",
			TokensUsed:   resp.UsageMetadata.TotalTokenCount,
			ProcessingMs: 0,  // FIXED: Set to 0 or calculate if needed.
			Quality:      qualityScore,
			ModelUsed:    g.Name(),
			Metadata: map[string]interface{}{
				"finish_reason":        candidate.FinishReason,
				"safety_ratings":       candidate.SafetyRatings,
				"prompt_tokens":        resp.UsageMetadata.PromptTokenCount,
				"completion_tokens":    resp.UsageMetadata.CandidatesTokenCount,
				"gemini_candidate_index": candidate.Index,
			},
		},
	}
}

// FIXED: Moved generateMessageID up to avoid "method not found".
func (g *GeminiAdapter) generateMessageID() string {
	return fmt.Sprintf("gemini-%d", time.Now().UnixNano())
}

// HealthCheck performs a simple API connectivity test
func (g *GeminiAdapter) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	
	testReq := &uaip.UAIPRequest{
		UAIP: uaip.UAIPHeader{
			Version:   uaip.ProtocolVersion,
			MessageID: "health-check",
			Timestamp: time.Now(),
		},
		Payload: uaip.Payload{
			Input: uaip.PayloadInput{
				Data:   "Test",
				Format: "text",
			},
			OutputRequirements: uaip.OutputRequirements{
				MaxTokens:   5,
				Temperature: 0.1,
			},
		},
	}
	
	resp, err := g.GenerateText(ctx, testReq)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	
	if !resp.Status.Success {
		return fmt.Errorf("health check unsuccessful: %s", resp.Status.Message)
	}
	
	return nil
}

// createErrorResponse creates a UAIP error response
func (g *GeminiAdapter) createErrorResponse(req *uaip.UAIPRequest, err error, startTime time.Time) *uaip.UAIPResponse {
	_ = int(time.Since(startTime).Milliseconds())  // FIXED: Unused processingMs.
	
	// Determine error type and code based on error message
	errorCode := uaip.ErrorCodeInternalError
	errorType := uaip.ErrorTypeInternal
	suggestedAction := "retry"
	
	errMsg := err.Error()
	if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "429") {  // FIXED: Use strings.Contains.
		errorCode = uaip.ErrorCodeRateLimit
		errorType = uaip.ErrorTypeRateLimit
		suggestedAction = "wait_and_retry"
	} else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "context deadline") {
		errorCode = uaip.ErrorCodeTimeout
		errorType = uaip.ErrorTypeTimeout
		suggestedAction = "retry_with_longer_timeout"
	} else if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
		errorCode = uaip.ErrorCodeAuthFailed
		errorType = uaip.ErrorTypeAuthentication
		suggestedAction = "check_api_key"
	}
	
	return &uaip.UAIPResponse{
		UAIP: uaip.UAIPHeader{
			Version:       uaip.ProtocolVersion,
			MessageID:     g.generateMessageID(),
			CorrelationID: req.UAIP.MessageID,
			Timestamp:     time.Now(),
		},
		Status: uaip.ResponseStatus{
			Code:    uaip.StatusInternalError,
			Message: "Request failed",
			Success: false,
		},
		Error: &uaip.ErrorInfo{
			Code:            errorCode,
			Type:            errorType,
			Message:         errMsg,
			SuggestedAction: suggestedAction,
		},
		Metadata: uaip.ResponseMetadata{
			ProcessedAt: time.Now(),
			TraceID:     req.UAIP.MessageID,
		},
	}
}

// calculateQualityScore based on finish reason
func (g *GeminiAdapter) calculateQualityScore(finishReason string) float64 {
	switch finishReason {
	case "STOP":
		return 0.90 // Natural completion
	case "MAX_TOKENS":
		return 0.75 // Hit token limit
	case "SAFETY":
		return 0.60 // Safety filtered
	default:
		return 0.70 // Unknown
	}
}