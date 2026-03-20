package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// stepsDir is resolved once in main and set via initBashSteps.
var stepsDir string

// BashStep implements Step by running bash scripts.
type BashStep struct {
	id        string
	name      string
	script    string // e.g. "yay.sh"
	critical  bool
	dependsOn []string
}

func (b *BashStep) ID() string         { return b.id }
func (b *BashStep) Name() string        { return b.name }
func (b *BashStep) Critical() bool      { return b.critical }
func (b *BashStep) DependsOn() []string { return b.dependsOn }

func (b *BashStep) verifyPath() string {
	base := b.script[:len(b.script)-len(".sh")]
	return filepath.Join(stepsDir, base+".verify.sh")
}

func (b *BashStep) scriptPath() string {
	return filepath.Join(stepsDir, b.script)
}

// HasVerify returns true if the verify script exists on disk.
func (b *BashStep) HasVerify() bool {
	_, err := os.Stat(b.verifyPath())
	return err == nil
}

// Verify runs the verify script and returns true on exit 0.
func (b *BashStep) Verify(env []string) bool {
	cmd := exec.Command("bash", b.verifyPath())
	cmd.Env = env
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// Execute runs the action script and returns a StepResult.
func (b *BashStep) Execute(env []string) StepResult {
	startedAt := nowISO()
	t0 := time.Now()

	cmd := exec.Command("bash", b.scriptPath())
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	elapsed := int(time.Since(t0).Seconds())
	exitCode := 0

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
		return StepResult{
			ID:         b.id,
			Name:       b.name,
			Status:     StatusError,
			Message:    fmt.Sprintf("exit code %d", exitCode),
			StartedAt:  startedAt,
			FinishedAt: nowISO(),
			DurationS:  elapsed,
			ExitCode:   &exitCode,
		}
	}

	return StepResult{
		ID:         b.id,
		Name:       b.name,
		Status:     StatusSuccess,
		Message:    fmt.Sprintf("%s concluído", b.name),
		StartedAt:  startedAt,
		FinishedAt: nowISO(),
		DurationS:  elapsed,
		ExitCode:   &exitCode,
	}
}

// initBashSteps sets the stepsDir for all registered BashSteps.
func initBashSteps(dir string) {
	stepsDir = dir
}
