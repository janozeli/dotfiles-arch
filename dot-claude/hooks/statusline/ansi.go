package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

var reAnsi = regexp.MustCompile(`\x1b\[[0-9;]*m|\x1b\]8;;[^\x1b]*\x1b\\`)

func stripAnsi(s string) string {
	return reAnsi.ReplaceAllString(s, "")
}

func visibleWidth(s string) int {
	clean := stripAnsi(s)
	w := 0
	for _, r := range clean {
		if r >= 0x1F300 {
			w += 2
		} else {
			w++
		}
	}
	return w
}

const (
	CPurple  = "\x1b[38;2;189;147;249m" // #BD93F9
	CPink    = "\x1b[38;2;255;121;198m" // #FF79C6
	CCyan    = "\x1b[38;2;139;233;253m" // #8BE9FD
	CGreen   = "\x1b[38;2;80;250;123m"  // #50FA7B
	CRed     = "\x1b[38;2;255;85;85m"   // #FF5555
	CYellow  = "\x1b[38;2;241;250;140m" // #F1FA8C
	CComment = "\x1b[38;2;98;114;164m"  // #6272A4
	Rst      = "\x1b[0m"
	Sep      = " \x1b[38;2;98;114;164m\u2502\x1b[0m "
)

func osc8Link(url, text string) string {
	return fmt.Sprintf("\x1b]8;;%s\x1b\\%s\x1b]8;;\x1b\\", url, text)
}

func editorFileURL(filePath string) string {
	editor := filepath.Base(os.Getenv("EDITOR"))
	schemes := map[string]string{
		"zed":     "zed://file",
		"zeditor": "zed://file",
		"code":    "vscode://file",
		"cursor":  "cursor://file",
	}
	scheme, ok := schemes[editor]
	if !ok {
		scheme = "file://"
	}
	return scheme + filePath
}

func thresholdColor(pct float64) string {
	if pct >= 85 {
		return CRed
	}
	if pct >= 60 {
		return CYellow
	}
	return CGreen
}

func formatResetCompact(resetsAt string) string {
	if resetsAt == "" {
		return ""
	}
	t, err := time.Parse(time.RFC3339, resetsAt)
	if err != nil {
		return ""
	}
	diff := time.Until(t)
	if diff <= 0 {
		return ""
	}
	totalHours := int(diff.Hours())
	minutes := int(diff.Minutes()) % 60
	days := totalHours / 24
	hours := totalHours % 24
	if days > 0 {
		return fmt.Sprintf(" \uf1da %dd%02dh", days, hours)
	}
	return fmt.Sprintf(" \uf1da %dh%02d", totalHours, minutes)
}
