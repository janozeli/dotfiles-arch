package luavm

import (
	"os"
	"path/filepath"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// registerFsAPI registers filesystem functions into the Lua state.
// unitDir is the absolute path to the unit's directory.
func registerFsAPI(L *lua.LState, unitDir string) {
	L.SetGlobal("read_file", L.NewFunction(luaReadFile))
	L.SetGlobal("read_lines", L.NewFunction(luaReadLines))
	L.SetGlobal("write_file", L.NewFunction(luaWriteFile))

	// unit_path resolves relative to the unit directory.
	L.SetGlobal("unit_path", L.NewFunction(func(L *lua.LState) int {
		rel := L.CheckString(1)
		L.Push(lua.LString(filepath.Join(unitDir, rel)))
		return 1
	}))
}

// read_file(path) — read entire file as string.
func luaReadFile(L *lua.LState) int {
	path := L.CheckString(1)
	data, err := os.ReadFile(path)
	if err != nil {
		L.RaiseError("read_file(%q): %s", path, err)
		return 0
	}
	L.Push(lua.LString(string(data)))
	return 1
}

// read_lines(path) — read file, return lines as table.
func luaReadLines(L *lua.LState) int {
	path := L.CheckString(1)
	data, err := os.ReadFile(path)
	if err != nil {
		L.RaiseError("read_lines(%q): %s", path, err)
		return 0
	}

	raw := strings.TrimRight(string(data), "\n")
	var lines []string
	if raw != "" {
		lines = strings.Split(raw, "\n")
	}

	L.Push(stringSliceToTable(L, lines))
	return 1
}

// write_file(path, data) — write string to file.
func luaWriteFile(L *lua.LState) int {
	path := L.CheckString(1)
	data := L.CheckString(2)
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		L.RaiseError("write_file(%q): %s", path, err)
	}
	return 0
}
