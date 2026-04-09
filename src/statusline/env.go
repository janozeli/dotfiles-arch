package main

import (
	"os"
	"path/filepath"
	"strings"
)

func envSegment(cwd, projectDir string) string {
	if cwd == "" {
		return ""
	}
	var parts []string

	home := os.Getenv("HOME")

	// Project root (only shown if cwd is inside a subdirectory)
	if projectDir != "" && projectDir != cwd {
		rootDisplay := filepath.Base(projectDir)
		if home != "" && strings.HasPrefix(projectDir, home) {
			rootDisplay = "~" + projectDir[len(home):]
		}
		parts = append(parts, CCyan+"\uf07c "+osc8Link(editorFileURL(projectDir), rootDisplay)+Rst)
	}

	// Directory with ~ shortening and editor deep link
	short := cwd
	if home != "" && strings.HasPrefix(cwd, home) {
		rest := cwd[len(home):]
		if rest == "" {
			short = "~/"
		} else {
			short = "~" + rest
		}
	}
	p := strings.Split(short, "/")
	if len(p) > 3 {
		short = p[0] + "/" + p[1] + "/.../" + p[len(p)-1]
	}
	parts = append(parts, CCyan+"\ue5ff "+osc8Link(editorFileURL(cwd), short)+Rst)

	return strings.Join(parts, Sep)
}
