package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
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

	var pretty bytes.Buffer
	if json.Indent(&pretty, data, "", "  ") == nil {
		data = pretty.Bytes()
	}

	ts := time.Now().Format("2006-01-02 15:04:05")
	content := fmt.Sprintf("// %s\n%s", ts, data)
	os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
}
