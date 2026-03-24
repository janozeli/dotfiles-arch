package luavm

import (
	"os/exec"
	"strings"

	lua "github.com/yuin/gopher-lua"
)

// registerShellAPI registers shell execution functions into the Lua state.
func registerShellAPI(L *lua.LState) {
	L.SetGlobal("shell", L.NewFunction(luaShell))
	L.SetGlobal("shell_ok", L.NewFunction(luaShellOk))
	L.SetGlobal("shell_lines", L.NewFunction(luaShellLines))
}

// shell(cmd) — execute command, return stdout. Error on non-zero exit.
func luaShell(L *lua.LState) int {
	cmdStr := L.CheckString(1)
	cmd := exec.Command("bash", "-c", cmdStr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		L.RaiseError("shell(%q): %s\n%s", cmdStr, err, string(out))
		return 0
	}
	L.Push(lua.LString(strings.TrimRight(string(out), "\n")))
	return 1
}

// shell_ok(cmd) — execute command, return true if exit 0.
func luaShellOk(L *lua.LState) int {
	cmdStr := L.CheckString(1)
	cmd := exec.Command("bash", "-c", cmdStr)
	err := cmd.Run()
	L.Push(lua.LBool(err == nil))
	return 1
}

// shell_lines(cmd) — execute command, return stdout lines as table.
func luaShellLines(L *lua.LState) int {
	cmdStr := L.CheckString(1)
	cmd := exec.Command("bash", "-c", cmdStr)
	out, err := cmd.Output()
	if err != nil {
		L.RaiseError("shell_lines(%q): %s", cmdStr, err)
		return 0
	}

	raw := strings.TrimRight(string(out), "\n")
	var lines []string
	if raw != "" {
		lines = strings.Split(raw, "\n")
	}

	L.Push(stringSliceToTable(L, lines))
	return 1
}
