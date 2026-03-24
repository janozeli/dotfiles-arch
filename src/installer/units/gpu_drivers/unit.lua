unit {
    id = "gpu_drivers",
    name = "GPU drivers",
    depends_on = { "packages" },
}

task "verify" {
    output = { vendor = "string" },
    run = function()
        local gpu = shell("lspci -nn | grep -iE 'vga|3d|display'"):lower()

        if gpu:find("nvidia") then
            context.set("vendor", "nvidia")
            return shell_ok("pacman -Qi nvidia-open-dkms >/dev/null 2>&1")
                and shell_ok("pacman -Qi nvidia-utils >/dev/null 2>&1")
        elseif gpu:find("amd") or gpu:find("radeon") then
            context.set("vendor", "amd")
            return true
        elseif gpu:find("intel") then
            context.set("vendor", "intel")
            return true
        end

        return false
    end,
}

task "setup" {
    input = { vendor = "string" },
    timeout = 300,
    run = function()
        local vendor = context.get("vendor")

        if vendor == "nvidia" then
            log("NVIDIA detected, installing proprietary drivers...")
            shell("sudo pacman -S --noconfirm --needed nvidia-open-dkms nvidia-utils nvidia-settings")
        elseif vendor == "amd" then
            log("AMD detected, open-source drivers (mesa) already covered.")
        elseif vendor == "intel" then
            log("Intel detected, open-source drivers (mesa) already covered.")
        else
            error("GPU not identified")
        end
    end,
}

task "teardown" {
    run = function()
        log("teardown not implemented")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
