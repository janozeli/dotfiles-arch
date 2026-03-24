unit {
    id = "oh_my_posh",
    name = "oh-my-posh",
}

task "verify" {
    run = function()
        return shell_ok("command -v oh-my-posh >/dev/null")
    end,
}

task "setup" {
    run = function()
        local home = env("HOME")
        shell("curl -s https://ohmyposh.dev/install.sh | bash -s -- -d " .. home .. "/.local/bin")
    end,
}

task "teardown" {
    run = function()
        local home = env("HOME")
        shell("rm -f " .. home .. "/.local/bin/oh-my-posh")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
