package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

func resolveUnitsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get working directory: %v\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(cwd, "src", "runner", "units")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: units directory not found: %s\n", dir)
		os.Exit(1)
	}
	return dir
}

func listUnits(units []Unit) {
	data := pterm.TableData{{"#", "ID", "Name", "Critical", "Stages"}}
	for i, unit := range units {
		stages := make([]string, len(unit.Stages))
		for j, s := range unit.Stages {
			stages[j] = s.Name
		}
		crit := ""
		if unit.Critical {
			crit = "yes"
		}
		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			unit.ID,
			unit.Name,
			crit,
			strings.Join(stages, " → "),
		})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
}

func main() {
	force := flag.Bool("force", false, "ignore verify checks, run everything")
	list := flag.Bool("list", false, "list available units and exit")
	unit := flag.String("unit", "", "run only this unit (by ID)")
	dryRun := flag.Bool("dry-run", false, "generate manifest without executing")
	verbose := flag.Bool("verbose", false, "detailed output")
	action := flag.String("action", "", "run an action on a unit (format: action_name:unit_id)")
	flag.Parse()

	unitsDir := resolveUnitsDir()

	// Load units from filesystem.
	units, err := LoadUnits(unitsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *list {
		listUnits(units)
		return
	}

	// Filter to single unit if --unit.
	if *unit != "" {
		var filtered []Unit
		for _, u := range units {
			if u.ID == *unit {
				filtered = append(filtered, u)
			}
		}
		if len(filtered) == 0 {
			fmt.Fprintf(os.Stderr, "error: unknown unit '%s'\n", *unit)
			os.Exit(1)
		}
		units = filtered
	}

	// Handle --action.
	if *action != "" {
		parts := strings.SplitN(*action, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "error: --action format is action_name:unit_id\n")
			os.Exit(1)
		}
		runAction(units, parts[0], parts[1], unitsDir)
		return
	}

	// Build DAG.
	dag, err := NewDAG(units)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Create manifest.
	manifest := NewManifest(dag, units)
	manifestPath := filepath.Join(unitsDir, "manifest.json")

	if *dryRun {
		manifest.FinishedAt = nowISO()
		manifest.Status = "dry-run"
		if err := manifest.Save(manifestPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		pterm.Success.Printfln("Manifest written to %s (%d waves, %d units)", manifestPath, len(dag.Waves()), len(units))
		return
	}

	// Run.
	pterm.DefaultHeader.Println("dotfiles runner")
	fmt.Println()

	runner := NewRunner(units, dag, manifest, &DefaultExecutor{}, RunOpts{
		Force:   *force,
		Verbose: *verbose,
	})

	ctx := context.Background()
	results, aborted := runner.Run(ctx)

	// Save manifest.
	manifest.FinishedAt = nowISO()
	manifest.Status = manifest.DeriveStatus(aborted)
	_ = manifest.Save(manifestPath)

	// Save JSON log.
	logsDir := filepath.Join(filepath.Dir(unitsDir), "logs")
	tsFile := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(logsDir, tsFile+".json")
	saveRunLog(results, logPath, manifest)
	pruneOldLogs(logsDir, 3)

	// Summary.
	printNewSummary(results, logPath)

	if aborted {
		os.Exit(1)
	}
}

func runAction(units []Unit, actionName, unitID, unitsDir string) {
	var target *Unit
	for i := range units {
		if units[i].ID == unitID {
			target = &units[i]
			break
		}
	}
	if target == nil {
		fmt.Fprintf(os.Stderr, "error: unknown unit '%s'\n", unitID)
		os.Exit(1)
	}

	act, ok := target.Actions[actionName]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unit '%s' has no action '%s'\n", unitID, actionName)
		os.Exit(1)
	}

	task := Task{
		Name: act.Task,
		Path: filepath.Join(target.Dir, "tasks", act.Task),
	}
	executor := &DefaultExecutor{}
	result := executor.Execute(task, ExecutionPlan{Shell: "bash"}, context.Background())

	if result.Status == StatusError {
		pterm.Error.Printfln("Action '%s' on unit '%s' failed: %s", actionName, unitID, result.Message)
		os.Exit(1)
	}
	pterm.Success.Printfln("Action '%s' on unit '%s' completed", actionName, unitID)
}

func saveRunLog(results []TaskResult, logPath string, manifest *Manifest) {
	data := &RunData{
		Host:       manifest.Host,
		User:       manifest.User,
		StartedAt:  manifest.StartedAt,
		FinishedAt: manifest.FinishedAt,
		Status:     manifest.Status,
	}

	// Convert TaskResults to StepResults for backward-compatible log format.
	for _, r := range results {
		data.Steps = append(data.Steps, StepResult{
			ID:        r.ID,
			Name:      r.UnitID,
			Status:    r.Status,
			Message:   r.Message,
			StartedAt: r.StartedAt,
			FinishedAt: r.StartedAt, // best approximation
			DurationS: r.Duration,
			ExitCode:  r.ExitCode,
		})
	}

	_ = writeJSON(logPath, data)
}

func printNewSummary(results []TaskResult, logPath string) {
	success, errors, skipped, totalS := 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case StatusSuccess:
			success++
		case StatusError:
			errors++
		case StatusSkipped:
			skipped++
		}
		totalS += r.Duration
	}

	fmt.Println()
	pterm.Printfln("Result: %d success, %d error, %d skipped   Total: %s",
		success, errors, skipped, formatDuration(totalS))
	pterm.FgGray.Printfln("Log: %s", logPath)
	fmt.Println()
}
