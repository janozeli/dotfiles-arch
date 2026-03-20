package main

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

func runCmd(timeout time.Duration, cwd, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	if cwd != "" {
		cmd.Dir = cwd
	}
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func gitCmd(cwd string, args ...string) string {
	out, err := runCmd(2*time.Second, cwd, "git", args...)
	if err != nil {
		return ""
	}
	return out
}
