unit {
    id = "yay",
    name = "yay (AUR helper)",
    critical = true,
    depends_on = { "sudoers" },
}

task "verify" {
    run = function()
        return shell_ok("command -v yay >/dev/null")
    end,
}

task "setup" {
    timeout = 300,
    run = function()
        shell("sudo pacman -S --noconfirm --needed git base-devel")
        shell("rm -rf /tmp/yay")
        shell("git clone https://aur.archlinux.org/yay.git /tmp/yay")
        shell("cd /tmp/yay && makepkg -si --noconfirm")
        shell("rm -rf /tmp/yay")
        shell("yay -Y --gendb")
        shell("yay -Y --devel --save")
    end,
}

task "teardown" {
    run = function()
        shell("sudo pacman -R --noconfirm yay")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
