package main

// Status represents the outcome of a step evaluation.
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

// StepResult holds the outcome of evaluating a single step.
type StepResult struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     Status `json:"status"`
	Message    string `json:"message"`
	StartedAt  string `json:"started_at"`
	FinishedAt string `json:"finished_at"`
	DurationS  int    `json:"duration_s"`
	ExitCode   *int   `json:"exit_code"`
}

// Step is the interface every install step must implement.
type Step interface {
	ID() string
	Name() string
	Critical() bool
	DependsOn() []string
	Verify(env []string) bool
	Execute(env []string) StepResult
}

// Checkable is optionally implemented by steps that rely on external files.
type Checkable interface {
	HasVerify() bool
}

var registry []Step

func register(s Step) Step {
	registry = append(registry, s)
	return s
}

// Steps returns all registered steps in registration order.
func Steps() []Step {
	return registry
}
