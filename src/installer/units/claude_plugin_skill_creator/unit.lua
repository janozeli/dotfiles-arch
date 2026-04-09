unit {
    id = "claude_plugin_skill_creator",
    name = "Claude Plugin Skill Creator",
    depends_on = { "claude_code" },
}

task "verify" {
    run = function()
        return shell_ok("claude plugin list --json 2>/dev/null | grep -q 'skill-creator@claude-plugins-official'")
    end,
}

task "setup" {
    run = function()
        shell("claude plugin install --scope user skill-creator@claude-plugins-official")
    end,
}

task "teardown" {
    run = function()
        if shell_ok("claude plugin list --json 2>/dev/null | grep -q 'skill-creator@claude-plugins-official'") then
            shell("claude plugin uninstall --scope user skill-creator@claude-plugins-official")
        end
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
