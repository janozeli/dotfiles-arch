package main

import (
	"fmt"
	"strings"
)

func evaluate(step Step, force, systemdOK bool, completed map[string]bool, env []string) StepResult {
	// 1. Missing verify script
	if c, ok := step.(Checkable); ok && !c.HasVerify() {
		return makeResult(step, StatusError, "missing verify script")
	}

	// 2. Requires systemd
	if step.RequiresSystemd() && !systemdOK {
		return makeResult(step, StatusSkipped, "requires systemd")
	}

	// 3. Dependencies not met
	if deps := step.DependsOn(); len(deps) > 0 {
		var missing []string
		for _, d := range deps {
			if !completed[d] {
				missing = append(missing, d)
			}
		}
		if len(missing) > 0 {
			return makeResult(step, StatusSkipped, fmt.Sprintf("dependency not met: %s", strings.Join(missing, ", ")))
		}
	}

	// 4. Pre-verify (skip if --force)
	if !force && step.Verify(env) {
		return makeResult(step, StatusSuccess, "already completed")
	}

	// 5. Execute
	result := step.Execute(env)
	if result.Status != StatusSuccess {
		return result
	}

	// 6. Post-verify
	if !step.Verify(env) {
		return StepResult{
			ID:         step.ID(),
			Name:       step.Name(),
			Status:     StatusError,
			Message:    "verification failed after execution",
			StartedAt:  result.StartedAt,
			FinishedAt: nowISO(),
			DurationS:  result.DurationS,
			ExitCode:   result.ExitCode,
		}
	}

	return result
}

func makeResult(step Step, status Status, message string) StepResult {
	ts := nowISO()
	return StepResult{
		ID:         step.ID(),
		Name:       step.Name(),
		Status:     status,
		Message:    message,
		StartedAt:  ts,
		FinishedAt: ts,
	}
}
