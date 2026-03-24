package runner

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"dotfiles/installer/units"

	"github.com/pterm/pterm"
)

// RunOpts holds configuration for a runner execution.
type RunOpts struct {
	Force   bool
	Verbose bool
}

// Runner orchestrates wave-based parallel execution of units.
type Runner struct {
	executor     Executor
	units        []units.Unit
	dag          *DAG
	manifest     *Manifest
	opts         RunOpts
	skippedUnits map[string]bool
	mu           sync.Mutex
}

// NewRunner creates a runner wired to execute the given units.
func NewRunner(u []units.Unit, dag *DAG, manifest *Manifest, executor Executor, opts RunOpts) *Runner {
	return &Runner{
		executor:     executor,
		units:        u,
		dag:          dag,
		manifest:     manifest,
		opts:         opts,
		skippedUnits: make(map[string]bool),
	}
}

// Run executes all waves in order. Tasks within a wave run in parallel.
func (r *Runner) Run(ctx context.Context) ([]units.TaskResult, bool) {
	var allResults []units.TaskResult
	aborted := false

	for i, wave := range r.dag.Waves() {
		if r.opts.Verbose {
			pterm.DefaultSection.WithLevel(2).Println(fmt.Sprintf("Wave %d (%d tasks)", i+1, len(wave)))
		}

		results := r.executeWave(ctx, wave)
		allResults = append(allResults, results...)

		for _, result := range results {
			r.manifest.UpdateTask(result.ID, result)
		}

		if r.hasCriticalFailure(results) {
			aborted = true
			break
		}
	}

	return allResults, aborted
}

func (r *Runner) executeWave(ctx context.Context, wave []*TaskNode) []units.TaskResult {
	results := make([]units.TaskResult, len(wave))
	var wg sync.WaitGroup

	for i, node := range wave {
		wg.Add(1)
		go func(idx int, n *TaskNode) {
			defer wg.Done()
			results[idx] = r.executeNode(ctx, n)
		}(i, node)
	}

	wg.Wait()

	for _, result := range results {
		r.printTaskResult(result)
	}

	return results
}

func (r *Runner) executeNode(ctx context.Context, node *TaskNode) units.TaskResult {
	r.mu.Lock()
	if r.skippedUnits[node.UnitID] {
		r.mu.Unlock()
		return units.TaskResult{
			ID:        node.GlobalID,
			UnitID:    node.UnitID,
			Status:    units.StatusSkipped,
			Message:   "already completed",
			StartedAt: NowISO(),
		}
	}
	r.mu.Unlock()

	result := r.executor.Execute(node.Task, node.Stage.ExecutionPlan, ctx)
	result.ID = node.GlobalID
	result.UnitID = node.UnitID

	if !r.opts.Force && strings.EqualFold(node.StageName, "verify") && result.Status == units.StatusSuccess {
		r.mu.Lock()
		r.skippedUnits[node.UnitID] = true
		r.mu.Unlock()
		result.Message = "already completed"
	}

	return result
}

func (r *Runner) hasCriticalFailure(results []units.TaskResult) bool {
	for _, result := range results {
		if result.Status != units.StatusError {
			continue
		}
		// Verify failures mean "needs setup", not a real failure.
		if result.Message == "needs setup" {
			continue
		}
		for _, u := range r.units {
			if u.ID == result.UnitID && u.Critical {
				pterm.Error.Printfln("[ABORT] Critical unit '%s' failed: %s", u.Name, result.Message)
				return true
			}
		}
	}
	return false
}

func (r *Runner) printTaskResult(result units.TaskResult) {
	switch result.Status {
	case units.StatusSuccess:
		if result.Message == "already completed" {
			pterm.FgGray.Printfln("  ✓ %s (already completed)", result.ID)
		} else if r.opts.Verbose {
			pterm.Success.Printfln("  %s (%ds)", result.ID, result.Duration)
		}
	case units.StatusSkipped:
		pterm.FgGray.Printfln("  ○ %s (skipped: %s)", result.ID, result.Message)
	case units.StatusError:
		pterm.Error.Printfln("  ✗ %s: %s", result.ID, result.Message)
	}
}
