unit {
    id = "claude_mcp_context7",
    name = "Claude MCP Context7",
    depends_on = { "claude_code", "mise" },
}

task "verify" {
    run = function()
        if env("CONTEXT7_API_KEY") == "" then
            return false
        end
        return shell_ok("claude mcp get context7 >/dev/null 2>&1")
    end,
}

task "setup" {
    run = function()
        if env("CONTEXT7_API_KEY") == "" then
            return
        end

        shell("claude mcp add --transport http --scope user "
            .. "context7 https://mcp.context7.com/mcp "
            .. "--header 'CONTEXT7_API_KEY: ${CONTEXT7_API_KEY}'")
    end,
}

task "teardown" {
    run = function()
        shell("claude mcp remove --scope user context7")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
