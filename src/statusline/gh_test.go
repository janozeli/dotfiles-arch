package main

import "testing"

func TestParseGhAuthStatus(t *testing.T) {
	tests := []struct {
		name, output, want string
	}{
		{"empty", "", ""},
		{"active account",
			"  Logged in to github.com account janozeli (keyring)\n  Active account: true\n",
			"janozeli"},
		{"inactive account",
			"  Logged in to github.com account janozeli (keyring)\n  Active account: false\n",
			""},
		{"no match", "some other output\n", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseGhAuthStatus(tt.output); got != tt.want {
				t.Errorf("parseGhAuthStatus(%q) = %q, want %q", tt.output, got, tt.want)
			}
		})
	}
}
