package runner

import (
	"testing"

	"dotfiles/installer/units"
)

func makeUnit(id string, deps []string, stages ...string) units.Unit {
	if len(stages) == 0 {
		stages = []string{"verify", "setup"}
	}
	u := units.Unit{
		ID:      id,
		Name:    id,
		Actions: map[string]units.Action{"teardown": {Task: "teardown"}},
	}
	for i, name := range stages {
		task := units.Task{Name: name, UnitID: id}
		if i > 0 && len(deps) > 0 {
			task.DependsOn = deps
		}
		u.Stages = append(u.Stages, units.Stage{
			Name:  name,
			Tasks: []units.Task{task},
		})
	}
	return u
}

func TestDAG_SingleUnit(t *testing.T) {
	dag, err := NewDAG([]units.Unit{makeUnit("a", nil)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 2 waves: verify, then setup.
	if len(dag.Waves()) != 2 {
		t.Fatalf("expected 2 waves, got %d", len(dag.Waves()))
	}
	if len(dag.Waves()[0]) != 1 || dag.Waves()[0][0].StageName != "verify" {
		t.Fatal("wave 0 should be verify")
	}
	if len(dag.Waves()[1]) != 1 || dag.Waves()[1][0].StageName != "setup" {
		t.Fatal("wave 1 should be setup")
	}
}

func TestDAG_ParallelUnits(t *testing.T) {
	dag, err := NewDAG([]units.Unit{
		makeUnit("a", nil),
		makeUnit("b", nil),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Wave 0: both verifies. Wave 1: both setups.
	if len(dag.Waves()) != 2 {
		t.Fatalf("expected 2 waves, got %d", len(dag.Waves()))
	}
	if len(dag.Waves()[0]) != 2 {
		t.Fatalf("expected 2 tasks in wave 0, got %d", len(dag.Waves()[0]))
	}
	if len(dag.Waves()[1]) != 2 {
		t.Fatalf("expected 2 tasks in wave 1, got %d", len(dag.Waves()[1]))
	}
}

func TestDAG_DependencyChain(t *testing.T) {
	// a → b: b depends on a.
	dag, err := NewDAG([]units.Unit{
		makeUnit("a", nil),
		makeUnit("b", []string{"a"}),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Wave 0: a/verify, b/verify
	// Wave 1: a/setup
	// Wave 2: b/setup (must wait for a/setup)
	if len(dag.Waves()) != 3 {
		t.Fatalf("expected 3 waves, got %d", len(dag.Waves()))
	}

	// Verify wave 2 contains b/setup.
	wave2 := dag.Waves()[2]
	if len(wave2) != 1 || wave2[0].UnitID != "b" || wave2[0].StageName != "setup" {
		t.Fatalf("expected wave 2 = [b/setup], got %v", wave2[0].GlobalID)
	}
}

func TestDAG_DependencyResolvesToSetup(t *testing.T) {
	// This is the bug we fixed: dependency must resolve to setup, not verify.
	dag, err := NewDAG([]units.Unit{
		makeUnit("a", nil),
		makeUnit("b", []string{"a"}),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// b/setup must be in a later wave than a/setup.
	aSetupWave, bSetupWave := -1, -1
	for i, wave := range dag.Waves() {
		for _, node := range wave {
			if node.GlobalID == "a/setup/setup" {
				aSetupWave = i
			}
			if node.GlobalID == "b/setup/setup" {
				bSetupWave = i
			}
		}
	}
	if bSetupWave <= aSetupWave {
		t.Fatalf("b/setup (wave %d) must be after a/setup (wave %d)", bSetupWave, aSetupWave)
	}
}

func TestDAG_ThreeUnitChain(t *testing.T) {
	// a → b → c
	dag, err := NewDAG([]units.Unit{
		makeUnit("a", nil),
		makeUnit("b", []string{"a"}),
		makeUnit("c", []string{"b"}),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 4 waves: all verify, a/setup, b/setup, c/setup.
	if len(dag.Waves()) != 4 {
		t.Fatalf("expected 4 waves, got %d", len(dag.Waves()))
	}
}

func TestDAG_CycleDetected(t *testing.T) {
	// a depends on b, b depends on a.
	_, err := NewDAG([]units.Unit{
		makeUnit("a", []string{"b"}),
		makeUnit("b", []string{"a"}),
	})
	if err == nil {
		t.Fatal("expected cycle error")
	}
	if !contains(err.Error(), "cycle") {
		t.Fatalf("expected cycle error, got: %v", err)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
