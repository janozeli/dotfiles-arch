unit {
    id = "claude_mcp_exa",
    name = "Claude MCP Exa",
    depends_on = { "claude_code", "mise" },
}

task "verify" {
    run = function()
        if env("EXA_API_KEY") == "" then
            return false
        end
        return shell_ok("claude mcp get exa >/dev/null 2>&1")
    end,
}

task "setup" {
    run = function()
        if env("EXA_API_KEY") == "" then
            return
        end

        shell("claude mcp add --transport http --scope user exa "
            .. "'https://mcp.exa.ai/mcp?exaApiKey=${EXA_API_KEY}"
            .. "&tools=web_search_exa,web_search_advanced_exa,get_code_context_exa,crawling_exa'")
    end,
}

task "teardown" {
    run = function()
        shell("claude mcp remove --scope user exa")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
