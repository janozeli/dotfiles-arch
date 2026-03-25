package luavm

import (
	"testing"

	lua "github.com/yuin/gopher-lua"
)

func newTestState() *lua.LState {
	return lua.NewState(lua.Options{SkipOpenLibs: true})
}

func TestGoValueToLua_String(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, "hello")
	if s, ok := v.(lua.LString); !ok || string(s) != "hello" {
		t.Fatalf("expected LString 'hello', got %v", v)
	}
}

func TestGoValueToLua_Bool(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, true)
	if b, ok := v.(lua.LBool); !ok || !bool(b) {
		t.Fatalf("expected LBool true, got %v", v)
	}
}

func TestGoValueToLua_Number(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, 42.5)
	if n, ok := v.(lua.LNumber); !ok || float64(n) != 42.5 {
		t.Fatalf("expected LNumber 42.5, got %v", v)
	}
}

func TestGoValueToLua_Int(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, 10)
	if n, ok := v.(lua.LNumber); !ok || float64(n) != 10 {
		t.Fatalf("expected LNumber 10, got %v", v)
	}
}

func TestGoValueToLua_Nil(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, nil)
	if v != lua.LNil {
		t.Fatalf("expected LNil, got %v", v)
	}
}

func TestGoValueToLua_StringSlice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, []string{"a", "b", "c"})
	tbl, ok := v.(*lua.LTable)
	if !ok {
		t.Fatalf("expected LTable, got %T", v)
	}
	if tbl.MaxN() != 3 {
		t.Fatalf("expected 3 elements, got %d", tbl.MaxN())
	}
	if tbl.RawGetInt(1).String() != "a" {
		t.Fatalf("expected 'a' at index 1, got %s", tbl.RawGetInt(1))
	}
}

func TestGoValueToLua_Map(t *testing.T) {
	L := newTestState()
	defer L.Close()

	v := GoValueToLua(L, map[string]any{"key": "val", "num": 1.0})
	tbl, ok := v.(*lua.LTable)
	if !ok {
		t.Fatalf("expected LTable, got %T", v)
	}
	if tbl.RawGetString("key").String() != "val" {
		t.Fatalf("expected 'val', got %s", tbl.RawGetString("key"))
	}
}

func TestLuaValueToGo_String(t *testing.T) {
	v := LuaValueToGo(lua.LString("hello"))
	if s, ok := v.(string); !ok || s != "hello" {
		t.Fatalf("expected 'hello', got %v", v)
	}
}

func TestLuaValueToGo_Bool(t *testing.T) {
	v := LuaValueToGo(lua.LBool(true))
	if b, ok := v.(bool); !ok || !b {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestLuaValueToGo_Number(t *testing.T) {
	v := LuaValueToGo(lua.LNumber(3.14))
	if n, ok := v.(float64); !ok || n != 3.14 {
		t.Fatalf("expected 3.14, got %v", v)
	}
}

func TestLuaValueToGo_Nil(t *testing.T) {
	v := LuaValueToGo(lua.LNil)
	if v != nil {
		t.Fatalf("expected nil, got %v", v)
	}
}

func TestLuaValueToGo_Array(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.Append(lua.LString("x"))
	tbl.Append(lua.LString("y"))

	v := LuaValueToGo(tbl)
	arr, ok := v.([]any)
	if !ok {
		t.Fatalf("expected []any, got %T", v)
	}
	if len(arr) != 2 || arr[0] != "x" || arr[1] != "y" {
		t.Fatalf("expected [x y], got %v", arr)
	}
}

func TestLuaValueToGo_Map(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("a", lua.LNumber(1))
	tbl.RawSetString("b", lua.LString("two"))

	v := LuaValueToGo(tbl)
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", v)
	}
	if m["a"] != float64(1) || m["b"] != "two" {
		t.Fatalf("expected {a:1 b:two}, got %v", m)
	}
}

func TestRoundtrip_String(t *testing.T) {
	L := newTestState()
	defer L.Close()

	original := "roundtrip"
	result := LuaValueToGo(GoValueToLua(L, original))
	if result != original {
		t.Fatalf("roundtrip failed: %v != %v", result, original)
	}
}

func TestRoundtrip_Bool(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := LuaValueToGo(GoValueToLua(L, true))
	if result != true {
		t.Fatalf("roundtrip failed: %v", result)
	}
}

func TestRoundtrip_Number(t *testing.T) {
	L := newTestState()
	defer L.Close()

	result := LuaValueToGo(GoValueToLua(L, 99.9))
	if result != 99.9 {
		t.Fatalf("roundtrip failed: %v", result)
	}
}

func TestRoundtrip_List(t *testing.T) {
	L := newTestState()
	defer L.Close()

	original := []any{"a", "b", "c"}
	result := LuaValueToGo(GoValueToLua(L, original))
	arr, ok := result.([]any)
	if !ok || len(arr) != 3 {
		t.Fatalf("roundtrip failed: %v", result)
	}
	for i, v := range original {
		if arr[i] != v {
			t.Fatalf("roundtrip index %d: %v != %v", i, arr[i], v)
		}
	}
}

func TestTableToStringSlice(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.Append(lua.LString("one"))
	tbl.Append(lua.LString("two"))

	s := tableToStringSlice(tbl)
	if len(s) != 2 || s[0] != "one" || s[1] != "two" {
		t.Fatalf("expected [one two], got %v", s)
	}
}

func TestTableToStringMap(t *testing.T) {
	L := newTestState()
	defer L.Close()

	tbl := L.NewTable()
	tbl.RawSetString("k", lua.LString("v"))

	m := tableToStringMap(tbl)
	if m["k"] != "v" {
		t.Fatalf("expected {k:v}, got %v", m)
	}
}
