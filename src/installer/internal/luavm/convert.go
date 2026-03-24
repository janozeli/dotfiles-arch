package luavm

import (
	"fmt"

	lua "github.com/yuin/gopher-lua"
)

// tableToStringSlice converts a Lua table (array) to a Go string slice.
func tableToStringSlice(t *lua.LTable) []string {
	var result []string
	t.ForEach(func(_, value lua.LValue) {
		if str, ok := value.(lua.LString); ok {
			result = append(result, string(str))
		}
	})
	return result
}

// tableToStringMap converts a Lua table (map) to a Go string map.
func tableToStringMap(t *lua.LTable) map[string]string {
	result := make(map[string]string)
	t.ForEach(func(key, value lua.LValue) {
		if k, ok := key.(lua.LString); ok {
			if v, ok := value.(lua.LString); ok {
				result[string(k)] = string(v)
			}
		}
	})
	return result
}

// stringSliceToTable converts a Go string slice to a Lua table (array).
func stringSliceToTable(L *lua.LState, s []string) *lua.LTable {
	t := L.NewTable()
	for _, v := range s {
		t.Append(lua.LString(v))
	}
	return t
}

// goValueToLua converts a Go value to a Lua value.
func goValueToLua(L *lua.LState, v any) lua.LValue {
	switch val := v.(type) {
	case string:
		return lua.LString(val)
	case bool:
		return lua.LBool(val)
	case int:
		return lua.LNumber(float64(val))
	case float64:
		return lua.LNumber(val)
	case []string:
		return stringSliceToTable(L, val)
	case []any:
		t := L.NewTable()
		for _, item := range val {
			t.Append(goValueToLua(L, item))
		}
		return t
	case map[string]any:
		t := L.NewTable()
		for k, item := range val {
			t.RawSetString(k, goValueToLua(L, item))
		}
		return t
	case nil:
		return lua.LNil
	default:
		return lua.LString(fmt.Sprintf("%v", val))
	}
}

// luaValueToGo converts a Lua value to a Go value.
func luaValueToGo(v lua.LValue) any {
	switch val := v.(type) {
	case *lua.LNilType:
		return nil
	case lua.LBool:
		return bool(val)
	case lua.LNumber:
		return float64(val)
	case lua.LString:
		return string(val)
	case *lua.LTable:
		// Check if it's an array (sequential integer keys) or a map.
		maxN := val.MaxN()
		if maxN > 0 {
			result := make([]any, 0, maxN)
			for i := 1; i <= maxN; i++ {
				result = append(result, luaValueToGo(val.RawGetInt(i)))
			}
			return result
		}
		result := make(map[string]any)
		val.ForEach(func(key, value lua.LValue) {
			if k, ok := key.(lua.LString); ok {
				result[string(k)] = luaValueToGo(value)
			}
		})
		return result
	default:
		return nil
	}
}
