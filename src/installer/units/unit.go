package units

import (
	"fmt"
	"path/filepath"
)

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

// Unit is a pure data struct parsed from a unit.yaml file.
// It declares what the unit needs; the runner decides how to fulfill it.
type Unit struct {
	ID       string            `yaml:"id"`
	Name     string            `yaml:"name"`
	Critical bool              `yaml:"critical"`
	Stages   []Stage           `yaml:"stages"`
	Actions  map[string]Action `yaml:"actions"`
	Flags    map[string]any    `yaml:"flags"`
	Dir      string            `yaml:"-"`
}

// Stage is a named block within a unit's pipeline (e.g., verify, setup).
type Stage struct {
	Name          string        `yaml:"name"`
	Tasks         []Task        `yaml:"tasks"`
	ExecutionPlan ExecutionPlan `yaml:"executor"`
}

// Task is the atomic unit of execution — a single executable file.
type Task struct {
	Name      string   `yaml:"task"`
	DependsOn []string `yaml:"depends_on"`
	Path      string   `yaml:"-"`
}

// ExecutionPlan declares how a stage wants to be executed.
type ExecutionPlan struct {
	Shell   string      `yaml:"shell"`
	Timeout int         `yaml:"timeout"`
	Sudo    bool        `yaml:"sudo"`
	Env     []string    `yaml:"env"`
	Retry   RetryPolicy `yaml:"retry"`
}

// RetryPolicy defines retry behavior for a stage.
type RetryPolicy struct {
	MaxAttempts int `yaml:"max_attempts"`
	Delay       int `yaml:"delay"`
}

// Action is an optional script that can be invoked on demand (e.g., teardown).
type Action struct {
	Task        string `yaml:"task"`
	Description string `yaml:"description"`
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

// ResolveUnitPaths fills Dir and Task.Path for all units based on unitsDir.
func ResolveUnitPaths(unitsDir string, all []Unit) {
	for i := range all {
		all[i].Dir = filepath.Join(unitsDir, all[i].ID)
		for j := range all[i].Stages {
			for k := range all[i].Stages[j].Tasks {
				t := &all[i].Stages[j].Tasks[k]
				t.Path = filepath.Join(all[i].Dir, "tasks", t.Name)
			}
		}
	}
}
