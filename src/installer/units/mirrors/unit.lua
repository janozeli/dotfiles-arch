unit {
    id = "mirrors",
    name = "Pacman mirrorlist",
    critical = true,
    depends_on = { "packages" },
}

task "verify" {
    run = function()
        if not shell_ok("command -v reflector >/dev/null") then
            return false
        end
        -- Skip if mirrorlist was updated < 24h ago.
        return shell_ok("test -f /etc/pacman.d/mirrorlist && " ..
            "test $(($(date +%s) - $(stat -c %Y /etc/pacman.d/mirrorlist))) -lt 86400")
    end,
}

task "setup" {
    timeout = 120,
    run = function()
        shell("sudo reflector --protocol https --sort rate --latest 10 --save /etc/pacman.d/mirrorlist")
    end,
}

task "teardown" {
    run = function()
        -- Regenerate default mirrorlist from pacman-mirrorlist package.
        shell("sudo cp /etc/pacman.d/mirrorlist.pacnew /etc/pacman.d/mirrorlist 2>/dev/null || true")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
