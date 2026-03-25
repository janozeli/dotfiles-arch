unit {
    name = "Missing ID",
}

task "verify" {
    run = function() return true end,
}

task "setup" {
    run = function() end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
