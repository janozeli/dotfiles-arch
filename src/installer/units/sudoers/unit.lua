unit {
    id = "sudoers",
    name = "NOPASSWD sudoers",
    critical = true,
}

task "verify" {
    run = function()
        return shell_ok("sudo -n true 2>/dev/null")
    end,
}

task "setup" {
    run = function()
        local user = env("USER")
        shell("echo '" .. user .. " ALL=(ALL) NOPASSWD: ALL' | sudo tee /etc/sudoers.d/" .. user .. " >/dev/null")
        shell("sudo chmod 440 /etc/sudoers.d/" .. user)
    end,
}

task "teardown" {
    run = function()
        local user = env("USER")
        shell("sudo rm -f /etc/sudoers.d/" .. user)
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
