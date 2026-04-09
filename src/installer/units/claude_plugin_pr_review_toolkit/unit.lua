unit {
    id = "claude_plugin_pr_review_toolkit",
    name = "Claude Plugin PR Review Toolkit",
    depends_on = { "claude_code" },
}

task "verify" {
    run = function()
        return shell_ok("claude plugin list --json 2>/dev/null | grep -q 'pr-review-toolkit@claude-plugins-official'")
    end,
}

task "setup" {
    run = function()
        shell("claude plugin install --scope user pr-review-toolkit@claude-plugins-official")
    end,
}

task "teardown" {
    run = function()
        if shell_ok("claude plugin list --json 2>/dev/null | grep -q 'pr-review-toolkit@claude-plugins-official'") then
            shell("claude plugin uninstall --scope user pr-review-toolkit@claude-plugins-official")
        end
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
