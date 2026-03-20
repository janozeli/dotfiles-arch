package main

import (
	"fmt"
	"strings"
)

func render(gsdLine, dir string) {
	lines := []string{gsdLine}

	root := gitCmd(dir, "rev-parse", "--show-toplevel")

	envSeg := envSegment(dir, root)
	ghSeg := ghSegment(dir)
	if envSeg != "" || ghSeg != "" {
		parts := filterEmpty(envSeg, ghSeg)
		lines = append(lines, strings.Join(parts, Sep))
	}

	cache, needFetch := loadUsageCache()
	usageSeg := usageSegment(cache)
	if usageSeg != "" {
		lines = append(lines, usageSeg)
	}

	gitSeg := ""
	if root != "" {
		gitSeg = gitSegment(dir)
	}
	if gitSeg != "" {
		lines = append(lines, gitSeg)
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

	// Deferred: refresh usage cache if stale (runs after output)
	if needFetch {
		saveUsage()
	}
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
