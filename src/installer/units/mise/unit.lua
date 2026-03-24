unit {
    id = "mise",
    name = "mise + Node/uv/pnpm",
}

task "verify" {
    run = function()
        return shell_ok("command -v mise >/dev/null")
            and shell_ok("mise which node >/dev/null 2>&1")
            and shell_ok("mise which uv >/dev/null 2>&1")
            and shell_ok("mise which pnpm >/dev/null 2>&1")
    end,
}

task "setup" {
    timeout = 300,
    run = function()
        local home = env("HOME")
        shell("curl https://mise.jdx.dev/install.sh | sh")
        shell(home .. '/.local/bin/mise install node@lts uv@latest pnpm@latest')
        shell(home .. '/.local/bin/mise use --global node@lts uv@latest pnpm@latest')
    end,
}

task "teardown" {
    run = function()
        local home = env("HOME")
        shell("rm -f " .. home .. "/.local/bin/mise")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
