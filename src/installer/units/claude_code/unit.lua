unit {
    id = "claude_code",
    name = "Claude Code",
}

task "verify" {
    run = function()
        return shell_ok("command -v claude >/dev/null")
    end,
}

task "setup" {
    run = function()
        shell("curl -fsSL https://claude.ai/install.sh | bash")
    end,
}

task "teardown" {
    run = function()
        local home = env("HOME")
        shell("rm -rf " .. home .. "/.claude")
        shell("rm -f " .. home .. "/.local/bin/claude")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
