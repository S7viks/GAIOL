package adapters

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "regexp"  // Add this
    "strings"
    "time"
    
    "gaiol/internal/models"
    "gaiol/internal/uaip"
)

// OpenRouterAdapter implements ModelAdapter for OpenRouter API
type OpenRouterAdapter struct {
    modelName    string
    baseURL      string
    client       *http.Client
    rateLimiter  *RateLimiter
    apiKey       string
    freeModels   []string
}

// OpenRouter uses OpenAI-compatible format
type OpenRouterRequest struct {
    Model       string    `json:"model"`
    Messages    []Message `json:"messages"`
    MaxTokens   int       `json:"max_tokens,omitempty"`
    Temperature float64   `json:"temperature,omitempty"`
    TopP        float64   `json:"top_p,omitempty"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type OpenRouterResponse struct {
    ID      string   `json:"id"`
    Choices []Choice `json:"choices"`
    Usage   Usage    `json:"usage"`
    Error   *APIError `json:"error,omitempty"`
}

type Choice struct {
    Index   int `json:"index"`
    Message struct {
        Role    string                 `json:"role"`
        Content string                 `json:"content"`
        Extra   map[string]interface{} `json:"extra,omitempty"` // Add this field
    } `json:"message"`
    FinishReason string `json:"finish_reason"`
}

type Usage struct {
    PromptTokens     int `json:"prompt_tokens"`
    CompletionTokens int `json:"completion_tokens"`
    TotalTokens      int `json:"total_tokens"`
}

type APIError struct {
    Message string `json:"message"`
    Type    string `json:"type"`
    Code    string `json:"code"`
}

// NewOpenRouterAdapter creates a new OpenRouter adapter
func NewOpenRouterAdapter(modelName, apiKey string) *OpenRouterAdapter {
    if modelName == "" {
        modelName = "qwen/qwq-32b:free" // QwQ showed actual content in debug
    }
    
    // Reorder based on debug results - QwQ showed content
    freeModels := []string{
        "qwen/qwq-32b:free",              // This showed actual content in debug
        "z-ai/glm-4.5-air:free",          // Keep as fallback
        "deepseek/deepseek-r1:free",      // Keep as fallback
        "moonshotai/kimi-k2:free",        // Final fallback
    }
    
    return &OpenRouterAdapter{
        modelName:   modelName,
        baseURL:     "https://openrouter.ai/api/v1",
        client:      &http.Client{Timeout: 60 * time.Second},
        rateLimiter: NewRateLimiter(),
        apiKey:      apiKey,
        freeModels:  freeModels,
    }
}


func (o *OpenRouterAdapter) Name() string {
    return o.modelName
}

func (o *OpenRouterAdapter) Provider() string {
    return "openrouter"
}

func (o *OpenRouterAdapter) SupportedTasks() []models.TaskType {
    return []models.TaskType{
        models.TaskGenerate,
        models.TaskAnalyze,
        models.TaskSummarize,
        models.TaskTransform,
        models.TaskCode,
    }
}

func (o *OpenRouterAdapter) RequiresAuth() bool {
    return true
}

func (o *OpenRouterAdapter) GetCapabilities() models.ModelCapabilities {
    return models.ModelCapabilities{
        MaxTokens:         2048, // Good for free models
        SupportsStreaming: false,
        Languages:         []string{"en", "zh", "es", "fr", "de", "ja", "ko"},
        ContextWindow:     32768, // Varies by model
        QualityScore:      0.85,  // Free models are quite good
        Multimodal:        false,
    }
}

func (o *OpenRouterAdapter) GetCost() models.CostInfo {
    return models.CostInfo{
        CostPerToken:    0.0, // Free models
        CostPerRequest:  0.0, // Free models
        FreeTierLimit:   20,  // 20 requests per minute
        RateLimitPerMin: 20,
    }
}

func (o *OpenRouterAdapter) HealthCheck() error {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
    
    testReq := &uaip.UAIPRequest{
        UAIP: uaip.UAIPHeader{
            Version:   uaip.ProtocolVersion,
            MessageID: "health-check",
            Timestamp: time.Now(),
        },
        Payload: uaip.Payload{
            Input: uaip.PayloadInput{
                Data:   "Hello",
                Format: "text",
            },
            OutputRequirements: uaip.OutputRequirements{
                MaxTokens:   10,
                Temperature: 0.1,
            },
        },
    }
    
    resp, err := o.GenerateText(ctx, testReq)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    
    if !resp.Status.Success {
        return fmt.Errorf("health check unsuccessful: %s", resp.Status.Message)
    }
    
    return nil
}

func (o *OpenRouterAdapter) GenerateText(ctx context.Context, req *uaip.UAIPRequest) (*uaip.UAIPResponse, error) {
    startTime := time.Now()
    
    // Rate limiting (20 req/min for free models)
    if err := o.rateLimiter.Wait(ctx); err != nil {
        return o.createErrorResponse(req, fmt.Errorf("rate limit error: %w", err), startTime), nil
    }
    
    // Try primary model first, then fallbacks
    modelsToTry := []string{o.modelName}
    modelsToTry = append(modelsToTry, o.freeModels...)
    
    var lastErr error
    for i, modelName := range modelsToTry {
        if i > 0 {
            fmt.Printf("   Trying fallback model: %s\n", modelName)
        }
        
        orReq := o.convertToOpenRouterRequest(req, modelName)
        orResp, err := o.callOpenRouterAPI(ctx, orReq)
        
        if err == nil && len(orResp.Choices) > 0 {
            // Success!
            if i > 0 {
                o.modelName = modelName
                fmt.Printf("   ✅ Switched to working model: %s\n", modelName)
            }
            return o.convertToUAIPResponse(orResp, req, startTime), nil
        }
        
        lastErr = err
        // Don't fallback on auth errors
        if strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "unauthorized") {
            break
        }
    }
    
    return o.createErrorResponse(req, lastErr, startTime), nil
}

func (o *OpenRouterAdapter) convertToOpenRouterRequest(req *uaip.UAIPRequest, modelName string) *OpenRouterRequest {
    maxTokens := req.Payload.OutputRequirements.MaxTokens
    if maxTokens > 1000 {
        maxTokens = 1000 // Keep reasonable for free tier
    }
    
    // Fix: Ensure minimum tokens for content generation
    if maxTokens < 50 {
        maxTokens = 150 // Minimum for meaningful response
    }
    
    // Fix: Optimize prompt for DeepSeek models
    prompt := req.Payload.Input.Data
    if strings.Contains(modelName, "deepseek") {
        prompt = "Please respond concisely: " + prompt
    }
    
    return &OpenRouterRequest{
        Model: modelName,
        Messages: []Message{
            {
                Role:    "user",
                Content: prompt,
            },
        },
        MaxTokens:   maxTokens,
        Temperature: req.Payload.OutputRequirements.Temperature,
        TopP:        0.9,
    }
}

func (o *OpenRouterAdapter) callOpenRouterAPI(ctx context.Context, req *OpenRouterRequest) (*OpenRouterResponse, error) {
    jsonData, err := json.Marshal(req)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal request: %w", err)
    }
    
    url := o.baseURL + "/chat/completions"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("failed to create HTTP request: %w", err)
    }
    
    // OpenRouter headers
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", o.apiKey))
    httpReq.Header.Set("HTTP-Referer", "https://gaiol.ai") // Your app URL
    httpReq.Header.Set("X-Title", "GAIOL Universal AI Interoperability")
    
    resp, err := o.client.Do(httpReq)
    if err != nil {
        return nil, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
    
    var orResp OpenRouterResponse
    if err := json.NewDecoder(resp.Body).Decode(&orResp); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    // Check for API errors
    if orResp.Error != nil {
        return nil, fmt.Errorf("OpenRouter API error: %s", orResp.Error.Message)
    }
    
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
    }
    
    return &orResp, nil
}
func (o *OpenRouterAdapter) extractFromReasoningFields(choice Choice) string {
    // Try reasoning field first
    if reasoning, ok := choice.Message.Extra["reasoning"].(string); ok {
        if text := o.extractAnswerFromReasoning(reasoning); text != "" {
            return text
        }
    }
    
    // Try reasoning_details field
    if reasoningDetails, ok := choice.Message.Extra["reasoning_details"].([]interface{}); ok && len(reasoningDetails) > 0 {
        if detailMap, ok := reasoningDetails[0].(map[string]interface{}); ok {
            if text, ok := detailMap["text"].(string); ok {
                return o.extractAnswerFromReasoning(text)
            }
        }
    }
    
    return ""
}
func (o *OpenRouterAdapter) convertToUAIPResponse(resp *OpenRouterResponse, originalReq *uaip.UAIPRequest, startTime time.Time) *uaip.UAIPResponse {
    processingMs := int(time.Since(startTime).Milliseconds())
    
    if len(resp.Choices) == 0 {
        return o.createEmptyResponse(originalReq)
    }
    
    choice := resp.Choices[0]
    var responseText string
    
    // Enhanced parsing specifically for GLM and reasoning models
    content := strings.TrimSpace(choice.Message.Content)
    
    if content != "" {
        // Handle regular content
        responseText = o.cleanReasoningArtifacts(content)
    } else {
        // GLM often puts the response in reasoning field - check raw response
        fmt.Printf("DEBUG: GLM empty content, checking reasoning fields...\n")
        fmt.Printf("DEBUG: Full choice: %+v\n", choice)
        
        // Try to access reasoning from raw JSON - GLM specific
        if rawData, ok := choice.Message.Extra["reasoning"].(string); ok {
            fmt.Printf("DEBUG: Found reasoning field: %s\n", rawData[:min(100, len(rawData))])
            responseText = o.extractAnswerFromReasoning(rawData)
        }
        
        // If still empty, try other GLM-specific fields
        if responseText == "" {
            // Try accessing the reasoning field directly from the choice
            // GLM sometimes structures responses differently
            responseText = o.extractGLMSpecificResponse(choice)
        }
    }
    
    // Fallback if still empty
    if responseText == "" {
        responseText = o.generateFallbackResponse(choice, resp.Usage.TotalTokens)
    }
    
    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     o.generateMessageID(),
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
            TokensUsed:   resp.Usage.TotalTokens,
            ProcessingMs: processingMs,
            Quality:      0.85,
            ModelUsed:    o.modelName,
            Metadata: map[string]interface{}{
                "model_name":         o.modelName,
                "provider":           "openrouter",
                "finish_reason":      choice.FinishReason,
                "is_reasoning_model": true,
                "raw_choice":         choice,
            },
        },
    }
}
func (o *OpenRouterAdapter) extractGLMSpecificResponse(choice Choice) string {
    // GLM may structure the response differently
    // Try to extract from any field that contains meaningful text
    
    // Check if there's a reasoning field in the message itself
    if reasoningField, exists := choice.Message.Extra["reasoning"]; exists {
        if reasoning, ok := reasoningField.(string); ok && len(reasoning) > 10 {
            // Extract the final answer from reasoning
            return o.extractAnswerFromReasoning(reasoning)
        }
    }
    
    // Check for any other text fields GLM might use
    for key, value := range choice.Message.Extra {
        if strValue, ok := value.(string); ok && len(strValue) > 10 && len(strValue) < 1000 {
            fmt.Printf("DEBUG: Found text in field '%s': %s\n", key, strValue[:min(50, len(strValue))])
            if o.looksLikeResponse(strValue) {
                return strValue
            }
        }
    }
    
    return ""
}

// Helper to check if text looks like a valid response
func (o *OpenRouterAdapter) looksLikeResponse(text string) bool {
    text = strings.ToLower(text)
    // Avoid internal reasoning text, prefer actual answers
    return !strings.Contains(text, "let me think") && 
           !strings.Contains(text, "okay, so") &&
           !strings.Contains(text, "hmm") &&
           len(text) > 20 && len(text) < 500
}

// Add min helper function
func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
func (o *OpenRouterAdapter) generateFallbackResponse(choice Choice, totalTokens int) string {
    if choice.FinishReason == "length" {
        return "Response was truncated due to length limits. The model processed the request but exceeded token limits."
    }
    return fmt.Sprintf("Generated %d tokens but parsing needs adjustment. Finish reason: %s", totalTokens, choice.FinishReason)
}
func (o *OpenRouterAdapter) cleanReasoningArtifacts(content string) string {
    // Remove <think> tags and their content
    thinkRegex := regexp.MustCompile(`<think>.*?</think>`)
    cleaned := thinkRegex.ReplaceAllString(content, "")
    
    // Remove reasoning prefixes
    prefixes := []string{"<think>", "</think>", "Okay,", "Let me think"}
    for _, prefix := range prefixes {
        if strings.HasPrefix(cleaned, prefix) {
            cleaned = strings.TrimPrefix(cleaned, prefix)
        }
    }
    
    return strings.TrimSpace(cleaned)
}
// Extract final answer from reasoning text
func (o *OpenRouterAdapter) extractAnswerFromReasoning(reasoning string) string {
    // Look for common patterns where models conclude their reasoning
    reasoning = strings.TrimSpace(reasoning)
    
    // Split by common conclusion markers
    conclusionMarkers := []string{
        "So the answer is:",
        "Therefore:",
        "In conclusion:",
        "The answer is:",
        "Final answer:",
        "So:",
    }
    
    for _, marker := range conclusionMarkers {
        if idx := strings.LastIndex(reasoning, marker); idx != -1 {
            answer := strings.TrimSpace(reasoning[idx+len(marker):])
            if len(answer) > 0 && len(answer) < 500 { // Reasonable length
                return answer
            }
        }
    }
    
    // If no conclusion marker, try to extract the last meaningful sentence
    sentences := strings.Split(reasoning, ".")
    for i := len(sentences) - 1; i >= 0; i-- {
        sentence := strings.TrimSpace(sentences[i])
        if len(sentence) > 10 && len(sentence) < 200 {
            return sentence + "."
        }
    }
    
    // Fallback: return first 200 characters of reasoning
    if len(reasoning) > 200 {
        return reasoning[:200] + "... (reasoning model response)"
    }
    return reasoning
}
func (o *OpenRouterAdapter) createEmptyResponse(req *uaip.UAIPRequest) *uaip.UAIPResponse {
    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     o.generateMessageID(),
            CorrelationID: req.UAIP.MessageID,
            Timestamp:     time.Now(),
        },
        Status: uaip.ResponseStatus{
            Code:    uaip.StatusInternalError,
            Message: "No response from model",
            Success: false,
        },
        Error: &uaip.ErrorInfo{
            Code:            uaip.ErrorCodeInternalError,
            Type:            uaip.ErrorTypeInternal,
            Message:         "OpenRouter returned empty response",
            SuggestedAction: "try_different_model",
        },
    }
}

func (o *OpenRouterAdapter) createErrorResponse(req *uaip.UAIPRequest, err error, startTime time.Time) *uaip.UAIPResponse {
    errorCode := uaip.ErrorCodeInternalError
    errorType := uaip.ErrorTypeInternal
    suggestedAction := "retry"
    
    errMsg := err.Error()
    if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "429") {
        errorCode = uaip.ErrorCodeRateLimit
        errorType = uaip.ErrorTypeRateLimit
        suggestedAction = "wait_and_retry"
    } else if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
        errorCode = uaip.ErrorCodeAuthFailed
        errorType = uaip.ErrorTypeAuthentication
        suggestedAction = "check_api_key"
    }
    
    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     o.generateMessageID(),
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
            Message:         fmt.Sprintf("OpenRouter API issue: %s", errMsg),
            SuggestedAction: suggestedAction,
        },
        Metadata: uaip.ResponseMetadata{
            ProcessedAt: time.Now(),
            TraceID:     req.UAIP.MessageID,
        },
    }
}

func (o *OpenRouterAdapter) generateMessageID() string {
    return fmt.Sprintf("or-%d", time.Now().UnixNano())
}
