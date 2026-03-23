package main

import (
	"fmt"
	"strings"
)

func usageSegment(limits RateLimits) string {
	var parts []string
	usageURL := "https://claude.ai/settings/usage"

	if limits.FiveHour.UsedPercentage > 0 {
		pct := limits.FiveHour.UsedPercentage
		color := thresholdColor(pct)
		reset := formatResetCompact(limits.FiveHour.ResetsAt)
		parts = append(parts, osc8Link(usageURL, CPurple+"\uf1da Session:"+Rst+" "+color+fmt.Sprintf("%.0f%%%s", pct, reset))+Rst)
	}

	if limits.SevenDay.UsedPercentage > 0 {
		pct := limits.SevenDay.UsedPercentage
		color := thresholdColor(pct)
		reset := formatResetCompact(limits.SevenDay.ResetsAt)
		parts = append(parts, osc8Link(usageURL, CPurple+"\uf073 Weekly:"+Rst+" "+color+fmt.Sprintf("%.0f%%%s", pct, reset))+Rst)
	}

	return strings.Join(parts, Sep)
}
