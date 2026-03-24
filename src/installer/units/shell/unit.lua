unit {
    id = "shell",
    name = "Shell (zsh)",
    depends_on = { "packages" },
}

task "verify" {
    run = function()
        local user = env("USER")
        return shell_ok("getent passwd " .. user .. " | grep -q '/zsh$'")
    end,
}

task "setup" {
    run = function()
        local user = env("USER")
        local zsh = shell("which zsh")
        shell("sudo chsh -s " .. zsh .. " " .. user)
    end,
}

task "teardown" {
    run = function()
        local user = env("USER")
        shell("sudo chsh -s /bin/bash " .. user)
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
