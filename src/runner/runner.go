package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/pterm/pterm"
)

// RunOpts holds configuration for a runner execution.
type RunOpts struct {
	Force   bool
	Verbose bool
}

// Runner orchestrates wave-based parallel execution of units.
type Runner struct {
	executor Executor
	units    []Unit
	dag      *DAG
	manifest *Manifest
	opts     RunOpts

	// skippedUnits tracks units whose verify passed (skip remaining stages).
	skippedUnits map[string]bool
	mu           sync.Mutex
}

// NewRunner creates a runner wired to execute the given units.
func NewRunner(units []Unit, dag *DAG, manifest *Manifest, executor Executor, opts RunOpts) *Runner {
	return &Runner{
		executor:     executor,
		units:        units,
		dag:          dag,
		manifest:     manifest,
		opts:         opts,
		skippedUnits: make(map[string]bool),
	}
}

// Run executes all waves in order. Tasks within a wave run in parallel.
func (r *Runner) Run(ctx context.Context) ([]TaskResult, bool) {
	var allResults []TaskResult
	aborted := false

	for i, wave := range r.dag.Waves() {
		if r.opts.Verbose {
			pterm.DefaultSection.WithLevel(2).Println(fmt.Sprintf("Wave %d (%d tasks)", i+1, len(wave)))
		}

		results := r.executeWave(ctx, wave)
		allResults = append(allResults, results...)

		// Update manifest with results.
		for _, result := range results {
			r.manifest.UpdateTask(result.ID, result)
		}

		// Check for critical failures — do not start next wave.
		if r.hasCriticalFailure(results) {
			aborted = true
			break
		}
	}

	return allResults, aborted
}

func (r *Runner) executeWave(ctx context.Context, wave []*TaskNode) []TaskResult {
	results := make([]TaskResult, len(wave))
	var wg sync.WaitGroup

	for i, node := range wave {
		wg.Add(1)
		go func(idx int, n *TaskNode) {
			defer wg.Done()
			results[idx] = r.executeNode(ctx, n)
		}(i, node)
	}

	wg.Wait()

	// Print results after the wave completes (serialized, no interleaving).
	for _, result := range results {
		r.printTaskResult(result)
	}

	return results
}

func (r *Runner) executeNode(ctx context.Context, node *TaskNode) TaskResult {
	// Check if this unit was already skipped (verify passed earlier).
	r.mu.Lock()
	if r.skippedUnits[node.UnitID] {
		r.mu.Unlock()
		return TaskResult{
			ID:        node.GlobalID,
			UnitID:    node.UnitID,
			Status:    StatusSkipped,
			Message:   "already completed",
			StartedAt: nowISO(),
		}
	}
	r.mu.Unlock()

	// Execute the task.
	result := r.executor.Execute(node.Task, node.Stage.ExecutionPlan, ctx)
	result.ID = node.GlobalID
	result.UnitID = node.UnitID

	// Verify-skip logic: if this is a "verify" stage task and it passed,
	// mark the entire unit as skipped (unless --force).
	if !r.opts.Force && strings.EqualFold(node.StageName, "verify") && result.Status == StatusSuccess {
		r.mu.Lock()
		r.skippedUnits[node.UnitID] = true
		r.mu.Unlock()
		result.Message = "already completed"
	}

	return result
}

func (r *Runner) hasCriticalFailure(results []TaskResult) bool {
	for _, result := range results {
		if result.Status != StatusError {
			continue
		}
		// Find the unit to check if it's critical.
		for _, unit := range r.units {
			if unit.ID == result.UnitID && unit.Critical {
				pterm.Error.Printfln("[ABORT] Critical unit '%s' failed: %s", unit.Name, result.Message)
				return true
			}
		}
	}
	return false
}

func (r *Runner) printTaskResult(result TaskResult) {
	switch result.Status {
	case StatusSuccess:
		if result.Message == "already completed" {
			pterm.FgGray.Printfln("  ✓ %s (already completed)", result.ID)
		} else if r.opts.Verbose {
			pterm.Success.Printfln("  %s (%ds)", result.ID, result.Duration)
		}
	case StatusSkipped:
		pterm.FgGray.Printfln("  ○ %s (skipped: %s)", result.ID, result.Message)
	case StatusError:
		pterm.Error.Printfln("  ✗ %s: %s", result.ID, result.Message)
	}
}
