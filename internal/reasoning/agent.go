package reasoning

import (
	"context"
	"fmt"
	"time"
)

// AgentRole defines the agent's purpose
type AgentRole string

const (
	RolePlanner     AgentRole = "planner"
	RoleExecutor    AgentRole = "executor"
	RoleCritic      AgentRole = "critic"
	RoleResearcher  AgentRole = "researcher"
	RoleSynthesizer AgentRole = "synthesizer"
)

// Agent represents an autonomous AI agent
type Agent struct {
	ID           string
	Name         string
	Role         AgentRole
	ModelID      string
	SystemPrompt string
	Memory       *AgentMemory
	WorldModel   *WorldModel // NEW: Access to global knowledge
	CreatedAt    time.Time
}

// AgentMemory stores agent context
type AgentMemory struct {
	ShortTerm  []string
	Facts      map[string]string
	LastActive time.Time
}

// NewAgent creates a new agent with optional world model access
func NewAgent(role AgentRole, modelID string, worldModel *WorldModel) *Agent {
	return &Agent{
		ID:           fmt.Sprintf("agent-%s-%d", role, time.Now().UnixNano()),
		Name:         string(role),
		Role:         role,
		ModelID:      modelID,
		SystemPrompt: getSystemPrompt(role),
		Memory: &AgentMemory{
			ShortTerm: make([]string, 0),
			Facts:     make(map[string]string),
		},
		WorldModel: worldModel, // NEW
		CreatedAt:  time.Now(),
	}
}

// Execute runs the agent on a task
func (a *Agent) Execute(ctx context.Context, orchestrator *Orchestrator, task AgentTask) (*AgentOutput, error) {
	prompt := a.buildPrompt(task)

	a.Memory.ShortTerm = append(a.Memory.ShortTerm, fmt.Sprintf("Task: %s", task.Description))

	// Use orchestrator to execute with ANY available model
	response, err := orchestrator.Query(ctx, a.ModelID, prompt)
	if err != nil {
		return nil, fmt.Errorf("agent %s execution failed: %w", a.ID, err)
	}

	output := &AgentOutput{
		AgentID:   a.ID,
		AgentRole: a.Role,
		TaskID:    task.ID,
		Response:  response,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// NEW: Extract and store facts from response
	if a.WorldModel != nil && a.Role == RoleExecutor {
		// Only executors store facts (planners and critics don't learn new info)
		extracted := a.WorldModel.ExtractFacts(ctx, response, string(a.Role), task.ID)
		if len(extracted) > 0 {
			fmt.Printf("Agent learned %d new facts\n", len(extracted))
		}
	}

	a.Memory.ShortTerm = append(a.Memory.ShortTerm, fmt.Sprintf("Completed: %s", task.ID))
	a.Memory.LastActive = time.Now()

	return output, nil
}

// buildPrompt constructs prompt with role context AND world model facts
func (a *Agent) buildPrompt(task AgentTask) string {
	prompt := fmt.Sprintf("%s\n\n", a.SystemPrompt)

	// Add world model context if available
	if a.WorldModel != nil {
		worldContext := a.WorldModel.GetContext(context.Background(), task.Description, 5)
		if worldContext != "" {
			prompt += worldContext
			prompt += "Use this knowledge if relevant to the current task.\n\n"
		}
	}

	// Add agent's short-term memory
	if len(a.Memory.ShortTerm) > 0 {
		prompt += "Recent context:\n"
		start := max(0, len(a.Memory.ShortTerm)-3)
		for _, item := range a.Memory.ShortTerm[start:] {
			prompt += fmt.Sprintf("- %s\n", item)
		}
		prompt += "\n"
	}

	prompt += fmt.Sprintf("Current task: %s\n", task.Description)

	if task.Context != "" {
		prompt += fmt.Sprintf("\nContext from previous agents:\n%s\n", task.Context)
	}

	return prompt
}

// getSystemPrompt returns role-specific instructions
func getSystemPrompt(role AgentRole) string {
	prompts := map[AgentRole]string{
		RolePlanner: `You are a PLANNING agent. Your role:
1. Break down the user's goal into 2-3 concrete steps
2. Keep it SIMPLE - max 3 steps
3. Format as:
   STEP 1: [Action]
   STEP 2: [Action]
   STEP 3: [Action]
Be concise and actionable.`,

		RoleExecutor: `You are an EXECUTION agent. Your role:
1. Complete the specific task given to you
2. Provide a clear, concise answer
3. Be direct and factual
Focus on completing the task, nothing more.`,

		RoleCritic: `You are a VALIDATION agent. Your role:
1. Review the execution output
2. Check for errors or missing information
3. Provide verdict: APPROVE or NEEDS_REVISION
Be quick but thorough.`,
	}

	return prompts[role]
}

// AgentTask represents a task for an agent
type AgentTask struct {
	ID          string
	Description string
	Context     string
}

// AgentOutput represents an agent's response
type AgentOutput struct {
	AgentID   string
	AgentRole AgentRole
	TaskID    string
	Response  string
	Timestamp time.Time
	Metadata  map[string]interface{}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
