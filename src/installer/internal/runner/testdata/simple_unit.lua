unit {
    id = "simple",
    name = "Simple Unit",
}

task "verify" {
    run = function()
        return true
    end,
}

task "setup" {
    run = function()
        log("setup ran")
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
