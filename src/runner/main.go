package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
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

func hasSystemd() bool {
	info, err := os.Stat("/run/systemd/system")
	return err == nil && info.IsDir()
}

func main() {
	force := flag.Bool("force", false, "skip pre-verify checks")
	flag.Parse()

	dir := resolveStepsDir()
	initBashSteps(dir)

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
	systemdOK := hasSystemd()
	completed := map[string]bool{}
	var results []StepResult
	aborted := false

	for _, step := range Steps() {
		fmt.Printf("\n\033[1;33m>>> [%s] %s\033[0m\n", step.ID(), step.Name())

		result := evaluate(step, *force, systemdOK, completed, env)

		if result.Status.Completed() {
			completed[step.ID()] = true
		}

		// Print inline status for non-executed steps
		if result.ExitCode == nil && result.Status != StatusError {
			fmt.Printf("  (%s: %s)\n", result.Status, result.Message)
		} else if result.Status == StatusError && result.ExitCode == nil {
			fmt.Printf("  \033[1;31m(error: %s)\033[0m\n", result.Message)
		}

		results = append(results, result)
		saveRun(runData, results, logPath)

		if result.Status == StatusError && step.Critical() {
			detail := result.Message
			if result.ExitCode != nil {
				detail = fmt.Sprintf("exit %d", *result.ExitCode)
			}
			fmt.Printf("\n\033[1;31m[ABORT] Critical step '%s' failed (%s)\033[0m\n", step.Name(), detail)
			aborted = true
			break
		}
	}

	runData.Status = deriveRunStatus(results, aborted)
	runData.FinishedAt = nowISO()
	_ = writeJSON(logPath, runData)
	printSummary(results, logPath)

	if aborted {
		os.Exit(1)
	}
}
