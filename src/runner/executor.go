package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

// Executor defines how tasks are executed.
// The runner calls Execute without knowing the underlying strategy.
type Executor interface {
	Execute(task Task, plan ExecutionPlan, ctx context.Context) TaskResult
}

// DefaultExecutor runs tasks as local processes.
// It detects whether the file is executable (uses shebang) or falls back to bash.
type DefaultExecutor struct{}

func (e *DefaultExecutor) Execute(task Task, plan ExecutionPlan, ctx context.Context) TaskResult {
	startedAt := nowISO()
	t0 := time.Now()

	// Apply timeout from execution plan.
	if plan.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(plan.Timeout)*time.Second)
		defer cancel()
	}

	var stdout, stderr bytes.Buffer
	var lastErr error
	attempts := max(plan.Retry.MaxAttempts, 1)

	for attempt := range attempts {
		stdout.Reset()
		stderr.Reset()

		cmd := e.buildCmd(ctx, task, plan)
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr

		lastErr = cmd.Run()
		if lastErr == nil {
			break
		}

		if attempt < attempts-1 && plan.Retry.Delay > 0 {
			time.Sleep(time.Duration(plan.Retry.Delay) * time.Second)
		}
	}

	elapsed := int(time.Since(t0).Seconds())
	exitCode := 0

	if lastErr != nil {
		if exitErr, ok := lastErr.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
		return TaskResult{
			ID:        task.Name,
			Status:    StatusError,
			Message:   fmt.Sprintf("exit code %d: %s", exitCode, stderr.String()),
			ExitCode:  &exitCode,
			StartedAt: startedAt,
			Duration:  elapsed,
		}
	}

	return TaskResult{
		ID:        task.Name,
		Status:    StatusSuccess,
		Message:   stdout.String(),
		ExitCode:  &exitCode,
		StartedAt: startedAt,
		Duration:  elapsed,
	}
}

func (e *DefaultExecutor) buildCmd(ctx context.Context, task Task, plan ExecutionPlan) *exec.Cmd {
	var cmd *exec.Cmd

	switch {
	case plan.Shell != "":
		cmd = exec.CommandContext(ctx, plan.Shell, task.Path)
	case isExecutable(task.Path):
		cmd = exec.CommandContext(ctx, task.Path)
	default:
		cmd = exec.CommandContext(ctx, "bash", task.Path)
	}

	if plan.Sudo {
		cmd = exec.CommandContext(ctx, "sudo", append(cmd.Args)...)
	}

	cmd.Env = os.Environ()
	for _, env := range plan.Env {
		cmd.Env = append(cmd.Env, env)
	}

	return cmd
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Mode()&0111 != 0
}
