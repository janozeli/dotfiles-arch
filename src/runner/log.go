package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/pterm/pterm"
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

func pruneOldLogs(logsDir string, keep int) {
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		return
	}
	var logs []string
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
			logs = append(logs, e.Name())
		}
	}
	sort.Strings(logs)
	if len(logs) <= keep {
		return
	}
	for _, name := range logs[:len(logs)-keep] {
		_ = os.Remove(filepath.Join(logsDir, name))
	}
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

	data := pterm.TableData{{"#", "Step", "Status", "Duration"}}
	for i, r := range results {
		data = append(data, []string{
			fmt.Sprintf("%d", i+1),
			r.Name,
			statusIcons[r.Status],
			formatDuration(r.DurationS),
		})
	}

	fmt.Println()
	pterm.DefaultTable.WithHasHeader().WithData(data).Render()

	pterm.Printfln("Result: %d success, %d warning, %d error, %d skipped   Total: %s",
		success, warnings, errors, skipped, formatDuration(totalS))
	pterm.FgGray.Printfln("Log: %s", logPath)
	fmt.Println()
}
