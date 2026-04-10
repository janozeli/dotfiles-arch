package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func autoFetch(root string) {
	fetchHead := filepath.Join(root, ".git", "FETCH_HEAD")
	if info, err := os.Stat(fetchHead); err == nil && time.Since(info.ModTime()) < 3*time.Minute {
		return
	}
	cmd := exec.Command("git", "-C", root, "fetch", "--quiet")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	cmd.Start()
}

func gitPathDisplay(root, cwd string) string {
	if root == cwd {
		return ""
	}
	rel, _ := filepath.Rel(root, cwd)
	repoBase := filepath.Base(root)
	relParts := strings.Split(rel, "/")
	if len(relParts) > 2 {
		rel = ".../" + relParts[len(relParts)-1]
	}
	return repoBase + "/" + rel
}

func remoteToRepoURL(remote string) string {
	if strings.HasPrefix(remote, "git@github.com:") {
		return "https://github.com/" + strings.TrimSuffix(remote[15:], ".git")
	}
	if strings.HasPrefix(remote, "https://github.com/") {
		return strings.TrimSuffix(remote, ".git")
	}
	return ""
}

func parseNumstat(numstat string) (ins, del int) {
	for _, line := range strings.Split(numstat, "\n") {
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[0] != "-" {
			n, _ := strconv.Atoi(fields[0])
			ins += n
			n, _ = strconv.Atoi(fields[1])
			del += n
		}
	}
	return
}

func parseAheadBehind(ab string) (ahead, behind int) {
	parts := strings.Fields(ab)
	if len(parts) == 2 {
		ahead, _ = strconv.Atoi(parts[0])
		behind, _ = strconv.Atoi(parts[1])
	}
	return
}

func gitSegment(cwd string) string {
	if cwd == "" {
		return ""
	}
	root := gitCmd(cwd, "rev-parse", "--show-toplevel")
	if root == "" {
		return ""
	}

	autoFetch(root)

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
		ins, del := parseNumstat(numstat)
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
		ahead, behind := parseAheadBehind(ab)
		if ahead > 0 {
			indicators += fmt.Sprintf(" %s\u2191%d%s", CCyan, ahead, CPink)
		}
		if behind > 0 {
			indicators += fmt.Sprintf(" %s\u2193%d%s", CCyan, behind, CPink)
		}
	}

	// Branch text with optional GitHub link
	branchText := branch + indicators
	if repoURL := remoteToRepoURL(gitCmd(cwd, "remote", "get-url", "origin")); repoURL != "" {
		branchText = osc8Link(repoURL, branchText)
	}

	seg := CPink + "\ue725 " + branchText + Rst
	if display := gitPathDisplay(root, cwd); display != "" {
		seg += Sep + CCyan + "\uf07c " + osc8Link(editorFileURL(cwd), display) + Rst
	}
	return seg
}
