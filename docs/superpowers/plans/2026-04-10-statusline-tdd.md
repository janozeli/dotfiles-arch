# Statusline TDD Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add comprehensive test coverage to the statusline Go project by testing existing pure functions and extracting testable logic from impure code, enabling TDD going forward.

**Architecture:** The statusline has two categories of code: (1) pure functions that do string formatting, parsing, and display logic -- these get tests directly; (2) functions that call external commands (git, gh) with inline parsing/formatting -- we extract the pure logic into standalone functions, test those, and update callers to use them.

**Tech Stack:** Go 1.23, standard `testing` package, table-driven tests

---

## File Structure

**New test files:**
- `src/statusline/ansi_test.go` -- tests for ANSI utilities and time formatters
- `src/statusline/git_segment_test.go` -- tests for extracted git parsing/display functions
- `src/statusline/gh_test.go` -- tests for extracted GH auth parsing
- `src/statusline/main_test.go` -- tests for extracted context display
- `src/statusline/assembly_test.go` -- tests for filterEmpty
- `src/statusline/env_test.go` -- tests for envSegment
- `src/statusline/usage_test.go` -- tests for usageSegment

**Modified source files (extractions):**
- `src/statusline/ansi.go` -- extract `formatDuration` from `formatResetCompact`
- `src/statusline/git_segment.go` -- extract `gitPathDisplay`, `remoteToRepoURL`, `parseNumstat`, `parseAheadBehind`
- `src/statusline/gh.go` -- extract `parseGhAuthStatus`
- `src/statusline/main.go` -- extract `contextDisplay`

---

### Task 1: Test ANSI pure functions

**Files:**
- Create: `src/statusline/ansi_test.go`

These functions already exist and work. We add tests to lock in their behavior.

- [ ] **Step 1: Write tests for stripAnsi, visibleWidth, osc8Link, thresholdColor, editorFileURL**

Create `src/statusline/ansi_test.go`:

```go
package main

import (
	"testing"
)

func TestStripAnsi(t *testing.T) {
	tests := []struct {
		name, input, want string
	}{
		{"plain text", "hello", "hello"},
		{"color code", "\x1b[38;2;189;147;249mhello\x1b[0m", "hello"},
		{"osc8 link", "\x1b]8;;https://example.com\x1b\\text\x1b]8;;\x1b\\", "text"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stripAnsi(tt.input); got != tt.want {
				t.Errorf("stripAnsi(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestVisibleWidth(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  int
	}{
		{"ascii", "hello", 5},
		{"with ansi", "\x1b[31mhello\x1b[0m", 5},
		{"emoji", "\U0001F480", 2},
		{"empty", "", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := visibleWidth(tt.input); got != tt.want {
				t.Errorf("visibleWidth(%q) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestOsc8Link(t *testing.T) {
	got := osc8Link("https://example.com", "click")
	want := "\x1b]8;;https://example.com\x1b\\click\x1b]8;;\x1b\\"
	if got != want {
		t.Errorf("osc8Link() = %q, want %q", got, want)
	}
}

func TestThresholdColor(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
		want string
	}{
		{"low", 30, CGreen},
		{"boundary 59", 59, CGreen},
		{"boundary 60", 60, CYellow},
		{"mid", 75, CYellow},
		{"boundary 85", 85, CRed},
		{"high", 99, CRed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := thresholdColor(tt.pct); got != tt.want {
				t.Errorf("thresholdColor(%.0f) = %q, want %q", tt.pct, got, tt.want)
			}
		})
	}
}

func TestEditorFileURL(t *testing.T) {
	tests := []struct {
		name, editor, path, want string
	}{
		{"zed", "zed", "/home/user/f.go", "zed://file/home/user/f.go"},
		{"zeditor", "zeditor", "/home/user/f.go", "zed://file/home/user/f.go"},
		{"code", "code", "/home/user/f.go", "vscode://file/home/user/f.go"},
		{"cursor", "cursor", "/home/user/f.go", "cursor://file/home/user/f.go"},
		{"vim fallback", "vim", "/home/user/f.go", "file:///home/user/f.go"},
		{"empty fallback", "", "/home/user/f.go", "file:///home/user/f.go"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("EDITOR", tt.editor)
			if got := editorFileURL(tt.path); got != tt.want {
				t.Errorf("editorFileURL(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to verify they pass**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run 'TestStripAnsi|TestVisibleWidth|TestOsc8Link|TestThresholdColor|TestEditorFileURL' -v`

