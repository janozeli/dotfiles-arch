unit {
    id = "claude_mcp",
    name = "Claude MCP servers",
    depends_on = { "claude_code", "mise" },
}

task "verify" {
    run = function()
        return shell_ok("claude mcp list 2>/dev/null | grep -q 'exa'")
            and shell_ok("claude mcp list 2>/dev/null | grep -q 'context7'")
    end,
}

task "setup" {
    run = function()
        shell("claude mcp add --scope user exa -- npx -y exa-mcp-server --tools=web_search_exa,get_code_context_exa,web_search_advanced_exa,crawling_exa")
        shell("claude mcp add --scope user context7 -- npx -y @upstash/context7-mcp")
    end,
}

task "teardown" {
    run = function()
        shell("claude mcp remove exa")
        shell("claude mcp remove context7")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
