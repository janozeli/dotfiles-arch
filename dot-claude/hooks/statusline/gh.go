package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func getGhAccount() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	out, _ := cmd.CombinedOutput()

	var current string
	for _, line := range strings.Split(string(out), "\n") {
		if idx := strings.Index(line, "Logged in to github.com account "); idx >= 0 {
			rest := line[idx+len("Logged in to github.com account "):]
			if fields := strings.Fields(rest); len(fields) > 0 {
				current = fields[0]
			}
		}
		if strings.Contains(line, "Active account: true") && current != "" {
			return current
		}
	}
	return ""
}

func ghSegment(cwd string) string {
	account := getGhAccount()
	if account == "" {
		return ""
	}

	seg := CPink + "\uf09b " + account + Rst

	gitUser := gitCmd(cwd, "config", "user.name")
	if gitUser != "" && !strings.EqualFold(gitUser, account) {
		seg += Sep + CCyan + "\ue702 " + gitUser + Rst
	}

	return seg
}
