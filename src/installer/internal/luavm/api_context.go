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

// GetData returns the value for a key from the context store.
func (c *contextStore) GetData(key string) (any, bool) {
	val, ok := c.data[key]
	return val, ok
}

// SetData stores a value in the context store.
func (c *contextStore) SetData(key string, value any) {
	c.data[key] = value
}

// registerContextAPI registers the context table into the Lua state.
func registerContextAPI(L *lua.LState, ctx *contextStore) {
	t := L.NewTable()

	L.SetField(t, "set", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		value := L.Get(2)
		ctx.data[key] = LuaValueToGo(value)
		return 0
	}))

	L.SetField(t, "get", L.NewFunction(func(L *lua.LState) int {
		key := L.CheckString(1)
		val, ok := ctx.data[key]
		if !ok {
			L.Push(lua.LNil)
			return 1
		}
		L.Push(GoValueToLua(L, val))
		return 1
	}))

	L.SetGlobal("context", t)
}
