unit {
    id = "git_configs",
    name = "Git workspace configs",
    depends_on = { "directories", "packages" },
}

task "verify" {
    run = function()
        local home = env("HOME")
        local dotfiles_root = shell("git rev-parse --show-toplevel")

        return shell_ok("test \"$(git config --global user.email)\" = 'lucasmjanozeli@gmail.com'")
            and shell_ok("test \"$(git config --global user.name)\" = 'janozeli'")
            and shell_ok("test -f " .. home .. "/workspace/github.com/janozeli/.gitconfig")
            and shell_ok("test -f " .. home .. "/workspace/github.com/hetosoft/.gitconfig")
            and shell_ok("test \"$(git -C " .. dotfiles_root .. " config user.name)\" = 'janozeli'")
    end,
}

task "setup" {
    run = function()
        local home = env("HOME")
        local dotfiles_root = shell("git rev-parse --show-toplevel")

        shell("git config --global user.email 'lucasmjanozeli@gmail.com'")
        shell("git config --global user.name 'janozeli'")

        write_file(home .. "/workspace/github.com/janozeli/.gitconfig",
            "[user]\n\temail = lucasmjanozeli@gmail.com\n\tname = janozeli\n")

        write_file(home .. "/workspace/github.com/hetosoft/.gitconfig",
            "[user]\n\temail = lucas.janozeli@hetosoft.com.br\n\tname = LucasJanozeli-Hetosoft\n")

        shell("git -C " .. dotfiles_root .. " config user.name 'janozeli'")
        shell("git -C " .. dotfiles_root .. " config user.email 'lucasmjanozeli@gmail.com'")
    end,
}

task "teardown" {
    run = function()
        log("teardown not implemented")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
