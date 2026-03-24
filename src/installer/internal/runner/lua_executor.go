package runner

import (
	"context"
	"fmt"
	"sync"
	"time"

	"dotfiles/installer/internal/luavm"
	"dotfiles/installer/units"

	lua "github.com/yuin/gopher-lua"
)

// Executor defines how tasks are executed.
type Executor interface {
	Execute(task units.Task, plan units.ExecutionPlan, ctx context.Context) units.TaskResult
}

// LuaExecutor runs tasks by calling Lua functions from loaded units.
type LuaExecutor struct {
	loadedUnits map[string]*luavm.LoadedUnit
	mutexes     map[string]*sync.Mutex
}

// NewLuaExecutor creates an executor wired to the given loaded units.
func NewLuaExecutor(loaded []*luavm.LoadedUnit) *LuaExecutor {
	lu := make(map[string]*luavm.LoadedUnit, len(loaded))
	mu := make(map[string]*sync.Mutex, len(loaded))
	for _, l := range loaded {
		lu[l.Unit.ID] = l
		mu[l.Unit.ID] = &sync.Mutex{}
	}
	return &LuaExecutor{loadedUnits: lu, mutexes: mu}
}

// Execute calls the Lua function for the given task.
func (e *LuaExecutor) Execute(task units.Task, plan units.ExecutionPlan, ctx context.Context) units.TaskResult {
	startedAt := NowISO()
	t0 := time.Now()

	loaded, ok := e.loadedUnits[task.UnitID]
	if !ok {
		return errorResult(task.Name, startedAt, 0, fmt.Sprintf("unit %q not found", task.UnitID))
	}

	tf, ok := loaded.TaskFuncs[task.Name]
	if !ok {
		return errorResult(task.Name, startedAt, 0, fmt.Sprintf("task %q not found in unit %q", task.Name, task.UnitID))
	}

	if tf.Fn == nil {
		return errorResult(task.Name, startedAt, 0, fmt.Sprintf("task %q has no run function", task.Name))
	}

	// Apply timeout.
	timeout := plan.Timeout
	if tf.Timeout > 0 {
		timeout = tf.Timeout
	}
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()
	}

	// Set context on Lua state for cancellation support.
	L := loaded.State
	L.SetContext(ctx)
	defer L.SetContext(nil)

	// Lock per-unit mutex (LState is not goroutine-safe).
	e.mutexes[task.UnitID].Lock()
	defer e.mutexes[task.UnitID].Unlock()

	// Build input table from context if task declares inputs.
	var args []lua.LValue
	if len(tf.Input) > 0 {
		inputTable := L.NewTable()
		for name := range tf.Input {
			val, ok := loaded.Context.GetData(name)
			if ok {
				inputTable.RawSetString(name, luavm.GoValueToLua(L, val))
			}
		}
		args = append(args, inputTable)
	}

	// Call the Lua function with NRet=2 (bool/nil + outputs table).
	err := L.CallByParam(lua.P{
		Fn:      tf.Fn,
		NRet:    2,
		Protect: true,
	}, args...)

	elapsed := int(time.Since(t0).Seconds())

	if err != nil {
		exitCode := 1
		return units.TaskResult{
			ID:        task.Name,
			Status:    units.StatusError,
			Message:   err.Error(),
			ExitCode:  &exitCode,
			StartedAt: startedAt,
			Duration:  elapsed,
		}
	}

	// Capture outputs (second return value) and store in context.
	outputVal := L.Get(-1)
	retVal := L.Get(-2)
	L.Pop(2)

	if len(tf.Output) > 0 {
		if outputTbl, ok := outputVal.(*lua.LTable); ok {
			for name := range tf.Output {
				val := outputTbl.RawGetString(name)
				if val != lua.LNil {
					loaded.Context.SetData(name, luavm.LuaValueToGo(val))
				}
			}
		}
	}

	// Check first return value: true = success, false/nil = needs work.
	exitCode := 0
	if retVal == lua.LTrue {
		return units.TaskResult{
			ID:        task.Name,
			Status:    units.StatusSuccess,
			ExitCode:  &exitCode,
			StartedAt: startedAt,
			Duration:  elapsed,
		}
	}

	if retVal == lua.LFalse {
		exitCode = 1
		return units.TaskResult{
			ID:        task.Name,
			Status:    units.StatusError,
			Message:   "needs setup",
			ExitCode:  &exitCode,
			StartedAt: startedAt,
			Duration:  elapsed,
		}
	}

	// nil or no return = success (setup/teardown tasks don't return bool).
	return units.TaskResult{
		ID:        task.Name,
		Status:    units.StatusSuccess,
		ExitCode:  &exitCode,
		StartedAt: startedAt,
		Duration:  elapsed,
	}
}

func errorResult(id, startedAt string, elapsed int, msg string) units.TaskResult {
	exitCode := 1
	return units.TaskResult{
		ID:        id,
		Status:    units.StatusError,
		Message:   msg,
		ExitCode:  &exitCode,
		StartedAt: startedAt,
		Duration:  elapsed,
	}
}
