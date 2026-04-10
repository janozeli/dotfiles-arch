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
