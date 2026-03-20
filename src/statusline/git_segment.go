package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func gitSegment(cwd string) string {
	if cwd == "" {
		return ""
	}
	root := gitCmd(cwd, "rev-parse", "--show-toplevel")
	if root == "" {
		return ""
	}

	branch := gitCmd(cwd, "rev-parse", "--abbrev-ref", "HEAD")
	if branch == "" {
		return ""
	}

	indicators := ""

	// Dirty (unstaged changes or untracked files)
	_, err := runCmd(2*time.Second, cwd, "git", "diff", "--quiet")
	if err != nil {
		indicators += " *"
	} else {
		untracked := gitCmd(cwd, "ls-files", "--others", "--exclude-standard")
		if untracked != "" {
			indicators += " *"
		}
	}

	// Staged changes
	_, err = runCmd(2*time.Second, cwd, "git", "diff", "--cached", "--quiet")
	if err != nil {
		indicators += " \u2713"
	}

	// Insertions/deletions
	numstat := gitCmd(cwd, "diff", "--numstat")
	if numstat != "" {
		ins, del := 0, 0
		for _, line := range strings.Split(numstat, "\n") {
			fields := strings.Fields(line)
			if len(fields) >= 2 && fields[0] != "-" {
				n, _ := strconv.Atoi(fields[0])
				ins += n
				n, _ = strconv.Atoi(fields[1])
				del += n
			}
		}
		if ins > 0 {
			indicators += fmt.Sprintf(" %s+%d%s", CGreen, ins, CPink)
		}
		if del > 0 {
			indicators += fmt.Sprintf(" %s-%d%s", CRed, del, CPink)
		}
	}

	// Ahead/behind remote
	ab := gitCmd(cwd, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if ab != "" {
		parts := strings.Fields(ab)
		if len(parts) == 2 {
			ahead, _ := strconv.Atoi(parts[0])
			behind, _ := strconv.Atoi(parts[1])
			if ahead > 0 {
				indicators += fmt.Sprintf(" %s\u2191%d%s", CCyan, ahead, CPink)
			}
			if behind > 0 {
				indicators += fmt.Sprintf(" %s\u2193%d%s", CCyan, behind, CPink)
			}
		}
	}

	// Branch text with optional GitHub link
	branchText := branch + indicators
	remote := gitCmd(cwd, "remote", "get-url", "origin")
	if remote != "" {
		var repoURL string
		if strings.HasPrefix(remote, "git@github.com:") {
			repoURL = "https://github.com/" + strings.TrimSuffix(remote[15:], ".git")
		} else if strings.HasPrefix(remote, "https://github.com/") {
			repoURL = strings.TrimSuffix(remote, ".git")
		}
		if repoURL != "" {
			branchText = osc8Link(repoURL, branchText)
		}
	}

	return CPink + "\ue725 " + branchText + Rst
}
