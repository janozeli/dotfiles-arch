package luavm

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func testdataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "testdata")
}

// loadFixture copies a fixture Lua file into a temp dir as unit.lua
// so LoadUnit can find it at the expected path.
func loadFixture(t *testing.T, filename string) (*LoadedUnit, error) {
	t.Helper()
	src := filepath.Join(testdataDir(), filename)
	data, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("failed to read fixture %s: %v", filename, err)
	}
	tmpDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tmpDir, "unit.lua"), data, 0644); err != nil {
		t.Fatalf("failed to write fixture: %v", err)
	}
	vm := NewVM(filepath.Dir(tmpDir))
	return vm.LoadUnit(tmpDir)
}

func TestLoadUnit_Valid(t *testing.T) {
	lu, err := loadFixture(t, "valid_unit.lua")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lu.State.Close()

	if lu.Unit.ID != "test_valid" {
		t.Fatalf("expected id 'test_valid', got %q", lu.Unit.ID)
	}
	if lu.Unit.Name != "Valid Test Unit" {
		t.Fatalf("expected name 'Valid Test Unit', got %q", lu.Unit.Name)
	}
	if len(lu.Unit.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(lu.Unit.Stages))
	}
	if lu.Unit.Stages[0].Name != "verify" {
		t.Fatalf("expected first stage 'verify', got %q", lu.Unit.Stages[0].Name)
	}
	if lu.Unit.Stages[1].Name != "setup" {
		t.Fatalf("expected second stage 'setup', got %q", lu.Unit.Stages[1].Name)
	}
	if _, ok := lu.TaskFuncs["verify"]; !ok {
		t.Fatal("missing verify task func")
	}
	if _, ok := lu.TaskFuncs["setup"]; !ok {
		t.Fatal("missing setup task func")
	}
	if _, ok := lu.TaskFuncs["teardown"]; !ok {
		t.Fatal("missing teardown task func")
	}
	if _, ok := lu.Unit.Actions["teardown"]; !ok {
		t.Fatal("missing teardown action")
	}
}

func TestLoadUnit_Contract(t *testing.T) {
	lu, err := loadFixture(t, "contract_unit.lua")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lu.State.Close()

	verify := lu.TaskFuncs["verify"]
	if verify.Output["result"] != "string" {
		t.Fatalf("expected verify output 'result:string', got %v", verify.Output)
	}

	setup := lu.TaskFuncs["setup"]
	if setup.Input["result"] != "string" {
		t.Fatalf("expected setup input 'result:string', got %v", setup.Input)
	}
}

func TestLoadUnit_InvalidNoID(t *testing.T) {
	_, err := loadFixture(t, "invalid_no_id.lua")
	if err == nil {
		t.Fatal("expected error for unit without id")
	}
	if !strings.Contains(err.Error(), "missing id") {
		t.Fatalf("expected 'missing id' error, got: %v", err)
	}
}

func TestLoadUnit_InvalidNoStages(t *testing.T) {
	_, err := loadFixture(t, "invalid_no_stages.lua")
	if err == nil {
		t.Fatal("expected error for unit without stages")
	}
	if !strings.Contains(err.Error(), "stages") {
		t.Fatalf("expected 'stages' error, got: %v", err)
	}
}

func TestLoadUnit_InvalidContract(t *testing.T) {
	_, err := loadFixture(t, "invalid_contract.lua")
	if err == nil {
		t.Fatal("expected error for invalid contract")
	}
	if !strings.Contains(err.Error(), "missing_data") {
		t.Fatalf("expected error about 'missing_data', got: %v", err)
	}
}

func TestLoadUnit_Timeout(t *testing.T) {
	lu, err := loadFixture(t, "valid_unit.lua")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer lu.State.Close()

	if lu.TaskFuncs["setup"].Timeout != 30 {
		t.Fatalf("expected timeout 30, got %d", lu.TaskFuncs["setup"].Timeout)
	}
}
