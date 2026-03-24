unit {
    id = "dotenv",
    name = ".env",
}

task "verify" {
    run = function()
        local home = env("HOME")
        return shell_ok("test -f " .. home .. "/.env")
    end,
}

task "setup" {
    run = function()
        local home = env("HOME")
        write_file(home .. "/.env", 'export EXA_API_KEY=""\nexport CONTEXT7_API_KEY=""\n')
    end,
}

task "teardown" {
    run = function()
        local home = env("HOME")
        shell("rm -f " .. home .. "/.env")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