Expected: All PASS

- [ ] **Step 3: Commit**

```bash
git add src/statusline/ansi_test.go
git commit -m "test: add tests for ANSI utility functions"
```

---

### Task 2: Extract formatDuration + test time formatters

**Files:**
- Modify: `src/statusline/ansi.go:71-90` -- extract `formatDuration` from `formatResetCompact`
- Modify: `src/statusline/ansi_test.go` -- add time formatter tests

`formatResetCompact` depends on `time.Now()`, making it untestable. We extract the pure duration-formatting logic into `formatDuration(d time.Duration) string`.

- [ ] **Step 1: Write tests for formatDuration, formatResetExact, formatResetExactWeekly**

Add to `src/statusline/ansi_test.go` (update import to include `"time"`):

```go
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		d    time.Duration
		want string
	}{
		{"negative", -1 * time.Hour, ""},
		{"zero", 0, ""},
		{"30min", 30 * time.Minute, "0h30"},
		{"2h15m", 2*time.Hour + 15*time.Minute, "2h15"},
		{"1d3h", 27 * time.Hour, "1d03h"},
		{"3d12h", (3*24 + 12) * time.Hour, "3d12h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatDuration(tt.d); got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}

func TestFormatResetExact(t *testing.T) {
	tests := []struct {
		name     string
		resetsAt int64
		want     string
	}{
		{"zero", 0, ""},
		{"afternoon", time.Date(2026, 4, 10, 14, 30, 0, 0, time.Local).Unix(), "14h30"},
		{"midnight", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local).Unix(), "00h00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatResetExact(tt.resetsAt); got != tt.want {
				t.Errorf("formatResetExact(%d) = %q, want %q", tt.resetsAt, got, tt.want)
			}
		})
	}
}

func TestFormatResetExactWeekly(t *testing.T) {
	tests := []struct {
		name     string
		resetsAt int64
		want     string
	}{
		{"zero", 0, ""},
		// 2026-04-10 is a Friday
		{"friday", time.Date(2026, 4, 10, 14, 30, 0, 0, time.Local).Unix(), "fri 14h30"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := formatResetExactWeekly(tt.resetsAt); got != tt.want {
				t.Errorf("formatResetExactWeekly(%d) = %q, want %q", tt.resetsAt, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests -- formatDuration should FAIL (function doesn't exist)**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run 'TestFormatDuration|TestFormatResetExact' -v`

Expected: FAIL -- `undefined: formatDuration`

- [ ] **Step 3: Extract formatDuration from formatResetCompact in ansi.go**

Replace `formatResetCompact` (lines 71-90 in `ansi.go`) with:

```go
func formatDuration(d time.Duration) string {
	if d <= 0 {
		return ""
	}
	totalHours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	days := totalHours / 24
	hours := totalHours % 24
	if days > 0 {
		return fmt.Sprintf("%dd%02dh", days, hours)
	}
	return fmt.Sprintf("%dh%02d", totalHours, minutes)
}

func formatResetCompact(resetsAt int64) string {
	if resetsAt == 0 {
		return ""
	}
	return formatDuration(time.Until(time.Unix(resetsAt, 0)))
}
```

- [ ] **Step 4: Run tests -- all should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run 'TestFormatDuration|TestFormatResetExact' -v`

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add src/statusline/ansi.go src/statusline/ansi_test.go
git commit -m "refactor: extract formatDuration, add time formatter tests"
```

---

