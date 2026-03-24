unit {
    id = "flatpak_apps",
    name = "Flatpak apps",
    depends_on = { "flatpak_repos" },
}

task "verify" {
    run = function()
        return shell_ok("flatpak info com.spotify.Client >/dev/null 2>&1")
    end,
}

task "setup" {
    run = function()
        shell("flatpak install -y flathub com.spotify.Client")
    end,
}

task "teardown" {
    run = function()
        shell("flatpak uninstall -y com.spotify.Client")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
