package main

import (
	"fmt"
	"strings"
)

func render(line, dir, projectDir string, rateLimits RateLimits) {
	lines := []string{line}

	envSeg := envSegment(dir, projectDir)
	if envSeg != "" {
		lines = append(lines, envSeg)
	}

	usageSeg := usageSegment(rateLimits)
	if usageSeg != "" {
		lines = append(lines, usageSeg)
	}

	gitSeg := gitSegment(dir)
	ghSeg := ghSegment(dir)
	if gitSeg != "" || ghSeg != "" {
		parts := filterEmpty(gitSeg, ghSeg)
		lines = append(lines, strings.Join(parts, Sep))
	}

	// Render: box layout with rounded corners
	maxW := 0
	for _, l := range lines {
		if w := visibleWidth(l); w > maxW {
			maxW = w
		}
	}

	top := CComment + "╭─" + strings.Repeat("─", maxW) + "─╮" + Rst
	bot := CComment + "╰─" + strings.Repeat("─", maxW) + "─╯" + Rst

	var out []string
	out = append(out, top)
	for _, l := range lines {
		pad := strings.Repeat(" ", maxW-visibleWidth(l))
		out = append(out, CComment+"│ "+Rst+l+pad+CComment+" │"+Rst)
	}
	out = append(out, bot)
	fmt.Print(strings.Join(out, "\n"))
}

func filterEmpty(ss ...string) []string {
	var result []string
	for _, s := range ss {
		if s != "" {
			result = append(result, s)
		}
	}
	return result
}
