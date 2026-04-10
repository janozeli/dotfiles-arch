package main

import (
	"strings"
	"testing"
)

func TestEnvSegment(t *testing.T) {
	t.Setenv("EDITOR", "vim")

	tests := []struct {
		name         string
		home         string
		cwd          string
		projectDir   string
		wantEmpty    bool
		wantContains []string
	}{
		{"empty cwd", "/home/user", "", "", true, nil},
		{"at home", "/home/user", "/home/user", "", false, []string{"~/"}},
		{"subdir", "/home/user", "/home/user/projects/foo", "", false, []string{"~/projects/foo"}},
		{"deep path", "/home/user", "/home/user/a/b/c/d", "", false, []string{"~/a/.../d"}},
		{"with project dir", "/home/user", "/home/user/proj/sub", "/home/user/proj", false, []string{"~/proj", "~/proj/sub"}},
		{"outside home", "/home/user", "/opt/project", "", false, []string{"/opt/project"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("HOME", tt.home)
			got := envSegment(tt.cwd, tt.projectDir)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("envSegment(%q, %q) = %q, want empty", tt.cwd, tt.projectDir, got)
				}
				return
			}
			visible := stripAnsi(got)
			for _, want := range tt.wantContains {
				if !strings.Contains(visible, want) {
					t.Errorf("envSegment(%q, %q) visible=%q, want contains %q", tt.cwd, tt.projectDir, visible, want)
				}
			}
		})
	}
}
