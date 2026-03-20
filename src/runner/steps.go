package main

// Steps are registered in execution order.
var (
	_ = register(&BashStep{
		id: "sudoers", name: "NOPASSWD sudoers", script: "sudoers.sh",
		critical: true,
	})
	_ = register(&BashStep{
		id: "yay", name: "yay (AUR helper)", script: "yay.sh",
		critical: true,
	})
	_ = register(&BashStep{
		id: "pacman_conf", name: "pacman.conf", script: "pacman_conf.sh",
		critical: true,
	})
	_ = register(&BashStep{
		id: "packages", name: "Pacotes (yay)", script: "packages.sh",
		critical: true,
	})
	_ = register(&BashStep{
		id: "directories", name: "Diretórios", script: "directories.sh",
		critical: true,
	})
	_ = register(&BashStep{
		id: "mise", name: "mise + Node/uv/pnpm", script: "mise.sh",
	})
	_ = register(&BashStep{
		id: "oh_my_posh", name: "oh-my-posh", script: "oh_my_posh.sh",
	})
	_ = register(&BashStep{
		id: "claude_code", name: "Claude Code", script: "claude_code.sh",
	})
	_ = register(&BashStep{
		id: "git_configs", name: "Git workspace configs", script: "git_configs.sh",
	})
	_ = register(&BashStep{
		id: "dotenv", name: ".env", script: "dotenv.sh",
	})
	_ = register(&BashStep{
		id: "claude_mcp", name: "Claude MCP servers", script: "claude_mcp.sh",
		dependsOn: []string{"claude_code"},
	})
	_ = register(&BashStep{
		id: "locale", name: "Locale/teclado", script: "locale.sh",
		requiresSystemd: true,
	})
	_ = register(&BashStep{
		id: "stow", name: "GNU Stow symlinks", script: "stow.sh",
	})
	_ = register(&BashStep{
		id: "services", name: "Serviços", script: "services.sh",
		requiresSystemd: true,
	})
	_ = register(&BashStep{
		id: "shell", name: "Shell (zsh)", script: "shell.sh",
	})
)
