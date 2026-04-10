package main

import (
	"testing"
)

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"plain text", "hello", "hello"},
		{"color code", "\x1b[38;2;189;147;249mhello\x1b[0m", "hello"},
		{"osc8 link", "\x1b]8;;https://example.com\x1b\\text\x1b]8;;\x1b\\", "text"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripAnsi(tt.input); got != tt.want {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii", "hello", 5},
		{"with ansi", "\x1b[31mhello\x1b[0m", 5},
		{"emoji", "\U0001F480", 2},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := visibleWidth(tt.input); got != tt.want {
				t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestOsc8Link(t *testing.T) {
	got := osc8Link("https://example.com", "click")
	want := "\x1b]8;;https://example.com\x1b\\click\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("osc8Link() = %q, want %q", got, want)
	}
}

func TestThresholdColor(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"low", 30, CGreen},
		{"boundary 59", 59, CGreen},
		{"boundary 60", 60, CYellow},
		{"mid", 75, CYellow},
		{"boundary 85", 85, CRed},
		{"high", 99, CRed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := thresholdColor(tt.pct); got != tt.want {
				t.Errorf("thresholdColor(%.0f) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestEditorFileURL(t *testing.T) {
	tests := []struct {
		name, editor, path, want string
	}{
		{"zed", "zed", "/home/user/f.go", "zed://file/home/user/f.go"},
		{"zeditor", "zeditor", "/home/user/f.go", "zed://file/home/user/f.go"},
		{"code", "code", "/home/user/f.go", "vscode://file/home/user/f.go"},
		{"cursor", "cursor", "/home/user/f.go", "cursor://file/home/user/f.go"},
		{"vim fallback", "vim", "/home/user/f.go", "file:///home/user/f.go"},
		{"empty fallback", "", "/home/user/f.go", "file:///home/user/f.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EDITOR", tt.editor)
			if got := editorFileURL(tt.path); got != tt.want {
				t.Errorf("editorFileURL(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