### Task 3: Extract + test gitPathDisplay and remoteToRepoURL

**Files:**
- Modify: `src/statusline/git_segment.go:114-123` -- extract `gitPathDisplay`
- Modify: `src/statusline/git_segment.go:100-111` -- extract `remoteToRepoURL`
- Create: `src/statusline/git_segment_test.go`

The path display and remote URL conversion logic in `gitSegment` are pure string operations buried inside a function that calls git commands. Extract them.

- [ ] **Step 1: Write tests for gitPathDisplay and remoteToRepoURL**

Create `src/statusline/git_segment_test.go`:

```go
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
```

- [ ] **Step 2: Run tests -- FAIL (functions don't exist)**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run 'TestGitPathDisplay|TestRemoteToRepoURL' -v`

Expected: FAIL -- `undefined: gitPathDisplay`

- [ ] **Step 3: Extract both functions in git_segment.go**

Add these functions before `gitSegment`:

```go
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
```

Update `gitSegment` -- replace the path display block (lines 114-123):

```go
	// OLD:
	if root != cwd {
		rel, _ := filepath.Rel(root, cwd)
		repoBase := filepath.Base(root)
		relParts := strings.Split(rel, "/")
		if len(relParts) > 2 {
			rel = ".../" + relParts[len(relParts)-1]
		}
		display := repoBase + "/" + rel
		seg += Sep + CCyan + "\uf07c " + osc8Link(editorFileURL(cwd), display) + Rst
	}

	// NEW:
	if display := gitPathDisplay(root, cwd); display != "" {
		seg += Sep + CCyan + "\uf07c " + osc8Link(editorFileURL(cwd), display) + Rst
	}
```

Replace the remote URL block (lines 100-111):

```go
	// OLD:
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

	// NEW:
	if repoURL := remoteToRepoURL(gitCmd(cwd, "remote", "get-url", "origin")); repoURL != "" {
		branchText = osc8Link(repoURL, branchText)
	}
```

- [ ] **Step 4: Run tests -- all should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -v ./...`

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add src/statusline/git_segment.go src/statusline/git_segment_test.go
git commit -m "refactor: extract gitPathDisplay and remoteToRepoURL, add tests"
```

---

### Task 4: Extract + test parseNumstat and parseAheadBehind

**Files:**
- Modify: `src/statusline/git_segment.go:62-79` -- extract `parseNumstat`
- Modify: `src/statusline/git_segment.go:83-96` -- extract `parseAheadBehind`
- Modify: `src/statusline/git_segment_test.go`

The numstat and ahead/behind parsing in `gitSegment` are pure string-to-int operations. Extract them.

- [ ] **Step 1: Write tests for parseNumstat and parseAheadBehind**

Add to `src/statusline/git_segment_test.go`:

```go
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
```

- [ ] **Step 2: Run tests -- FAIL (functions don't exist)**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run 'TestParseNumstat|TestParseAheadBehind' -v`

Expected: FAIL -- `undefined: parseNumstat`

- [ ] **Step 3: Extract both functions in git_segment.go**

Add these functions before `gitSegment`:

```go
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
```

Update `gitSegment` -- replace numstat block:

```go
	// OLD:
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

	// NEW:
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
```

Replace ahead/behind block:

```go
	// OLD:
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

	// NEW:
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
```

- [ ] **Step 4: Run all tests -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -v ./...`

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add src/statusline/git_segment.go src/statusline/git_segment_test.go
git commit -m "refactor: extract parseNumstat and parseAheadBehind, add tests"
```

---

### Task 5: Extract + test parseGhAuthStatus

**Files:**
- Modify: `src/statusline/gh.go:17-28` -- extract `parseGhAuthStatus`
- Create: `src/statusline/gh_test.go`

The `getGhAccount` function runs `gh auth status` and parses the output inline. Extract the parsing logic.

- [ ] **Step 1: Write test for parseGhAuthStatus**

Create `src/statusline/gh_test.go`:

```go
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
```

- [ ] **Step 2: Run test -- FAIL (function doesn't exist)**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run TestParseGhAuthStatus -v`

Expected: FAIL -- `undefined: parseGhAuthStatus`

- [ ] **Step 3: Extract parseGhAuthStatus in gh.go**

Add the new function:

```go
func parseGhAuthStatus(output string) string {
	var current string
	for _, line := range strings.Split(output, "\n") {
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
```

Replace `getGhAccount` body:

```go
func getGhAccount() string {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "gh", "auth", "status")
	out, _ := cmd.CombinedOutput()
	return parseGhAuthStatus(string(out))
}
```

- [ ] **Step 4: Run all tests -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -v ./...`

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add src/statusline/gh.go src/statusline/gh_test.go
git commit -m "refactor: extract parseGhAuthStatus, add test"
```

---

### Task 6: Extract + test contextDisplay

**Files:**
- Modify: `src/statusline/main.go:66-91` -- extract `contextDisplay`
- Create: `src/statusline/main_test.go`

The context window rendering in `main()` is a pure function of percentage and total tokens. Extract it.

- [ ] **Step 1: Write test for contextDisplay**

Create `src/statusline/main_test.go`:

```go
package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestContextDisplay(t *testing.T) {
	tests := []struct {
		name         string
		usedPct      int
		totalTokens  int
		wantColor    string
		wantTokenStr string
	}{
		{"zero", 0, 200000, CGreen, "0"},
		{"low", 10, 200000, CGreen, "20k"},
		{"boundary 49", 49, 200000, CGreen, "98k"},
		{"boundary 50", 50, 200000, CYellow, "100k"},
		{"boundary 64", 64, 200000, CYellow, "128k"},
		{"boundary 65", 65, 200000, CRed, "130k"},
		{"boundary 79", 79, 200000, CRed, "158k"},
		{"boundary 80", 80, 200000, "\x1b[5;38;2;255;85;85m", "160k"},
		{"critical", 90, 200000, "\x1b[5;38;2;255;85;85m", "180k"},
		{"small tokens", 50, 1000, CYellow, "500"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contextDisplay(tt.usedPct, tt.totalTokens)
			if !strings.Contains(got, tt.wantColor) {
				t.Errorf("contextDisplay(%d, %d) missing color %q", tt.usedPct, tt.totalTokens, tt.wantColor)
			}
			visible := stripAnsi(got)
			if !strings.Contains(visible, tt.wantTokenStr) {
				t.Errorf("contextDisplay(%d, %d) visible=%q, want token %q", tt.usedPct, tt.totalTokens, visible, tt.wantTokenStr)
			}
			wantPct := fmt.Sprintf("%d%%", tt.usedPct)
			if !strings.Contains(visible, wantPct) {
				t.Errorf("contextDisplay(%d, %d) visible=%q, want pct %q", tt.usedPct, tt.totalTokens, visible, wantPct)
			}
		})
	}
}
```

- [ ] **Step 2: Run test -- FAIL (function doesn't exist)**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run TestContextDisplay -v`

Expected: FAIL -- `undefined: contextDisplay`

- [ ] **Step 3: Extract contextDisplay in main.go**

Add the function:

```go
func contextDisplay(usedPct, totalTokens int) string {
	usedTokens := totalTokens * usedPct / 100
	var tokenStr string
	if usedTokens >= 1000 {
		tokenStr = fmt.Sprintf("%dk", (usedTokens+500)/1000)
	} else {
		tokenStr = fmt.Sprintf("%d", usedTokens)
	}

	filled := usedPct / 10
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", 10-filled)

	if usedPct < 50 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CGreen, tokenStr, bar, usedPct, Rst)
	} else if usedPct < 65 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CYellow, tokenStr, bar, usedPct, Rst)
	} else if usedPct < 80 {
		return fmt.Sprintf(" %s%s %s %d%%%s", CRed, tokenStr, bar, usedPct, Rst)
	}
	return fmt.Sprintf(" \x1b[5;38;2;255;85;85m\U0001F480 %s %s %d%%%s", tokenStr, bar, usedPct, Rst)
}
```

Replace lines 66-91 in `main()`:

```go
	// OLD: (22 lines of inline context display logic)
	// NEW:
	ctx := ""
	if data.ContextWindow.UsedPercentage != nil {
		ctx = contextDisplay(*data.ContextWindow.UsedPercentage, data.ContextWindow.TotalTokens)
	}
```

- [ ] **Step 4: Run all tests -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -v ./...`

Expected: All PASS

- [ ] **Step 5: Commit**

```bash
git add src/statusline/main.go src/statusline/main_test.go
git commit -m "refactor: extract contextDisplay, add test"
```

---

### Task 7: Test filterEmpty

**Files:**
- Create: `src/statusline/assembly_test.go`

- [ ] **Step 1: Write test for filterEmpty**

Create `src/statusline/assembly_test.go`:

```go
package main

import (
	"slices"
	"testing"
)

func TestFilterEmpty(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  []string
	}{
		{"all empty", []string{"", "", ""}, nil},
		{"no empty", []string{"a", "b"}, []string{"a", "b"}},
		{"mixed", []string{"a", "", "b"}, []string{"a", "b"}},
		{"none", []string{}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterEmpty(tt.input...)
			if !slices.Equal(got, tt.want) {
				t.Errorf("filterEmpty(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: Run test -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run TestFilterEmpty -v`

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add src/statusline/assembly_test.go
git commit -m "test: add filterEmpty test"
```

---

### Task 8: Test envSegment

**Files:**
- Create: `src/statusline/env_test.go`

Test `envSegment` by setting env vars and checking visible output (ANSI-stripped).

- [ ] **Step 1: Write test for envSegment**

Create `src/statusline/env_test.go`:

```go
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
```

- [ ] **Step 2: Run test -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run TestEnvSegment -v`

Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add src/statusline/env_test.go
git commit -m "test: add envSegment tests"
```

---

### Task 9: Test usageSegment

**Files:**
- Create: `src/statusline/usage_test.go`

Test `usageSegment` with zero-valued `ResetsAt` to avoid time-dependent output.

- [ ] **Step 1: Write test for usageSegment**

Create `src/statusline/usage_test.go`:

```go
package main

import (
	"strings"
	"testing"
)

func TestUsageSegment(t *testing.T) {
	tests := []struct {
		name         string
		limits       RateLimits
		wantEmpty    bool
		wantContains []string
	}{
		{"empty", RateLimits{}, true, nil},
		{"five hour only", func() RateLimits {
			var l RateLimits
			l.FiveHour.UsedPercentage = 50
			return l
		}(), false, []string{"Session:", "50%"}},
		{"seven day only", func() RateLimits {
			var l RateLimits
			l.SevenDay.UsedPercentage = 30
			return l
		}(), false, []string{"Weekly:", "30%"}},
		{"both", func() RateLimits {
			var l RateLimits
			l.FiveHour.UsedPercentage = 50
			l.SevenDay.UsedPercentage = 30
			return l
		}(), false, []string{"Session:", "50%", "Weekly:", "30%"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := usageSegment(tt.limits)
			if tt.wantEmpty {
				if got != "" {
					t.Errorf("usageSegment() = %q, want empty", got)
				}
				return
			}
			visible := stripAnsi(got)
			for _, want := range tt.wantContains {
				if !strings.Contains(visible, want) {
					t.Errorf("usageSegment() visible=%q, want contains %q", visible, want)
				}
			}
		})
	}
}
```

- [ ] **Step 2: Run test -- should PASS**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -run TestUsageSegment -v`

Expected: PASS

- [ ] **Step 3: Run full test suite**

Run: `cd /home/janozeli/dotfiles-arch/src/statusline && go test -v ./...`

Expected: All PASS

- [ ] **Step 4: Commit**

```bash
git add src/statusline/usage_test.go
git commit -m "test: add usageSegment tests"
```
