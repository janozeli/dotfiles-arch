package luavm

import (
	"fmt"
	"os"

	lua "github.com/yuin/gopher-lua"
)

// registerUtilAPI registers utility functions into the Lua state.
func registerUtilAPI(L *lua.LState) {
	L.SetGlobal("diff", L.NewFunction(luaDiff))
	L.SetGlobal("contains", L.NewFunction(luaContains))
	L.SetGlobal("log", L.NewFunction(luaLog))
	L.SetGlobal("env", L.NewFunction(luaEnv))
}

// diff(list1, list2) — items in list1 not in list2.
func luaDiff(L *lua.LState) int {
	t1 := L.CheckTable(1)
	t2 := L.CheckTable(2)

	set := make(map[string]bool)
	t2.ForEach(func(_, v lua.LValue) {
		if s, ok := v.(lua.LString); ok {
			set[string(s)] = true
		}
	})

	result := L.NewTable()
	t1.ForEach(func(_, v lua.LValue) {
		if s, ok := v.(lua.LString); ok {
			if !set[string(s)] {
				result.Append(v)
			}
		}
	})

	L.Push(result)
	return 1
}

// contains(list, item) — check if item is in list.
func luaContains(L *lua.LState) int {
	t := L.CheckTable(1)
	item := L.CheckString(2)

	found := false
	t.ForEach(func(_, v lua.LValue) {
		if s, ok := v.(lua.LString); ok && string(s) == item {
			found = true
		}
	})

	L.Push(lua.LBool(found))
	return 1
}

// log(msg) — print to stderr.
func luaLog(L *lua.LState) int {
	msg := L.CheckString(1)
	fmt.Fprintln(os.Stderr, msg)
	return 0
}

// env(name) — safe os.Getenv without exposing os.*.
func luaEnv(L *lua.LState) int {
	name := L.CheckString(1)
	L.Push(lua.LString(os.Getenv(name)))
	return 1
}
