unit {
    id = "test_valid",
    name = "Valid Test Unit",
}

task "verify" {
    run = function()
        return true
    end,
}

task "setup" {
    timeout = 30,
    run = function()
        log("setup ran")
    end,
}

task "teardown" {
    run = function()
        log("teardown ran")
    end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
