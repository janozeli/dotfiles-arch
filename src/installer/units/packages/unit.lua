unit {
    id = "packages",
    name = "Pacotes (yay)",
    critical = true,
    depends_on = { "yay" },
}

local packages = {
    "stow", "zsh", "git", "less", "wget", "curl", "unzip",
    "fzf", "zoxide", "eza", "bat", "tldr",
    "kitty", "github-cli", "wl-clipboard", "xclip",
    "mpv", "yt-dlp",
    "zed", "firefox", "zen-browser-bin", "google-chrome",
    "ttf-firacode-nerd", "ttf-jetbrains-mono-nerd",
    "docker", "docker-compose", "teams-for-linux-bin",
    "flatpak",
    "visual-studio-code-bin", "obsidian-bin", "ufw",
    "snap-pac",
}

task "verify" {
    output = { missing = "list" },
    run = function()
        local missing = {}
        for _, pkg in ipairs(packages) do
            if not shell_ok("yay -Q " .. pkg .. " >/dev/null 2>&1") then
                table.insert(missing, pkg)
            end
        end
        return #missing == 0, { missing = missing }
    end,
}

task "setup" {
    input = { missing = "list" },
    timeout = 600,
    run = function(input)
        if #input.missing > 0 then
            shell("yay -S --noconfirm --needed " .. table.concat(input.missing, " "))
        end
    end,
}

task "teardown" {
    run = function()
        shell("yay -Rns --noconfirm " .. table.concat(packages, " "))
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
