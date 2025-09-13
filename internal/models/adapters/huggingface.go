package adapters

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
    "time"
	"io"
    "gaiol/internal/models"
    "gaiol/internal/uaip"
)

// HuggingFaceAdapter with better error handling and model fallbacks
// Updated to use new Inference Providers API (chat/completions) to avoid 404 issues with legacy API
type HuggingFaceAdapter struct {
    modelName     string
    baseURL       string
    client        *http.Client
    rateLimiter   *RateLimiter
    apiKey        string
    fallbackModels []string
}

// New structs for Chat Completions API
type ChatRequest struct {
    Model       string                 `json:"model"`
    Messages    []map[string]string    `json:"messages"`
    MaxTokens   int                    `json:"max_tokens,omitempty"`
    Temperature float64                `json:"temperature,omitempty"`
    TopP        float64                `json:"top_p,omitempty"`
    Stream      bool                   `json:"stream,omitempty"`
}

type ChatChoice struct {
    Message struct {
        Role    string `json:"role"`
        Content string `json:"content"`
    } `json:"message"`
    FinishReason string `json:"finish_reason"`
}

type ChatResponse struct {
    ID      string        `json:"id"`
    Choices []ChatChoice  `json:"choices"`
    Usage   map[string]int `json:"usage,omitempty"`
}

type ChatError struct {
    Error struct {
        Message string `json:"message"`
        Type    string `json:"type"`
        Param   string `json:"param,omitempty"`
        Code    string `json:"code,omitempty"`
    } `json:"error"`
}

// NewHuggingFaceAdapter with fallback models
func NewHuggingFaceAdapter(modelName, apiKey string) *HuggingFaceAdapter {
    if modelName == "" {
        modelName = "gpt2" // Start with reliable causal LM
    }
 
    // Fallback models in order of reliability (chat-compatible causal LMs)
    fallbacks := []string{
        "gpt2",
        "distilgpt2", 
        "microsoft/DialoGPT-small",
    }
 
    return &HuggingFaceAdapter{
        modelName:       modelName,
        baseURL:         "https://router.huggingface.co",
        client:          &http.Client{Timeout: 90 * time.Second},
        rateLimiter:     NewRateLimiter(),
        apiKey:          apiKey,
        fallbackModels: fallbacks,
    }
}

func (h *HuggingFaceAdapter) Name() string {
    return h.modelName
}

func (h *HuggingFaceAdapter) Provider() string {
    return "huggingface"
}

func (h *HuggingFaceAdapter) SupportedTasks() []models.TaskType {
    return []models.TaskType{
        models.TaskGenerate,
        models.TaskAnalyze,
        models.TaskSummarize,
        models.TaskClassify,
    }
}

func (h *HuggingFaceAdapter) RequiresAuth() bool {
    return true
}

func (h *HuggingFaceAdapter) GetCapabilities() models.ModelCapabilities {
    return models.ModelCapabilities{
        MaxTokens:        512, // Increased for new API
        SupportsStreaming: false,
        Languages:        []string{"en"},
        ContextWindow:    2048,
        QualityScore:     0.75,
        Multimodal:       false,
    }
}

func (h *HuggingFaceAdapter) GetCost() models.CostInfo {
    return models.CostInfo{
        CostPerToken:    0.0,
        CostPerRequest:  0.0,
        FreeTierLimit:   1000,
        RateLimitPerMin: 10,
    }
}

func (h *HuggingFaceAdapter) HealthCheck() error {
    // Simple health check using a minimal chat request
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()
 
    chatReq := &ChatRequest{
        Model:     h.modelName,
        Messages:  []map[string]string{{"role": "user", "content": "test"}},
        MaxTokens: 1,
    }
 
    _, err := h.callHuggingFaceAPI(ctx, chatReq)
    if err != nil {
        return fmt.Errorf("health check failed: %w", err)
    }
    return nil
}

func (h *HuggingFaceAdapter) GenerateText(ctx context.Context, req *uaip.UAIPRequest) (*uaip.UAIPResponse, error) {
    startTime := time.Now()
 
    // Rate limiting
    if err := h.rateLimiter.Wait(ctx); err != nil {
        return h.createErrorResponse(req, fmt.Errorf("rate limit error: %w", err), startTime), nil
    }
 
    // Try primary model first, then fallbacks
    modelsToTry := []string{h.modelName}
    modelsToTry = append(modelsToTry, h.fallbackModels...)
 
    var lastErr error
    for i, modelName := range modelsToTry {
        if i > 0 {
            fmt.Printf("   Trying fallback model: %s\n", modelName)
        }
 
        chatReq := h.convertToChatRequest(req, modelName)
        chatResp, err := h.callHuggingFaceAPI(ctx, chatReq)
 
        if err == nil && len(chatResp.Choices) > 0 {
            // Success! Update our model name if we used a fallback
            if i > 0 {
                h.modelName = modelName
                fmt.Printf("   ✅ Switched to working model: %s\n", modelName)
            }
            return h.convertToUAIPResponseFromChat(chatResp, req, startTime, modelName), nil
        }
 
        lastErr = err
        // Continue on most errors to try fallbacks (except rate limits, etc.)
        if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "401") {
            break // Don't fallback on auth/rate issues
        }
    }
 
    return h.createErrorResponse(req, lastErr, startTime), nil
}

