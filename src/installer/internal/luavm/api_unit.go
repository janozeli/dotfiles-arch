package luavm

import (
	"dotfiles/installer/units"

	lua "github.com/yuin/gopher-lua"
)

// registerUnitAPI registers the unit declaration functions into the Lua state.
func registerUnitAPI(L *lua.LState, builder *unitBuilder) {
	L.SetGlobal("unit", L.NewFunction(luaUnitDecl(builder)))
	L.SetGlobal("task", L.NewFunction(luaTaskDecl(L, builder)))
	L.SetGlobal("stages", L.NewFunction(luaStagesDecl(builder)))
	L.SetGlobal("actions", L.NewFunction(luaActionsDecl(builder)))
}

// unit { id = "x", name = "y", critical = true, depends_on = {"a"} }
func luaUnitDecl(builder *unitBuilder) lua.LGFunction {
	return func(L *lua.LState) int {
		tbl := L.CheckTable(1)

		if id := tbl.RawGetString("id"); id != lua.LNil {
			builder.unit.ID = id.String()
		}
		if name := tbl.RawGetString("name"); name != lua.LNil {
			builder.unit.Name = name.String()
		}
		if crit := tbl.RawGetString("critical"); crit == lua.LTrue {
			builder.unit.Critical = true
		}
		if deps := tbl.RawGetString("depends_on"); deps != lua.LNil {
			if depTbl, ok := deps.(*lua.LTable); ok {
				builder.dependsOn = tableToStringSlice(depTbl)
			}
		}

		return 0
	}
}

// task "name" { input = {...}, output = {...}, timeout = N, run = function() ... end }
// This is curried: task("name") returns a function that accepts the config table.
func luaTaskDecl(L *lua.LState, builder *unitBuilder) lua.LGFunction {
	return func(L *lua.LState) int {
		name := L.CheckString(1)

		// Return a closure that accepts the config table.
		L.Push(L.NewFunction(func(L *lua.LState) int {
			tbl := L.CheckTable(1)

			tf := &TaskFunc{Name: name}

			// Extract run function.
			if fn := tbl.RawGetString("run"); fn != lua.LNil {
				if lfn, ok := fn.(*lua.LFunction); ok {
					tf.Fn = lfn
				}
			}

			// Extract input signature.
			if input := tbl.RawGetString("input"); input != lua.LNil {
				if inputTbl, ok := input.(*lua.LTable); ok {
					tf.Input = tableToStringMap(inputTbl)
				}
			}

			// Extract output signature.
			if output := tbl.RawGetString("output"); output != lua.LNil {
				if outputTbl, ok := output.(*lua.LTable); ok {
					tf.Output = tableToStringMap(outputTbl)
				}
			}

			// Extract timeout.
			if timeout := tbl.RawGetString("timeout"); timeout != lua.LNil {
				if num, ok := timeout.(lua.LNumber); ok {
					tf.Timeout = int(num)
				}
			}

			builder.tasks[name] = tf
			return 0
		}))

		return 1
	}
}

// stages { "verify", "setup" }
func luaStagesDecl(builder *unitBuilder) lua.LGFunction {
	return func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		builder.stageOrder = tableToStringSlice(tbl)
		return 0
	}
}

// actions { teardown = "teardown", list = "list" }
func luaActionsDecl(builder *unitBuilder) lua.LGFunction {
	return func(L *lua.LState) int {
		tbl := L.CheckTable(1)
		tbl.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				if v, ok := value.(lua.LString); ok {
					builder.actions[string(k)] = units.Action{
						Task: string(v),
					}
				}
			}
		})
		return 0
	}
}
