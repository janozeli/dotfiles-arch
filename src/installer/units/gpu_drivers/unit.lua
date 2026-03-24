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
            local ok = shell_ok("pacman -Qi nvidia-open-dkms >/dev/null 2>&1")
                and shell_ok("pacman -Qi nvidia-utils >/dev/null 2>&1")
            return ok, { vendor = "nvidia" }
        elseif gpu:find("amd") or gpu:find("radeon") then
            return true, { vendor = "amd" }
        elseif gpu:find("intel") then
            return true, { vendor = "intel" }
        end

        return false, { vendor = "unknown" }
    end,
}

task "setup" {
    input = { vendor = "string" },
    timeout = 300,
    run = function(input)
        if input.vendor == "nvidia" then
            log("NVIDIA detected, installing proprietary drivers...")
            shell("sudo pacman -S --noconfirm --needed nvidia-open-dkms nvidia-utils nvidia-settings")
        elseif input.vendor == "amd" then
            log("AMD detected, open-source drivers (mesa) already covered.")
        elseif input.vendor == "intel" then
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