func (h *HuggingFaceAdapter) convertToChatRequest(req *uaip.UAIPRequest, modelName string) *ChatRequest {
    maxTokens := req.Payload.OutputRequirements.MaxTokens
    if maxTokens > 512 {
        maxTokens = 512 // Limit for free tier stability
    }
 
    return &ChatRequest{
        Model:       modelName,
        Messages: []map[string]string{
            {
                "role":    "user",
                "content": req.Payload.Input.Data,
            },
        },
        MaxTokens:   maxTokens,
        Temperature: req.Payload.OutputRequirements.Temperature,
        TopP:        0.9,
    }
}

func (h *HuggingFaceAdapter) callHuggingFaceAPI(ctx context.Context, chatReq *ChatRequest) (ChatResponse, error) {
    jsonData, err := json.Marshal(chatReq)
    if err != nil {
        return ChatResponse{}, fmt.Errorf("failed to marshal request: %w", err)
    }
 
    url := h.baseURL + "/v1/chat/completions"
    httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return ChatResponse{}, fmt.Errorf("failed to create HTTP request: %w", err)
    }
 
    httpReq.Header.Set("Content-Type", "application/json")
    httpReq.Header.Set("User-Agent", "GAIOL/1.0")
 
    if h.apiKey != "" {
        httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", h.apiKey))
    }
 
    resp, err := h.client.Do(httpReq)
    if err != nil {
        return ChatResponse{}, fmt.Errorf("HTTP request failed: %w", err)
    }
    defer resp.Body.Close()
 
    if resp.StatusCode != http.StatusOK {
        var chatErr ChatError
        if json.NewDecoder(resp.Body).Decode(&chatErr) == nil && chatErr.Error.Message != "" {
            return ChatResponse{}, fmt.Errorf("HuggingFace API error (%d): %s", resp.StatusCode, chatErr.Error.Message)
        }
        bodyBytes, _ := io.ReadAll(resp.Body) // Need to import "io"
        return ChatResponse{}, fmt.Errorf("HTTP %d: %s (body: %s)", resp.StatusCode, resp.Status, string(bodyBytes))
    }
 
    var chatResp ChatResponse
    if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
        return ChatResponse{}, fmt.Errorf("failed to decode response: %w", err)
    }
 
    return chatResp, nil
}

func (h *HuggingFaceAdapter) convertToUAIPResponseFromChat(resp ChatResponse, originalReq *uaip.UAIPRequest, startTime time.Time, modelUsed string) *uaip.UAIPResponse {
    processingMs := int(time.Since(startTime).Milliseconds())

    if len(resp.Choices) == 0 {
        return h.createEmptyResponse(originalReq) // <-- FIXED: removed startTime
    }

    responseText := strings.TrimSpace(resp.Choices[0].Message.Content)
    if responseText == "" {
        responseText = "Generated response (empty content)"
    }

    tokensUsed := len(responseText) / 4 // Rough estimate

    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     h.generateMessageID(),
            CorrelationID: originalReq.UAIP.MessageID,
            Timestamp:     time.Now(),
        },
        Status: uaip.ResponseStatus{
            Code:    uaip.StatusOK,
            Message: "Generated successfully",
            Success: true,
        },
        Result: uaip.Result{
            Data:          responseText,
            Format:        "text",
            TokensUsed:    tokensUsed,
            ProcessingMs:  processingMs,
            Quality:       0.75,
            ModelUsed:     modelUsed,
            Metadata: map[string]interface{}{
                "model_name":  modelUsed,
                "provider":    "huggingface",
                "finish_reason": resp.Choices[0].FinishReason,
            },
        },
    }
}


func (h *HuggingFaceAdapter) createEmptyResponse(req *uaip.UAIPRequest) *uaip.UAIPResponse {
    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     h.generateMessageID(),
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
            Message:         "HuggingFace returned empty response",
            SuggestedAction: "try_different_model",
        },
    }
}


func (h *HuggingFaceAdapter) createErrorResponse(req *uaip.UAIPRequest, err error, startTime time.Time) *uaip.UAIPResponse {
    // processingMs := int(time.Since(startTime).Milliseconds()) // Remove if unused

    errorCode := uaip.ErrorCodeInternalError
    errorType := uaip.ErrorTypeInternal
    suggestedAction := "retry"

    errMsg := err.Error()
    if strings.Contains(errMsg, "model not found") || strings.Contains(errMsg, "404") {
        errorCode = uaip.ErrorCodeModelNotFound
        suggestedAction = "try_different_model"
    } else if strings.Contains(errMsg, "rate limit") || strings.Contains(errMsg, "429") {
        errorCode = uaip.ErrorCodeRateLimit
        errorType = uaip.ErrorTypeRateLimit
        suggestedAction = "wait_and_retry"
    } else if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "context deadline") {
        errorCode = uaip.ErrorCodeTimeout
        errorType = uaip.ErrorTypeTimeout
        suggestedAction = "retry_with_longer_timeout"
    } else if strings.Contains(errMsg, "unauthorized") || strings.Contains(errMsg, "401") {
        // If ErrorCodeAuthError is not defined, use ErrorCodeInternalError or define it in uaip
        errorCode = uaip.ErrorCodeAuthError
        suggestedAction = "check_api_key"
    }

    return &uaip.UAIPResponse{
        UAIP: uaip.UAIPHeader{
            Version:       uaip.ProtocolVersion,
            MessageID:     h.generateMessageID(),
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
            Message:         fmt.Sprintf("HuggingFace API issue: %s", errMsg),
            SuggestedAction: suggestedAction,
        },
        Metadata: uaip.ResponseMetadata{
            ProcessedAt: time.Now(),
            TraceID:     req.UAIP.MessageID,
        },
    }
}
func (h *HuggingFaceAdapter) generateMessageID() string {
    return fmt.Sprintf("hf-%d", time.Now().UnixNano())
}