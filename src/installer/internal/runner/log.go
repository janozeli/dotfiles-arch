package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"dotfiles/installer/units"
)

// stepResult is the legacy JSON log format. Unexported — only used by SaveRunLog.
type stepResult struct {
	ID         string       `json:"id"`
	Name       string       `json:"name"`
	Status     units.Status `json:"status"`
	Message    string       `json:"message"`
	StartedAt  string       `json:"started_at"`
	FinishedAt string       `json:"finished_at"`
	DurationS  int          `json:"duration_s"`
	ExitCode   *int         `json:"exit_code"`
}

// runData is the top-level JSON log structure.
type runData struct {
	Host       string       `json:"host"`
	User       string       `json:"user"`
	StartedAt  string       `json:"started_at"`
	FinishedAt string       `json:"finished_at"`
	Status     string       `json:"status"`
	Steps      []stepResult `json:"steps"`
}

// NowISO returns the current time in ISO 8601 format.
func NowISO() string {
	return time.Now().Format("2006-01-02T15:04:05")
}

// FormatDuration formats seconds as "Xm YYs" or "Xs".
func FormatDuration(seconds int) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	m := seconds / 60
	s := seconds % 60
	return fmt.Sprintf("%dm%02ds", m, s)
}

// SaveRunLog writes the run results to a JSON log file.
func SaveRunLog(results []units.TaskResult, logPath string, manifest *Manifest) {
	data := &runData{
		Host:       manifest.Host,
		User:       manifest.User,
		StartedAt:  manifest.StartedAt,
		FinishedAt: manifest.FinishedAt,
		Status:     manifest.Status,
	}

	for _, r := range results {
		data.Steps = append(data.Steps, stepResult{
			ID:         r.ID,
			Name:       r.UnitID,
			Status:     r.Status,
			Message:    r.Message,
			StartedAt:  r.StartedAt,
			FinishedAt: r.StartedAt,
			DurationS:  r.Duration,
			ExitCode:   r.ExitCode,
		})
	}

	dir := filepath.Dir(logPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(logPath, b, 0o644)
}

// PruneOldLogs keeps only the last N log files.
func PruneOldLogs(logsDir string, keep int) {
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
