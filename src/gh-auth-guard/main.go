package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

type hookInput struct {
	ToolName  string `json:"tool_name"`
	ToolInput struct {
		Command string `json:"command"`
	} `json:"tool_input"`
}

func run(name string, args ...string) string {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

var reAccount = regexp.MustCompile(`Logged in to github\.com account (\S+)`)

func checkGhAuth() {
	gitUser := run("git", "config", "user.name")
	if gitUser == "" {
		return
	}

	ghStatus := run("gh", "auth", "status")
	if ghStatus == "" {
		return
	}

	var activeAccount string
	var currentAccount string
	for _, line := range strings.Split(ghStatus, "\n") {
		if m := reAccount.FindStringSubmatch(line); m != nil {
			currentAccount = m[1]
		}
		if strings.Contains(line, "Active account: true") && currentAccount != "" {
			activeAccount = currentAccount
		}
	}

	if activeAccount == "" {
		return
	}
	if strings.EqualFold(gitUser, activeAccount) {
		return
	}

	result := run("gh", "auth", "switch", "--user", gitUser)
	if result != "" || run("gh", "auth", "status") != "" {
		fmt.Printf("gh auth: switched %q → %q (matches git config)\n", activeAccount, gitUser)
	} else {
		fmt.Printf("⚠ gh account %q does not match git user %q.\nAuto-switch failed. Run manually: gh auth switch --user %s\n", activeAccount, gitUser, gitUser)
	}
}

var (
	reGh     = regexp.MustCompile(`\bgh\s`)
	reGhAuth = regexp.MustCompile(`\bgh\s+auth\s`)
	reGitNet = regexp.MustCompile(`\bgit\s+(push|pull|fetch|clone)\b`)
)

func main() {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(0)
	}

	var input hookInput
	if json.Unmarshal(data, &input) != nil {
		os.Exit(0)
	}

	if input.ToolName != "Bash" {
		os.Exit(0)
	}

	cmd := input.ToolInput.Command
	needsAuth := (reGh.MatchString(cmd) && !reGhAuth.MatchString(cmd)) || reGitNet.MatchString(cmd)

	if needsAuth {
		checkGhAuth()
	}
}
