package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
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
		ProjectDir string `json:"project_dir"`
	} `json:"workspace"`
	ContextWindow struct {
		UsedPercentage *int `json:"used_percentage"`
		TotalTokens    int  `json:"context_window_size"`
	} `json:"context_window"`
	RateLimits RateLimits `json:"rate_limits"`
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

	debugDump("statusline-raw-input.jsonc", raw)

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
	// Context window display
	ctx := ""
	if data.ContextWindow.UsedPercentage != nil {
		ctx = contextDisplay(*data.ContextWindow.UsedPercentage, data.ContextWindow.TotalTokens)
	}

	// Build status line
	ctxSep := " " + CComment + "\u2502" + Rst
	line := CPurple + "\uee9c " + model + Rst
	if ctx != "" {
		line += ctxSep + ctx
	}

	projectDir := data.Workspace.ProjectDir
	render(line, dir, projectDir, data.RateLimits)
}

func contextDisplay(usedPct, totalTokens int) string {
	usedTokens := totalTokens * usedPct / 100
	var tokenStr string
	if usedTokens >= 1000 {
		tokenStr = fmt.Sprintf("%dk", (usedTokens+500)/1000)
	} else {
		tokenStr = fmt.Sprintf("%d", usedTokens)
	}

	filled := usedPct / 10
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", 10-filled)

	if usedPct < 50 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CGreen, tokenStr, bar, usedPct, Rst)
	} else if usedPct < 65 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CYellow, tokenStr, bar, usedPct, Rst)
	} else if usedPct < 80 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CRed, tokenStr, bar, usedPct, Rst)
	}
	return fmt.Sprintf(" \x1b[5;38;2;255;85;85m\U0001F480 %s %s %d%%%s", tokenStr, bar, usedPct, Rst)
}
