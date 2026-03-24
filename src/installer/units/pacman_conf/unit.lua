unit {
    id = "pacman_conf",
    name = "pacman.conf",
    critical = true,
    depends_on = { "sudoers" },
}

task "verify" {
    run = function()
        return shell_ok("grep -q '^Color' /etc/pacman.conf")
            and shell_ok("grep -q '^ParallelDownloads' /etc/pacman.conf")
            and shell_ok("grep -q '^ILoveCandy' /etc/pacman.conf")
    end,
}

task "setup" {
    run = function()
        shell("sudo sed -i 's/^#\\?Color.*/Color/' /etc/pacman.conf")
        shell("sudo sed -i 's/^#\\?ParallelDownloads.*/ParallelDownloads = 15/' /etc/pacman.conf")
        if not shell_ok("grep -q '^ILoveCandy' /etc/pacman.conf") then
            shell("sudo sed -i '/^ParallelDownloads/a ILoveCandy' /etc/pacman.conf")
        end
    end,
}

task "teardown" {
    run = function()
        shell("sudo sed -i 's/^Color/#Color/' /etc/pacman.conf")
        shell("sudo sed -i 's/^ParallelDownloads/#ParallelDownloads/' /etc/pacman.conf")
        shell("sudo sed -i '/^ILoveCandy/d' /etc/pacman.conf")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
