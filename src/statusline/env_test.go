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
		// cwd vazio
		{"empty cwd", "/home/user", "", "", true, nil},

		// cwd sob HOME, sem abreviacao (parts <= 3)
		{"at home", "/home/user", "/home/user", "", false, []string{"~/"}},
		{"one level", "/home/user", "/home/user/foo", "", false, []string{"~/foo"}},
		{"two levels", "/home/user", "/home/user/foo/bar", "", false, []string{"~/foo/bar"}},

		// cwd sob HOME, com abreviacao (parts > 3)
		{"three levels abbreviated", "/home/user", "/home/user/a/b/c", "", false, []string{"~/a/.../c"}},
		{"four levels abbreviated", "/home/user", "/home/user/a/b/c/d", "", false, []string{"~/a/.../d"}},

		// cwd fora de HOME
		{"outside home shallow", "/home/user", "/opt/project", "", false, []string{"/opt/project"}},
		{"outside home deep", "/home/user", "/opt/a/b/c", "", false, []string{"/opt/.../c"}},

		// HOME vazio
		{"no home shallow", "", "/opt/proj", "", false, []string{"/opt/proj"}},
		{"no home deep", "", "/home/user/a/b/c", "", false, []string{"/home/.../c"}},

		// projectDir sob HOME
		{"project under home", "/home/user", "/home/user/proj/sub", "/home/user/proj", false, []string{"~/proj", "~/proj/sub"}},
		{"project under home nested", "/home/user", "/home/user/work/proj/sub", "/home/user/work/proj", false, []string{"~/work/proj", "~/work/.../sub"}},

		// projectDir fora de HOME
		{"project outside home", "/home/user", "/opt/proj/sub", "/opt/proj", false, []string{"proj", "/opt/.../sub"}},

		// projectDir com HOME vazio
		{"project no home", "", "/opt/proj/sub", "/opt/proj", false, []string{"proj", "/opt/.../sub"}},
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
