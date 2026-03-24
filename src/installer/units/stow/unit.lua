unit {
    id = "stow",
    name = "GNU Stow symlinks",
    depends_on = { "packages" },
}

task "verify" {
    run = function()
        local home = env("HOME")
        return shell_ok("test -L " .. home .. "/.zshrc")
    end,
}

task "setup" {
    run = function()
        local root = shell("git rev-parse --show-toplevel")
        shell("cd " .. root .. " && stow --dotfiles --no-folding -t $HOME .")
    end,
}

task "teardown" {
    run = function()
        local root = shell("git rev-parse --show-toplevel")
        shell("cd " .. root .. " && stow -D --dotfiles -t $HOME .")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
