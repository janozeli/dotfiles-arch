package main

import (
	"fmt"
	"os"
	"path/filepath"

	"dotfiles/installer/internal/luavm"
)

func main() {
	unitsDir := resolveUnitsDir()

	vm := luavm.NewVM(unitsDir)
	loaded, err := vm.LoadAll()
	if err != nil {
		fmt.Fprintf(os.Stderr, "FAIL: %v\n", err)
		os.Exit(1)
	}
	defer luavm.CloseAll(loaded)

	// Report data flows.
	flows := 0
	for _, lu := range loaded {
		dataFlow := describeDataFlow(lu)
		fmt.Printf("  ✓ %s: %s\n", lu.Unit.ID, dataFlow)
		if dataFlow != "no data flow" {
			flows++
		}
	}

	fmt.Printf("\n  %d/%d units valid, %d data flows verified\n", len(loaded), len(loaded), flows)
}

func describeDataFlow(lu *luavm.LoadedUnit) string {
	var producers []string
	var consumers []string

	for name, tf := range lu.TaskFuncs {
		for outName, outType := range tf.Output {
			producers = append(producers, fmt.Sprintf("%s produces %s:%s", name, outName, outType))
		}
		for inName, inType := range tf.Input {
			consumers = append(consumers, fmt.Sprintf("%s consumes %s:%s", name, inName, inType))
		}
	}

	if len(producers) == 0 && len(consumers) == 0 {
		return "no data flow"
	}

	// Build concise description.
	flows := ""
	for _, tf := range lu.TaskFuncs {
		for name, typ := range tf.Output {
			flows += fmt.Sprintf("%s: %s", name, typ)
		}
	}

	if flows == "" {
		return "no data flow"
	}
	return flows
}

func resolveUnitsDir() string {
	// When invoked via `go run -C src/installer`, cwd is the module root.
	// When invoked from repo root, cwd is the repo root.
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Try relative to cwd (module root)
	dir := filepath.Join(cwd, "units")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}

	// Try from repo root
	dir = filepath.Join(cwd, "src", "installer", "units")
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}

	fmt.Fprintf(os.Stderr, "error: units directory not found\n")
	os.Exit(1)
	return ""
}
