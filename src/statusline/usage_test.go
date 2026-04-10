package main

import (
	"strings"
	"testing"
)

func TestUsageSegment(t *testing.T) {
	tests := []struct {
		name         string
		limits       RateLimits
		wantEmpty    bool
		wantContains []string
	}{
		{"empty", RateLimits{}, true, nil},
		{"five hour only", func() RateLimits {
			var l RateLimits
			l.FiveHour.UsedPercentage = 50
			return l
		}(), false, []string{"Session:", "50%"}},
		{"seven day only", func() RateLimits {
			var l RateLimits
			l.SevenDay.UsedPercentage = 30
			return l
		}(), false, []string{"Weekly:", "30%"}},
		{"both", func() RateLimits {
			var l RateLimits
			l.FiveHour.UsedPercentage = 50
			l.SevenDay.UsedPercentage = 30
			return l
		}(), false, []string{"Session:", "50%", "Weekly:", "30%"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usageSegment(tt.limits)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("usageSegment() = %q, want empty", got)
				}
				return
			}
			visible := stripAnsi(got)
			for _, want := range tt.wantContains {
				if !strings.Contains(visible, want) {
					t.Errorf("usageSegment() visible=%q, want contains %q", visible, want)
				}
			}
		})
	}
}
