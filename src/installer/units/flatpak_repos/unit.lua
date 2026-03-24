unit {
    id = "flatpak_repos",
    name = "Flatpak repos (Flathub)",
    depends_on = { "packages" },
}

task "verify" {
    run = function()
        return shell_ok("command -v flatpak >/dev/null 2>&1")
            and shell_ok("flatpak remote-list | grep -q flathub")
    end,
}

task "setup" {
    run = function()
        shell("flatpak remote-add --if-not-exists flathub https://dl.flathub.org/repo/flathub.flatpakrepo")
    end,
}

task "teardown" {
    run = function()
        shell("flatpak remote-delete flathub")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
