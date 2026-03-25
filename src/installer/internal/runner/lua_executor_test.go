package runner

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"dotfiles/installer/internal/luavm"
	"dotfiles/installer/units"
)

func runnerTestdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

func loadRunnerFixture(t *testing.T, filename string) *luavm.LoadedUnit {
	t.Helper()
	src := filepath.Join(runnerTestdataDir(), filename)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "unit.lua"), data, 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	vm := luavm.NewVM(filepath.Dir(tmpDir))
	lu, err := vm.LoadUnit(tmpDir)
	if err != nil {
		t.Fatalf("failed to load fixture %s: %v", filename, err)
	}
	return lu
}

func TestExecute_VerifyReturnsTrue(t *testing.T) {
	lu := loadRunnerFixture(t, "simple_unit.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})
	result := exec.Execute(
		units.Task{Name: "verify", UnitID: "simple"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if result.Status != units.StatusSuccess {
		t.Fatalf("expected success, got %s: %s", result.Status, result.Message)
	}
}

func TestExecute_VerifyReturnsFalse(t *testing.T) {
	lu := loadRunnerFixture(t, "dep_a.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})
	result := exec.Execute(
		units.Task{Name: "verify", UnitID: "dep_a"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if result.Status != units.StatusError {
		t.Fatalf("expected error, got %s", result.Status)
	}
	if result.Message != "needs setup" {
		t.Fatalf("expected 'needs setup', got %q", result.Message)
	}
}

func TestExecute_SetupNoReturn(t *testing.T) {
	lu := loadRunnerFixture(t, "simple_unit.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})
	result := exec.Execute(
		units.Task{Name: "setup", UnitID: "simple"},
		units.ExecutionPlan{},
		context.Background(),
	)
	// setup doesn't return bool → treated as success.
	if result.Status != units.StatusSuccess {
		t.Fatalf("expected success, got %s: %s", result.Status, result.Message)
	}
}

func TestExecute_ContractWiring(t *testing.T) {
	lu := loadRunnerFixture(t, "contract_verify.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})

	// Run verify first — produces output.
	verifyResult := exec.Execute(
		units.Task{Name: "verify", UnitID: "contract_test"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if verifyResult.Status != units.StatusError || verifyResult.Message != "needs setup" {
		t.Fatalf("verify should return false (needs setup), got %s: %s", verifyResult.Status, verifyResult.Message)
	}

	// Check that the context was populated.
	val, ok := lu.Context.GetData("value")
	if !ok {
		t.Fatal("expected 'value' in context after verify")
	}
	if val != "wired_data" {
		t.Fatalf("expected 'wired_data', got %v", val)
	}

	// Run setup — receives wired input.
	setupResult := exec.Execute(
		units.Task{Name: "setup", UnitID: "contract_test"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if setupResult.Status != units.StatusSuccess {
		t.Fatalf("setup with wired input should succeed, got %s: %s", setupResult.Status, setupResult.Message)
	}
}

func TestExecute_SetupError(t *testing.T) {
	lu := loadRunnerFixture(t, "critical_unit.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})
	result := exec.Execute(
		units.Task{Name: "setup", UnitID: "critical_test"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if result.Status != units.StatusError {
		t.Fatalf("expected error, got %s", result.Status)
	}
	if result.Message == "" {
		t.Fatal("expected error message")
	}
}

func TestExecute_UnknownUnit(t *testing.T) {
	exec := NewLuaExecutor([]*luavm.LoadedUnit{})
	result := exec.Execute(
		units.Task{Name: "verify", UnitID: "nonexistent"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if result.Status != units.StatusError {
		t.Fatalf("expected error for unknown unit, got %s", result.Status)
	}
}

func TestExecute_UnknownTask(t *testing.T) {
	lu := loadRunnerFixture(t, "simple_unit.lua")
	defer lu.State.Close()

	exec := NewLuaExecutor([]*luavm.LoadedUnit{lu})
	result := exec.Execute(
		units.Task{Name: "nonexistent", UnitID: "simple"},
		units.ExecutionPlan{},
		context.Background(),
	)
	if result.Status != units.StatusError {
		t.Fatalf("expected error for unknown task, got %s", result.Status)
	}
}
