#!/usr/bin/env bash
set -euo pipefail

claude mcp add --scope user exa -- npx -y exa-mcp-server --tools=web_search_exa,get_code_context_exa,web_search_advanced_exa,crawling_exa
claude mcp add --scope user context7 -- npx -y @upstash/context7-mcp
