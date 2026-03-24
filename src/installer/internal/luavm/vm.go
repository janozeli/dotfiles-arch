package luavm

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"dotfiles/installer/units"

	lua "github.com/yuin/gopher-lua"
)

// TaskFunc holds a Lua function and its declared signature for a single task.
type TaskFunc struct {
	Name    string
	Fn      *lua.LFunction
	Input   map[string]string
	Output  map[string]string
	Timeout int
}

// LoadedUnit holds a parsed unit with its Lua task functions and state.
type LoadedUnit struct {
	Unit      units.Unit
	TaskFuncs map[string]*TaskFunc
	State     *lua.LState
	Context   *contextStore
}

// unitBuilder accumulates declarations from a unit.lua file.
type unitBuilder struct {
	unit       units.Unit
	dependsOn  []string
	tasks      map[string]*TaskFunc
	stageOrder []string
	actions    map[string]units.Action
}

// build validates and produces the final Unit from accumulated declarations.
func (b *unitBuilder) build() (units.Unit, error) {
	if b.unit.ID == "" {
		return units.Unit{}, fmt.Errorf("unit{} missing id")
	}
	if b.unit.Name == "" {
		return units.Unit{}, fmt.Errorf("unit{} missing name")
	}
	if len(b.stageOrder) == 0 {
		return units.Unit{}, fmt.Errorf("stages{} not declared")
	}

	// Build stages from stage order + task declarations.
	for _, name := range b.stageOrder {
		tf, ok := b.tasks[name]
		if !ok {
			return units.Unit{}, fmt.Errorf("stage %q references undefined task", name)
		}

		task := units.Task{
			Name:   name,
			UnitID: b.unit.ID,
		}

		// If this task has dependencies from unit.depends_on, apply them
		// to the first stage's task (same as current DAG behavior).
		if name == b.stageOrder[0] && len(b.dependsOn) == 0 {
			// No deps — leave empty.
		}

		stage := units.Stage{
			Name:  name,
			Tasks: []units.Task{task},
		}

		if tf.Timeout > 0 {
			stage.ExecutionPlan.Timeout = tf.Timeout
		}

		b.unit.Stages = append(b.unit.Stages, stage)
	}

	// Apply cross-unit dependencies to the setup task's depends_on.
	// Find the task that is NOT the first stage (verify) — typically "setup".
	if len(b.dependsOn) > 0 && len(b.unit.Stages) > 1 {
		for i := 1; i < len(b.unit.Stages); i++ {
			for j := range b.unit.Stages[i].Tasks {
				b.unit.Stages[i].Tasks[j].DependsOn = b.dependsOn
			}
		}
	}

	// Build actions.
	b.unit.Actions = b.actions

	// Validate required stages and actions.
	if err := b.unit.Validate(); err != nil {
		return units.Unit{}, err
	}

	return b.unit, nil
}

// VM manages Lua state lifecycle for loading units.
type VM struct {
	unitsDir string
}

// NewVM creates a VM that discovers units in the given directory.
func NewVM(unitsDir string) *VM {
	return &VM{unitsDir: unitsDir}
}

// LoadUnit loads a single unit.lua file from the given directory.
func (v *VM) LoadUnit(unitDir string) (*LoadedUnit, error) {
	luaPath := filepath.Join(unitDir, "unit.lua")

	L := lua.NewState(lua.Options{SkipOpenLibs: true})
	// Open only safe libs.
	for _, pair := range []struct {
		name string
		fn   lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage},
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
		{lua.StringLibName, lua.OpenString},
		{lua.MathLibName, lua.OpenMath},
	} {
		L.Push(L.NewFunction(pair.fn))
		L.Push(lua.LString(pair.name))
		L.Call(1, 0)
	}

	builder := &unitBuilder{
		tasks:   make(map[string]*TaskFunc),
		actions: make(map[string]units.Action),
	}
	builder.unit.Dir = unitDir

	ctx := newContextStore()

	// Register all API functions.
	registerUnitAPI(L, builder)
	registerShellAPI(L)
	registerFsAPI(L, unitDir)
	registerContextAPI(L, ctx)
	registerUtilAPI(L)

	// Execute the unit.lua file.
	if err := L.DoFile(luaPath); err != nil {
		L.Close()
		return nil, fmt.Errorf("execute %s: %w", luaPath, err)
	}

	// Build the unit from collected declarations.
	unit, err := builder.build()
	if err != nil {
		L.Close()
		return nil, fmt.Errorf("build %s: %w", filepath.Base(unitDir), err)
	}

	return &LoadedUnit{
		Unit:      unit,
		TaskFuncs: builder.tasks,
		State:     L,
		Context:   ctx,
	}, nil
}

// LoadAll discovers and loads all unit.lua files under the units directory.
func (v *VM) LoadAll() ([]*LoadedUnit, error) {
	entries, err := os.ReadDir(v.unitsDir)
	if err != nil {
		return nil, fmt.Errorf("read units dir: %w", err)
	}

	var loaded []*LoadedUnit
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		unitDir := filepath.Join(v.unitsDir, entry.Name())
		luaPath := filepath.Join(unitDir, "unit.lua")

		if _, err := os.Stat(luaPath); err != nil {
			continue // skip directories without unit.lua
		}

		lu, err := v.LoadUnit(unitDir)
		if err != nil {
			return nil, fmt.Errorf("load unit %s: %w", entry.Name(), err)
		}

		loaded = append(loaded, lu)
	}

	// Sort by ID for deterministic order.
	sort.Slice(loaded, func(i, j int) bool {
		return loaded[i].Unit.ID < loaded[j].Unit.ID
	})

	// Validate cross-unit dependencies.
	if err := validateDeps(loaded); err != nil {
		return nil, err
	}

	return loaded, nil
}

// Close closes all Lua states.
func CloseAll(loaded []*LoadedUnit) {
	for _, lu := range loaded {
		lu.State.Close()
	}
}

func validateDeps(loaded []*LoadedUnit) error {
	knownIDs := make(map[string]bool, len(loaded))
	for _, lu := range loaded {
		knownIDs[lu.Unit.ID] = true
	}

	for _, lu := range loaded {
		for _, stage := range lu.Unit.Stages {
			for _, task := range stage.Tasks {
				for _, dep := range task.DependsOn {
					if strings.Contains(dep, ".") || strings.Contains(dep, "/") {
						continue
					}
					if !knownIDs[dep] {
						return fmt.Errorf("unit %q: unknown dependency %q", lu.Unit.ID, dep)
					}
				}
			}
		}
	}

	return nil
}
