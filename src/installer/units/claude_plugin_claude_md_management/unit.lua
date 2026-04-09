unit {
    id = "claude_plugin_claude_md_management",
    name = "Claude Plugin Claude MD Management",
    depends_on = { "claude_code" },
}

task "verify" {
    run = function()
        return shell_ok("claude plugin list --json 2>/dev/null | grep -q 'claude-md-management@claude-plugins-official'")
    end,
}

task "setup" {
    run = function()
        shell("claude plugin install --scope user claude-md-management@claude-plugins-official")
    end,
}

task "teardown" {
    run = function()
        if shell_ok("claude plugin list --json 2>/dev/null | grep -q 'claude-md-management@claude-plugins-official'") then
            shell("claude plugin uninstall --scope user claude-md-management@claude-plugins-official")
        end
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
