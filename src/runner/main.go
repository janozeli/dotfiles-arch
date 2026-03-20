package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pterm/pterm"
)

func resolveStepsDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: cannot get working directory: %v\n", err)
		os.Exit(1)
	}
	dir := filepath.Join(cwd, "src", "steps")
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: steps directory not found: %s\n", dir)
		os.Exit(1)
	}
	return dir
}

func listSteps() {
	data := pterm.TableData{{"#", "ID", "Name"}}
	for i, step := range Steps() {
		data = append(data, []string{fmt.Sprintf("%d", i+1), step.ID(), step.Name()})
	}
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()
}

func main() {
	force := flag.Bool("force", false, "skip pre-verify checks")
	list := flag.Bool("list", false, "list available steps and exit")
	only := flag.String("step", "", "run only these steps (comma-separated IDs)")
	flag.Parse()

	dir := resolveStepsDir()
	initBashSteps(dir)

	if *list {
		listSteps()
		return
	}

	// Build filter set from --step
	filter := map[string]bool{}
	if *only != "" {
		for _, id := range strings.Split(*only, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				filter[id] = true
			}
		}
		// Validate IDs
		allIDs := map[string]bool{}
		for _, s := range Steps() {
			allIDs[s.ID()] = true
		}
		for id := range filter {
			if !allIDs[id] {
				fmt.Fprintf(os.Stderr, "error: unknown step '%s'\n", id)
				os.Exit(1)
			}
		}
	}

	startedAt := nowISO()
	tsFile := time.Now().Format("2006-01-02T15-04-05")
	logsDir := filepath.Join(filepath.Dir(dir), "logs")
	logPath := filepath.Join(logsDir, tsFile+".json")

	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	runData := &RunData{
		Host:      hostname,
		User:      user,
		StartedAt: startedAt,
	}

	env := os.Environ()
	completed := map[string]bool{}
	var results []StepResult
	aborted := false

	for _, step := range Steps() {
		if len(filter) > 0 && !filter[step.ID()] {
			continue
		}

		printStepHeader(step.ID(), step.Name())

		sp, _ := pterm.DefaultSpinner.Start(step.Name())
		result := evaluate(step, *force, completed, env)
		if result.Status == StatusSuccess || result.Status == StatusWarning {
			sp.Success(result.Message)
		} else if result.Status == StatusError {
			sp.Fail(result.Message)
		} else {
			sp.Info(result.Message)
		}

		if result.Status.Completed() {
			completed[step.ID()] = true
		}

		printStepResult(result)

		results = append(results, result)
		saveRun(runData, results, logPath)

		if result.Status == StatusError && step.Critical() {
			detail := result.Message
			if result.ExitCode != nil {
				detail = fmt.Sprintf("exit %d", *result.ExitCode)
			}
			printAbort(step.Name(), detail)
			aborted = true
			break
		}
	}

	runData.Status = deriveRunStatus(results, aborted)
	runData.FinishedAt = nowISO()
	if err := writeJSON(logPath, runData); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to write log: %v\n", err)
	}
	pruneOldLogs(filepath.Dir(logPath), 3)
	printSummary(results, logPath)

	if aborted {
		os.Exit(1)
	}
}
