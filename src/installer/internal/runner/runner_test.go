package runner

import (
	"context"
	"testing"

	"dotfiles/installer/internal/luavm"
	"dotfiles/installer/units"
)

func buildTestRunner(t *testing.T, loaded []*luavm.LoadedUnit, opts RunOpts) (*Runner, *DAG, []*luavm.LoadedUnit) {
	t.Helper()
	allUnits := make([]units.Unit, len(loaded))
	for i, lu := range loaded {
		allUnits[i] = lu.Unit
	}
	dag, err := NewDAG(allUnits)
	if err != nil {
		t.Fatalf("failed to build DAG: %v", err)
	}
	manifest := NewManifest(dag, allUnits)
	exec := NewLuaExecutor(loaded)
	r := NewRunner(allUnits, dag, manifest, exec, opts)
	return r, dag, loaded
}

func TestRunner_SkipSetupWhenVerifyPasses(t *testing.T) {
	lu := loadRunnerFixture(t, "simple_unit.lua")
	defer lu.State.Close()

	r, _, _ := buildTestRunner(t, []*luavm.LoadedUnit{lu}, RunOpts{})
	results, aborted := r.Run(context.Background())
	if aborted {
		t.Fatal("unexpected abort")
	}

	// verify returns true → setup should be skipped.
	var setupResult *units.TaskResult
	for _, res := range results {
		if res.ID == "simple/setup/setup" {
			setupResult = &res
		}
	}
	if setupResult == nil {
		t.Fatal("missing setup result")
	}
	if setupResult.Status != units.StatusSkipped {
		t.Fatalf("expected setup to be skipped, got %s", setupResult.Status)
	}
}

func TestRunner_ForceRunsSetupEvenWhenVerifyPasses(t *testing.T) {
	lu := loadRunnerFixture(t, "simple_unit.lua")
	defer lu.State.Close()

	r, _, _ := buildTestRunner(t, []*luavm.LoadedUnit{lu}, RunOpts{Force: true})
	results, aborted := r.Run(context.Background())
	if aborted {
		t.Fatal("unexpected abort")
	}

	var setupResult *units.TaskResult
	for _, res := range results {
		if res.ID == "simple/setup/setup" {
			setupResult = &res
		}
	}
	if setupResult == nil {
		t.Fatal("missing setup result")
	}
	if setupResult.Status != units.StatusSuccess {
		t.Fatalf("expected setup to run with --force, got %s: %s", setupResult.Status, setupResult.Message)
	}
}

func TestRunner_VerifyNeedsSetupDoesNotAbortCritical(t *testing.T) {
	lu := loadRunnerFixture(t, "dep_a.lua")
	defer lu.State.Close()

	// Make the unit critical.
	lu.Unit.Critical = true

	r, _, _ := buildTestRunner(t, []*luavm.LoadedUnit{lu}, RunOpts{})
	_, aborted := r.Run(context.Background())

	// verify returns false ("needs setup") on a critical unit.
	// This should NOT abort — only setup failures should abort.
	if aborted {
		t.Fatal("critical unit should not abort on verify 'needs setup'")
	}
}

func TestRunner_CriticalSetupFailureAborts(t *testing.T) {
	lu := loadRunnerFixture(t, "critical_unit.lua")
	defer lu.State.Close()

	r, _, _ := buildTestRunner(t, []*luavm.LoadedUnit{lu}, RunOpts{})
	_, aborted := r.Run(context.Background())

	// verify returns false, setup throws error() on a critical unit → should abort.
	if !aborted {
		t.Fatal("expected abort on critical setup failure")
	}
}

func TestRunner_DependencyOrdering(t *testing.T) {
	luA := loadRunnerFixture(t, "dep_a.lua")
	luB := loadRunnerFixture(t, "dep_b.lua")
	defer luA.State.Close()
	defer luB.State.Close()

	r, _, _ := buildTestRunner(t, []*luavm.LoadedUnit{luA, luB}, RunOpts{})
	results, aborted := r.Run(context.Background())
	if aborted {
		t.Fatal("unexpected abort")
	}

	// Find execution order of setup tasks.
	aSetupIdx, bSetupIdx := -1, -1
	for i, res := range results {
		if res.ID == "dep_a/setup/setup" {
			aSetupIdx = i
		}
		if res.ID == "dep_b/setup/setup" {
			bSetupIdx = i
		}
	}
	if aSetupIdx == -1 || bSetupIdx == -1 {
		t.Fatal("missing setup results")
	}
	if bSetupIdx <= aSetupIdx {
		t.Fatalf("dep_b/setup (idx %d) should run after dep_a/setup (idx %d)", bSetupIdx, aSetupIdx)
	}
}
