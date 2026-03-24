unit {
    id = "locale",
    name = "Locale/teclado",
}

task "verify" {
    run = function()
        return shell_ok("grep -q '^LANG=pt_BR' /etc/locale.conf 2>/dev/null")
            and shell_ok("locale -a 2>/dev/null | grep -q 'pt_BR'")
            and shell_ok("localectl status 2>/dev/null | grep -q 'X11 Layout.*us,br'")
    end,
}

task "setup" {
    run = function()
        shell("sudo sed -i 's/^#pt_BR.UTF-8/pt_BR.UTF-8/' /etc/locale.gen")
        shell("sudo locale-gen")
        shell("sudo localectl set-locale LANG=pt_BR.UTF-8")
        shell("sudo localectl set-x11-keymap us,br pc105 altgr-intl,")
    end,
}

task "teardown" {
    run = function()
        log("teardown not implemented")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
