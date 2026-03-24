unit {
    id = "system_update",
    name = "System update",
    critical = true,
    depends_on = { "yay" },
}

task "verify" {
    run = function()
        -- Always run update; skip only if last update was < 1h ago.
        local ok = shell_ok("test -f /tmp/.dotfiles_updated && " ..
            "test $(($(date +%s) - $(stat -c %Y /tmp/.dotfiles_updated))) -lt 3600")
        return ok
    end,
}

task "setup" {
    timeout = 600,
    run = function()
        shell("yay -Syyuu --noconfirm")
        shell("touch /tmp/.dotfiles_updated")
    end,
}

task "teardown" {
    run = function()
        shell("rm -f /tmp/.dotfiles_updated")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
