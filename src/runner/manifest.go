package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// Manifest records the execution plan and results.
type Manifest struct {
	Host       string       `json:"host"`
	User       string       `json:"user"`
	StartedAt  string       `json:"started_at"`
	FinishedAt string       `json:"finished_at"`
	Status     string       `json:"status"`
	Waves      []WaveResult `json:"waves"`
	Units      []UnitInfo   `json:"units"`

	mu sync.Mutex
}

// WaveResult holds the tasks and results for a single execution wave.
type WaveResult struct {
	Index int          `json:"index"`
	Tasks []TaskResult `json:"tasks"`
}

// UnitInfo is a summary of a unit for the manifest.
type UnitInfo struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Critical bool              `json:"critical,omitempty"`
	Stages   []string          `json:"stages"`
	Actions  map[string]string `json:"actions,omitempty"`
}

// NewManifest pre-populates a manifest from the DAG with all tasks pending.
func NewManifest(dag *DAG, units []Unit) *Manifest {
	hostname, _ := os.Hostname()
	user := os.Getenv("USER")
	if user == "" {
		user = "unknown"
	}

	m := &Manifest{
		Host:      hostname,
		User:      user,
		StartedAt: nowISO(),
	}

	// Populate waves with pending tasks.
	for i, wave := range dag.Waves() {
		wr := WaveResult{Index: i}
		for _, node := range wave {
			wr.Tasks = append(wr.Tasks, TaskResult{
				ID:     node.GlobalID,
				UnitID: node.UnitID,
				Status: "pending",
			})
		}
		m.Waves = append(m.Waves, wr)
	}

	// Populate unit info.
	for _, unit := range units {
		info := UnitInfo{
			ID:       unit.ID,
			Name:     unit.Name,
			Critical: unit.Critical,
		}
		for _, stage := range unit.Stages {
			info.Stages = append(info.Stages, stage.Name)
		}
		if len(unit.Actions) > 0 {
			info.Actions = make(map[string]string)
			for name, action := range unit.Actions {
				info.Actions[name] = action.Description
			}
		}
		m.Units = append(m.Units, info)
	}

	return m
}

// UpdateTask sets the result for a specific task by its global ID.
func (m *Manifest) UpdateTask(globalID string, result TaskResult) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.Waves {
		for j := range m.Waves[i].Tasks {
			if m.Waves[i].Tasks[j].ID == globalID {
				m.Waves[i].Tasks[j] = result
				m.Waves[i].Tasks[j].ID = globalID
				return
			}
		}
	}
}

// DeriveStatus determines the overall run status from task results.
func (m *Manifest) DeriveStatus(aborted bool) string {
	if aborted {
		return "error"
	}
	hasError := false
	hasWarning := false
	for _, wave := range m.Waves {
		for _, task := range wave.Tasks {
			if task.Status == StatusError {
				hasError = true
			}
			if task.Status == StatusWarning {
				hasWarning = true
			}
		}
	}
	if hasError {
		return "partial"
	}
	if hasWarning {
		return "warning"
	}
	return "success"
}

// Save writes the manifest to disk as JSON.
func (m *Manifest) Save(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o644)
}
