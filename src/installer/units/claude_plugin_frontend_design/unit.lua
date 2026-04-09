unit {
    id = "claude_plugin_frontend_design",
    name = "Claude Plugin Frontend Design",
    depends_on = { "claude_code" },
}

task "verify" {
    run = function()
        return shell_ok("claude plugin list --json 2>/dev/null | grep -q 'frontend-design@claude-plugins-official'")
    end,
}

task "setup" {
    run = function()
        shell("claude plugin install --scope user frontend-design@claude-plugins-official")
    end,
}

task "teardown" {
    run = function()
        if shell_ok("claude plugin list --json 2>/dev/null | grep -q 'frontend-design@claude-plugins-official'") then
            shell("claude plugin uninstall --scope user frontend-design@claude-plugins-official")
        end
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
