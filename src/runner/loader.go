package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LoadUnits discovers and parses all unit.yaml files under unitsDir.
// It resolves task paths, validates references, and applies defaults.
func LoadUnits(unitsDir string) ([]Unit, error) {
	entries, err := os.ReadDir(unitsDir)
	if err != nil {
		return nil, fmt.Errorf("read units dir: %w", err)
	}

	var units []Unit
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitDir := filepath.Join(unitsDir, entry.Name())
		yamlPath := filepath.Join(unitDir, "unit.yaml")

		if _, err := os.Stat(yamlPath); err != nil {
			continue // skip directories without unit.yaml
		}

		unit, err := loadUnit(unitDir, yamlPath)
		if err != nil {
			return nil, fmt.Errorf("load unit %s: %w", entry.Name(), err)
		}

		units = append(units, unit)
	}

	if err := validateDependencies(units); err != nil {
		return nil, err
	}

	return units, nil
}

func loadUnit(unitDir, yamlPath string) (Unit, error) {
	data, err := os.ReadFile(yamlPath)
	if err != nil {
		return Unit{}, fmt.Errorf("read yaml: %w", err)
	}

	var unit Unit
	if err := yaml.Unmarshal(data, &unit); err != nil {
		return Unit{}, fmt.Errorf("parse yaml: %w", err)
	}

	unit.Dir = unitDir

	if err := unit.Validate(); err != nil {
		return Unit{}, err
	}

	// Resolve task paths and apply defaults.
	for i := range unit.Stages {
		stage := &unit.Stages[i]
		applyStageDefaults(stage)

		for j := range stage.Tasks {
			task := &stage.Tasks[j]
			if task.Name == "" {
				return Unit{}, fmt.Errorf("stage %q: task at index %d missing 'task' field", stage.Name, j)
			}
			task.Path = filepath.Join(unitDir, "tasks", task.Name)
			if _, err := os.Stat(task.Path); err != nil {
				return Unit{}, fmt.Errorf("stage %q: task file not found: %s", stage.Name, task.Path)
			}
		}
	}

	// Resolve action task paths.
	for name, action := range unit.Actions {
		actionPath := filepath.Join(unitDir, "tasks", action.Task)
		if _, err := os.Stat(actionPath); err != nil {
			return Unit{}, fmt.Errorf("action %q: task file not found: %s", name, actionPath)
		}
	}

	return unit, nil
}

func applyStageDefaults(stage *Stage) {
	if stage.ExecutionPlan.Shell == "" {
		stage.ExecutionPlan.Shell = "bash"
	}
}

func validateDependencies(units []Unit) error {
	knownIDs := make(map[string]bool, len(units))
	for _, u := range units {
		knownIDs[u.ID] = true
	}

	for _, u := range units {
		for _, stage := range u.Stages {
			for _, task := range stage.Tasks {
				for _, dep := range task.DependsOn {
					// Local dependency (within same unit) — just a task name.
					if !strings.Contains(dep, ".") && !strings.Contains(dep, "/") {
						continue
					}

					// Cross-unit dependency — unit ID.
					unitID := dep
					if idx := strings.IndexAny(dep, "./"); idx > 0 {
						unitID = dep[:idx]
					}
					if !knownIDs[unitID] {
						return fmt.Errorf("unit %q, task %q: unknown dependency %q", u.ID, task.Name, dep)
					}
				}
			}
		}
	}

	return nil
}
