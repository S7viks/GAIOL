package llm

import (
	"context"
	"testing"

	"gaiol/internal/models"
	"gaiol/internal/uaip"
)

type nilResponseAdapter struct{}

func (nilResponseAdapter) Name() string                         { return "nil-test" }
func (nilResponseAdapter) Provider() string                     { return "nil-test" }
func (nilResponseAdapter) SupportedTasks() []models.TaskType      { return nil }
func (nilResponseAdapter) RequiresAuth() bool                   { return false }
func (nilResponseAdapter) GetCapabilities() models.ModelCapabilities { return models.ModelCapabilities{} }
func (nilResponseAdapter) GetCost() models.CostInfo             { return models.CostInfo{} }
func (nilResponseAdapter) HealthCheck() error                   { return nil }
func (nilResponseAdapter) GenerateText(_ context.Context, _ string, _ *uaip.UAIPRequest) (*uaip.UAIPResponse, error) {
	return nil, nil
}

func TestModelAdapterBridge_NilResponse(t *testing.T) {
	t.Parallel()
	b := &ModelAdapterBridge{Provider: "test", Adapter: nilResponseAdapter{}}
	_, err := b.Generate(context.Background(), GenerateParams{
		Model:    "m1",
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for nil adapter response")
	}
}
