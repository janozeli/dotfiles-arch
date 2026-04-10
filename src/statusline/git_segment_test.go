package main

import "testing"

func TestGitPathDisplay(t *testing.T) {
	tests := []struct {
		name, root, cwd, want string
	}{
		{"same dir", "/home/user/repo", "/home/user/repo", ""},
		{"one level", "/home/user/repo", "/home/user/repo/src", "repo/src"},
		{"two levels", "/home/user/repo", "/home/user/repo/src/pkg", "repo/src/pkg"},
		{"deep path", "/home/user/repo", "/home/user/repo/a/b/c", "repo/.../c"},
		{"very deep", "/home/user/repo", "/home/user/repo/a/b/c/d/e", "repo/.../e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := gitPathDisplay(tt.root, tt.cwd); got != tt.want {
				t.Errorf("gitPathDisplay(%q, %q) = %q, want %q", tt.root, tt.cwd, got, tt.want)
			}
		})
	}
}

func TestRemoteToRepoURL(t *testing.T) {
	tests := []struct {
		name, remote, want string
	}{
		{"ssh with .git", "git@github.com:user/repo.git", "https://github.com/user/repo"},
		{"ssh without .git", "git@github.com:user/repo", "https://github.com/user/repo"},
		{"https with .git", "https://github.com/user/repo.git", "https://github.com/user/repo"},
		{"https without .git", "https://github.com/user/repo", "https://github.com/user/repo"},
		{"non-github", "git@gitlab.com:user/repo.git", ""},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := remoteToRepoURL(tt.remote); got != tt.want {
				t.Errorf("remoteToRepoURL(%q) = %q, want %q", tt.remote, got, tt.want)
			}
		})
	}
}

func TestParseNumstat(t *testing.T) {
	tests := []struct {
		name             string
		numstat          string
		wantIns, wantDel int
	}{
		{"empty", "", 0, 0},
		{"single file", "10\t5\tfile.go", 10, 5},
		{"multiple files", "10\t5\ta.go\n3\t2\tb.go", 13, 7},
		{"binary file", "-\t-\timage.png", 0, 0},
		{"mixed", "10\t5\ta.go\n-\t-\tb.png\n3\t0\tc.go", 13, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ins, del := parseNumstat(tt.numstat)
			if ins != tt.wantIns || del != tt.wantDel {
				t.Errorf("parseNumstat(%q) = (%d, %d), want (%d, %d)", tt.numstat, ins, del, tt.wantIns, tt.wantDel)
			}
		})
	}
}

func TestParseAheadBehind(t *testing.T) {
	tests := []struct {
		name                   string
		input                  string
		wantAhead, wantBehind int
	}{
		{"empty", "", 0, 0},
		{"ahead only", "3\t0", 3, 0},
		{"behind only", "0\t5", 0, 5},
		{"both", "2\t4", 2, 4},
		{"malformed", "abc", 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ahead, behind := parseAheadBehind(tt.input)
			if ahead != tt.wantAhead || behind != tt.wantBehind {
				t.Errorf("parseAheadBehind(%q) = (%d, %d), want (%d, %d)", tt.input, ahead, behind, tt.wantAhead, tt.wantBehind)
			}
		})
	}
}
