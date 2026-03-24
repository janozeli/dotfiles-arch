unit {
    id = "directories",
    name = "Diretórios",
    critical = true,
}

task "verify" {
    run = function()
        local home = env("HOME")
        return shell_ok("test -d " .. home .. "/.local/bin")
            and shell_ok("test -d " .. home .. "/.config")
            and shell_ok("test -d " .. home .. "/workspace/github.com/janozeli")
            and shell_ok("test -d " .. home .. "/workspace/github.com/hetosoft")
    end,
}

task "setup" {
    run = function()
        local home = env("HOME")
        shell("mkdir -p " .. home .. "/.local/bin")
        shell("mkdir -p " .. home .. "/.config")
        shell("mkdir -p " .. home .. "/workspace/github.com/janozeli")
        shell("mkdir -p " .. home .. "/workspace/github.com/hetosoft")
    end,
}

task "teardown" {
    run = function()
        log("teardown not implemented")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
