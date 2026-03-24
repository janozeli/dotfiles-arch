package units

import "fmt"

// Status represents the outcome of a task evaluation.
type Status string

const (
	StatusSuccess Status = "success"
	StatusWarning Status = "warning"
	StatusSkipped Status = "skipped"
	StatusError   Status = "error"
)

func (s Status) Completed() bool {
	return s == StatusSuccess || s == StatusWarning
}

// Unit declares what a unit needs; the runner decides how to fulfill it.
type Unit struct {
	ID       string
	Name     string
	Critical bool
	Stages   []Stage
	Actions  map[string]Action
	Flags    map[string]any
	Dir      string
}

// Stage is a named block within a unit's pipeline.
type Stage struct {
	Name          string
	Tasks         []Task
	ExecutionPlan ExecutionPlan
}

// Task is the atomic unit of execution.
type Task struct {
	Name      string
	DependsOn []string
	Path      string
	UnitID    string
}

// ExecutionPlan declares how a stage wants to be executed.
type ExecutionPlan struct {
	Shell   string
	Timeout int
	Sudo    bool
	Env     []string
	Retry   RetryPolicy
}

// RetryPolicy defines retry behavior for a stage.
type RetryPolicy struct {
	MaxAttempts int
	Delay       int
}

// Action is an optional task that can be invoked on demand (e.g., teardown).
type Action struct {
	Task        string
	Description string
}

// TaskResult holds the outcome of executing a single task.
type TaskResult struct {
	ID        string `json:"id"`
	UnitID    string `json:"unit_id"`
	Status    Status `json:"status"`
	Message   string `json:"message"`
	ExitCode  *int   `json:"exit_code"`
	StartedAt string `json:"started_at"`
	Duration  int    `json:"duration_s"`
	LogPath   string `json:"log_path,omitempty"`
}

var (
	requiredStages  = []string{"verify", "setup"}
	requiredActions = []string{"teardown"}
)

// Validate checks that a Unit meets all invariants.
func (u Unit) Validate() error {
	if u.ID == "" {
		return fmt.Errorf("missing required field: id")
	}
	if u.Name == "" {
		return fmt.Errorf("missing required field: name")
	}

	stageNames := make(map[string]bool, len(u.Stages))
	for _, s := range u.Stages {
		stageNames[s.Name] = true
	}
	for _, required := range requiredStages {
		if !stageNames[required] {
			return fmt.Errorf("missing required stage: %q", required)
		}
	}

	for _, required := range requiredActions {
		if _, ok := u.Actions[required]; !ok {
			return fmt.Errorf("missing required action: %q", required)
		}
	}

	return nil
}
