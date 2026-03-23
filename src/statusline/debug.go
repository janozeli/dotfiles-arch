package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

var debugEnabled = slices.Contains(os.Args[1:], "--debug")

func debugDump(name string, data []byte) {
	if !debugEnabled {
		return
	}
	home, _ := os.UserHomeDir()
	dir := filepath.Join(home, "dotfiles-arch", "src", "statusline", "logs")
	os.MkdirAll(dir, 0755)

	ext := filepath.Ext(name)
	base := strings.TrimSuffix(name, ext)
	ts := time.Now().Format("20060102-150405")

	// Remove previous dumps with same base name
	matches, _ := filepath.Glob(filepath.Join(dir, base+"-*"+ext))
	for _, f := range matches {
		os.Remove(f)
	}

	var pretty bytes.Buffer
	if json.Indent(&pretty, data, "", "  ") == nil {
		data = pretty.Bytes()
	}
	os.WriteFile(filepath.Join(dir, fmt.Sprintf("%s-%s%s", base, ts, ext)), data, 0644)
}
