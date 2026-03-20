package main

import (
	"fmt"

	"github.com/pterm/pterm"
)

func printStepHeader(id, name string) {
	pterm.DefaultSection.WithLevel(2).Println(fmt.Sprintf("[%s] %s", id, name))
}

func printStepResult(r StepResult) {
	switch r.Status {
	case StatusSkipped:
		pterm.FgGray.Printfln("  (%s: %s)", r.Status, r.Message)
	case StatusError:
		if r.ExitCode == nil {
			pterm.Error.Printfln("  %s", r.Message)
		}
	}
}

func printAbort(name, detail string) {
	pterm.Error.Printfln("[ABORT] Critical step '%s' failed (%s)", name, detail)
}
