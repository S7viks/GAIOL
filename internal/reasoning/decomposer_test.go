package reasoning

import (
	"context"
	"testing"

	"gaiol/internal/models"
)

func TestDecomposerTaskTypeDetection(t *testing.T) {
	// Setup registry and router
	reg := models.NewRegistry(nil, nil)
	router := models.NewModelRouter(reg, nil)
	decomposer := NewDecomposer(router)

	prompt := "Write a python script to scrape a website and then analyze the common keywords."

	ctx := context.Background()
	steps, err := decomposer.DecomposePrompt(ctx, prompt)
	if err != nil {
		t.Fatalf("Decomposition failed: %v", err)
	}

	if len(steps) == 0 {
		t.Fatal("No steps produced")
	}

	foundCode := false
	foundAnalyze := false

	for _, step := range steps {
		t.Logf("Step: %s, TaskType: %s", step.Title, step.TaskType)
		if step.TaskType == models.TaskCode {
			foundCode = true
		}
		if step.TaskType == models.TaskAnalyze || step.TaskType == models.TaskLogic {
			foundAnalyze = true
		}
	}

	if !foundCode {
		t.Error("Expected at least one 'code' task")
	}
	if !foundAnalyze {
		t.Error("Expected at least one 'analyze' or 'logic' task")
	}
}
