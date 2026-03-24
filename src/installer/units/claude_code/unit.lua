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
        log("teardown not implemented")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
