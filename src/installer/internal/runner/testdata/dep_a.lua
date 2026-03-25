unit {
    id = "dep_a",
    name = "Dependency A",
}

task "verify" {
    run = function()
        return false
    end,
}

task "setup" {
    run = function()
        log("dep_a setup")
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
