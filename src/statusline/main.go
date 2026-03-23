package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type RateLimits struct {
	FiveHour struct {
		UsedPercentage float64 `json:"used_percentage"`
		ResetsAt       int64   `json:"resets_at"`
	} `json:"five_hour"`
	SevenDay struct {
		UsedPercentage float64 `json:"used_percentage"`
		ResetsAt       int64   `json:"resets_at"`
	} `json:"seven_day"`
}

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
	RateLimits RateLimits `json:"rate_limits"`
}

func subcommand() string {
	for _, arg := range os.Args[1:] {
		if arg == "--debug" {
			continue
		}
		return arg
	}
	return ""
}

func main() {
	if sub := subcommand(); sub != "" {
		gsdDispatch(sub)
		return
	}

	// Timeout guard: if stdin doesn't close within 3s, exit silently
	timer := time.AfterFunc(3*time.Second, func() {
		os.Exit(0)
	})

	raw, err := io.ReadAll(os.Stdin)
	if err != nil {
		return
	}
	timer.Stop()

	debugDump("statusline-raw-input.json", raw)

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

	// GSD update available? (only if current project uses GSD)
	gsdUpdate := ""
	if _, err := os.Stat(filepath.Join(dir, ".planning")); err == nil {
		gsdCacheFile := filepath.Join(claudeDir, "cache", "gsd-update-check.json")
		gsdUpdate = checkGsdUpdate(gsdCacheFile)
	}

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

	render(gsdLine, dir, data.RateLimits)
}

