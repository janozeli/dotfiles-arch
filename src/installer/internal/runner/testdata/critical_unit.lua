unit {
    id = "critical_test",
    name = "Critical Test",
    critical = true,
}

task "verify" {
    run = function()
        return false
    end,
}

task "setup" {
    run = function()
        error("critical failure")
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
