unit {
    id = "services",
    name = "Serviços",
    depends_on = { "packages" },
}

task "verify" {
    run = function()
        return shell_ok("systemctl is-active --quiet docker")
            and shell_ok("systemctl is-active --quiet fstrim.timer")
            and shell_ok("systemctl is-active --quiet ufw")
            and shell_ok("systemctl --user is-active --quiet spotifyd")
    end,
}

task "setup" {
    run = function()
        local user = env("USER")
        shell("sudo systemctl enable --now fstrim.timer")
        shell("sudo systemctl enable --now docker.service")
        shell("sudo usermod -aG docker " .. user)
        shell("sudo systemctl enable --now ufw.service")
        shell("sudo ufw default deny incoming")
        shell("sudo ufw default allow outgoing")
        shell("sudo ufw enable")
        shell("systemctl --user enable --now spotifyd.service")
    end,
}

task "teardown" {
    run = function()
        shell("sudo systemctl disable --now docker.service")
        shell("sudo systemctl disable --now ufw.service")
        shell("sudo systemctl disable --now fstrim.timer")
        shell("systemctl --user disable --now spotifyd.service")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
