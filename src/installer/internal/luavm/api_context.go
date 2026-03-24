package luavm

import (
	lua "github.com/yuin/gopher-lua"
)

// contextStore holds inter-task data for a single unit.
type contextStore struct {
	data map[string]any
}

func newContextStore() *contextStore {
	return &contextStore{data: make(map[string]any)}
}

// registerContextAPI registers the context table into the Lua state.
func registerContextAPI(L *lua.LState, ctx *contextStore) {
	t := L.NewTable()

	L.SetField(t, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.Get(2)
		ctx.data[key] = luaValueToGo(value)
		return 0
	}))

	L.SetField(t, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val, ok := ctx.data[key]
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(goValueToLua(L, val))
		return 1
	}))

	L.SetGlobal("context", t)
}
