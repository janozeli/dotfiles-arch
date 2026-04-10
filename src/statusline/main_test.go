package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestContextDisplay(t *testing.T) {
	tests := []struct {
		name         string
		usedPct      int
		totalTokens  int
		wantColor    string
		wantTokenStr string
	}{
		{"zero", 0, 200000, CGreen, "0"},
		{"low", 10, 200000, CGreen, "20k"},
		{"boundary 49", 49, 200000, CGreen, "98k"},
		{"boundary 50", 50, 200000, CYellow, "100k"},
		{"boundary 64", 64, 200000, CYellow, "128k"},
		{"boundary 65", 65, 200000, CRed, "130k"},
		{"boundary 79", 79, 200000, CRed, "158k"},
		{"boundary 80", 80, 200000, "\x1b[5;38;2;255;85;85m", "160k"},
		{"critical", 90, 200000, "\x1b[5;38;2;255;85;85m", "180k"},
		{"small tokens", 50, 1000, CYellow, "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contextDisplay(tt.usedPct, tt.totalTokens)
			if !strings.Contains(got, tt.wantColor) {
				t.Errorf("contextDisplay(%d, %d) missing color %q", tt.usedPct, tt.totalTokens, tt.wantColor)
			}
			visible := stripAnsi(got)
			if !strings.Contains(visible, tt.wantTokenStr) {
				t.Errorf("contextDisplay(%d, %d) visible=%q, want token %q", tt.usedPct, tt.totalTokens, visible, tt.wantTokenStr)
			}
			wantPct := fmt.Sprintf("%d%%", tt.usedPct)
			if !strings.Contains(visible, wantPct) {
				t.Errorf("contextDisplay(%d, %d) visible=%q, want pct %q", tt.usedPct, tt.totalTokens, visible, wantPct)
			}
		})
	}
}
