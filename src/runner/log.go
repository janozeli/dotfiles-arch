package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// RunData is the top-level JSON log structure.
type RunData struct {
	Host       string       `json:"host"`
	User       string       `json:"user"`
	StartedAt  string       `json:"started_at"`
	FinishedAt string       `json:"finished_at"`
	Status     string       `json:"status"`
	Steps      []StepResult `json:"steps"`
}

func nowISO() string {
	return time.Now().Format("2006-01-02T15:04:05")
}

func formatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%dm%02ds", m, s)
}

func writeJSON(logPath string, data *RunData) error {
	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(logPath, b, 0o644)
}

func saveRun(runData *RunData, results []StepResult, logPath string) {
	runData.Steps = results
	runData.FinishedAt = nowISO()
	_ = writeJSON(logPath, runData)
}

func deriveRunStatus(results []StepResult, aborted bool) string {
	if aborted {
		return "error"
	}
	for _, r := range results {
		if r.Status == StatusError {
			return "partial"
		}
	}
	for _, r := range results {
		if r.Status == StatusWarning {
			return "warning"
		}
	}
	return "success"
}

func printSummary(results []StepResult, logPath string) {
	success, errors, skipped, warnings, totalS := 0, 0, 0, 0, 0
	for _, r := range results {
		switch r.Status {
		case StatusSuccess:
			success++
		case StatusError:
			errors++
		case StatusSkipped:
			skipped++
		case StatusWarning:
			warnings++
		}
		totalS += r.DurationS
	}

	statusIcons := map[Status]string{
		StatusSuccess: "✓ success",
		StatusWarning: "⚠ warning",
		StatusError:   "✗ error",
		StatusSkipped: "○ skipped",
	}

	fmt.Println()
	fmt.Printf(" %2s  %-28s %-14s %8s\n", "#", "Step", "Status", "Duration")
	fmt.Println("─────────────────────────────────────────────────────────")
	for i, r := range results {
		icon := statusIcons[r.Status]
		dur := formatDuration(r.DurationS)
		msg := ""
		if r.Status != StatusError && r.ExitCode == nil {
			msg = fmt.Sprintf("  (%s)", r.Message)
		}
		fmt.Printf(" %2d  %-28s %-14s %8s%s\n", i+1, r.Name, icon, dur, msg)
	}
	fmt.Println("─────────────────────────────────────────────────────────")
	fmt.Printf(" Result: %d success, %d warning, %d error, %d skipped   Total: %s\n", success, warnings, errors, skipped, formatDuration(totalS))
	fmt.Println()
	fmt.Printf(" Log: %s\n", logPath)
	fmt.Println()
}
