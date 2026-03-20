package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Input struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	SessionID     string `json:"session_id"`
	ContextWindow struct {
		RemainingPercentage *float64 `json:"remaining_percentage"`
		TotalTokens         int      `json:"total_tokens"`
	} `json:"context_window"`
}

func main() {
	// Timeout guard: if stdin doesn't close within 3s, exit silently
	timer := time.AfterFunc(3*time.Second, func() {
		os.Exit(0)
	})

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	timer.Stop()

	var data Input
	if err := json.Unmarshal(raw, &data); err != nil {
		return
	}

	model := data.Model.DisplayName
	if model == "" {
		model = "Claude"
	}
	dir := data.Workspace.CurrentDir
	if dir == "" {
		dir, _ = os.Getwd()
	}
	session := data.SessionID

	// Context window display (shows USED percentage scaled to usable context)
	const autoCompactBufferPct = 16.5
	ctx := ""
	if data.ContextWindow.RemainingPercentage != nil {
		remaining := *data.ContextWindow.RemainingPercentage
		usableRemaining := math.Max(0, (remaining-autoCompactBufferPct)/(100-autoCompactBufferPct)*100)
		used := int(math.Max(0, math.Min(100, math.Round(100-usableRemaining))))

		// Bridge file for context-monitor
		if session != "" {
			bridgePath := filepath.Join(os.TempDir(), fmt.Sprintf("claude-ctx-%s.json", session))
			bridgeData, _ := json.Marshal(map[string]interface{}{
				"session_id":           session,
				"remaining_percentage": remaining,
				"used_pct":            used,
				"timestamp":           time.Now().Unix(),
			})
			os.WriteFile(bridgePath, bridgeData, 0644)
		}

		totalTokens := data.ContextWindow.TotalTokens
		if totalTokens == 0 {
			totalTokens = 200000
		}
		usedTokens := int(math.Round(float64(totalTokens) * float64(used) / 100))
		var tokenStr string
		if usedTokens >= 1000 {
			tokenStr = fmt.Sprintf("%dk", int(math.Round(float64(usedTokens)/1000)))
		} else {
			tokenStr = fmt.Sprintf("%d", usedTokens)
		}

		filled := used / 10
		bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", 10-filled)

		if used < 50 {
			ctx = fmt.Sprintf(" %s%s %s %d%%%s", CGreen, tokenStr, bar, used, Rst)
		} else if used < 65 {
			ctx = fmt.Sprintf(" %s%s %s %d%%%s", CYellow, tokenStr, bar, used, Rst)
		} else if used < 80 {
			ctx = fmt.Sprintf(" %s%s %s %d%%%s", CRed, tokenStr, bar, used, Rst)
		} else {
			ctx = fmt.Sprintf(" \x1b[5;38;2;255;85;85m\U0001F480 %s %s %d%%%s", tokenStr, bar, used, Rst)
		}
	}

	// Current task from todos
	task := ""
	homeDir, _ := os.UserHomeDir()
	claudeDir := os.Getenv("CLAUDE_CONFIG_DIR")
	if claudeDir == "" {
		claudeDir = filepath.Join(homeDir, ".claude")
	}
	todosDir := filepath.Join(claudeDir, "todos")
	if session != "" {
		task = findCurrentTask(todosDir, session)
	}

	// GSD update available?
	gsdUpdate := ""
	gsdCacheFile := filepath.Join(claudeDir, "cache", "gsd-update-check.json")
	gsdUpdate = checkGsdUpdate(gsdCacheFile)

	// Build GSD line
	ctxSep := " " + CComment + "\u2502" + Rst
	var gsdLine string
	if task != "" {
		gsdLine = gsdUpdate + CPurple + "\uee9c " + model + Rst + Sep + "\x1b[1m" + task + Rst
		if ctx != "" {
			gsdLine += ctxSep + ctx
		}
	} else {
		gsdLine = gsdUpdate + CPurple + "\uee9c " + model + Rst
		if ctx != "" {
			gsdLine += ctxSep + ctx
		}
	}

	render(gsdLine, dir)
}

func findCurrentTask(todosDir, session string) string {
	entries, err := os.ReadDir(todosDir)
	if err != nil {
		return ""
	}

	type fileEntry struct {
		name  string
		mtime time.Time
	}
	var matches []fileEntry
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, session) && strings.Contains(name, "-agent-") && strings.HasSuffix(name, ".json") {
			info, err := e.Info()
			if err != nil {
				continue
			}
			matches = append(matches, fileEntry{name: name, mtime: info.ModTime()})
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].mtime.After(matches[j].mtime)
	})

	if len(matches) == 0 {
		return ""
	}

	data, err := os.ReadFile(filepath.Join(todosDir, matches[0].name))
	if err != nil {
		return ""
	}

	var todos []struct {
		Status     string `json:"status"`
		ActiveForm string `json:"activeForm"`
	}
	if err := json.Unmarshal(data, &todos); err != nil {
		return ""
	}

	for _, t := range todos {
		if t.Status == "in_progress" {
			return t.ActiveForm
		}
	}
	return ""
}

func checkGsdUpdate(cacheFile string) string {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return ""
	}

	var cache struct {
		UpdateAvailable bool     `json:"update_available"`
		StaleHooks      []string `json:"stale_hooks"`
	}
	if err := json.Unmarshal(data, &cache); err != nil {
		return ""
	}

	result := ""
	if cache.UpdateAvailable {
		result += "\x1b[33m\u2b06 /gsd:update\x1b[0m \u2502 "
	}
	if len(cache.StaleHooks) > 0 {
		result += "\x1b[31m\u26a0 stale hooks \u2014 run /gsd:update\x1b[0m \u2502 "
	}
	return result
}
