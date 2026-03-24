package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"dotfiles/installer/internal/luavm"
	"dotfiles/installer/internal/runner"
	"dotfiles/installer/units"

	"github.com/pterm/pterm"
)

func resolveUnitsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get working directory: %v\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(cwd, "src", "installer", "units")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: units directory not found: %s\n", dir)
		os.Exit(1)
	}
	return dir
}

func listUnits(allUnits []units.Unit) {
	data := pterm.TableData{{"#", "ID", "Name", "Critical", "Stages"}}
	for i, u := range allUnits {
		stages := make([]string, len(u.Stages))
		for j, s := range u.Stages {
			stages[j] = s.Name
		}
		crit := ""
		if u.Critical {
			crit = "yes"
		}
		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			u.ID,
			u.Name,
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

	// Load units from Lua files.
	vm := luavm.NewVM(unitsDir)
	loadedUnits, err := vm.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer luavm.CloseAll(loadedUnits)

	// Extract Unit structs for the runner.
	allUnits := make([]units.Unit, len(loadedUnits))
	for i, lu := range loadedUnits {
		allUnits[i] = lu.Unit
	}

	if *list {
		listUnits(allUnits)
		return
	}

	if *unit != "" {
		var filteredUnits []units.Unit
		var filteredLoaded []*luavm.LoadedUnit
		for i, u := range allUnits {
			if u.ID == *unit {
				filteredUnits = append(filteredUnits, u)
				filteredLoaded = append(filteredLoaded, loadedUnits[i])
			}
		}
		if len(filteredUnits) == 0 {
			fmt.Fprintf(os.Stderr, "error: unknown unit '%s'\n", *unit)
			os.Exit(1)
		}
		allUnits = filteredUnits
		loadedUnits = filteredLoaded
	}

	if *action != "" {
		parts := strings.SplitN(*action, ":", 2)
		if len(parts) != 2 {
			fmt.Fprintf(os.Stderr, "error: --action format is action_name:unit_id\n")
			os.Exit(1)
		}
		runAction(loadedUnits, parts[0], parts[1])
		return
	}

	dag, err := runner.NewDAG(allUnits)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	manifest := runner.NewManifest(dag, allUnits)
	manifestPath := filepath.Join(unitsDir, "manifest.json")

	if *dryRun {
		manifest.FinishedAt = runner.NowISO()
		manifest.Status = "dry-run"
		if err := manifest.Save(manifestPath); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		pterm.Success.Printfln("Manifest written to %s (%d waves, %d units)", manifestPath, len(dag.Waves()), len(allUnits))
		return
	}

	pterm.DefaultHeader.Println("dotfiles runner")
	fmt.Println()

	luaExec := runner.NewLuaExecutor(loadedUnits)
	r := runner.NewRunner(allUnits, dag, manifest, luaExec, runner.RunOpts{
		Force:   *force,
		Verbose: *verbose,
	})

	ctx := context.Background()
	results, aborted := r.Run(ctx)

	manifest.FinishedAt = runner.NowISO()
	manifest.Status = manifest.DeriveStatus(aborted)
	_ = manifest.Save(manifestPath)

	logsDir := filepath.Join(filepath.Dir(unitsDir), "logs")
	tsFile := time.Now().Format("2006-01-02T15-04-05")
	logPath := filepath.Join(logsDir, tsFile+".json")
	runner.SaveRunLog(results, logPath, manifest)
	runner.PruneOldLogs(logsDir, 3)

	printSummary(results, logPath)

	if aborted {
		os.Exit(1)
	}
}

func runAction(loadedUnits []*luavm.LoadedUnit, actionName, unitID string) {
	var target *luavm.LoadedUnit
	for _, lu := range loadedUnits {
		if lu.Unit.ID == unitID {
			target = lu
			break
		}
	}
	if target == nil {
		fmt.Fprintf(os.Stderr, "error: unknown unit '%s'\n", unitID)
		os.Exit(1)
	}

	act, ok := target.Unit.Actions[actionName]
	if !ok {
		fmt.Fprintf(os.Stderr, "error: unit '%s' has no action '%s'\n", unitID, actionName)
		os.Exit(1)
	}

	// The action references a task by name — execute it via LuaExecutor.
	task := units.Task{
		Name:   act.Task,
		UnitID: unitID,
	}
	luaExec := runner.NewLuaExecutor([]*luavm.LoadedUnit{target})
	result := luaExec.Execute(task, units.ExecutionPlan{}, context.Background())

	if result.Status == units.StatusError {
		pterm.Error.Printfln("Action '%s' on unit '%s' failed: %s", actionName, unitID, result.Message)
		os.Exit(1)
	}
	pterm.Success.Printfln("Action '%s' on unit '%s' completed", actionName, unitID)
}

func printSummary(results []units.TaskResult, logPath string) {
	success, errors, skipped, totalS := 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case units.StatusSuccess:
			success++
		case units.StatusError:
			errors++
		case units.StatusSkipped:
			skipped++
		}
		totalS += r.Duration
	}

	fmt.Println()
	pterm.Printfln("Result: %d success, %d error, %d skipped   Total: %s",
		success, errors, skipped, runner.FormatDuration(totalS))
	pterm.FgGray.Printfln("Log: %s", logPath)
	fmt.Println()
}
