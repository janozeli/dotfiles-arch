unit {
    id = "dep_b",
    name = "Dependency B",
    depends_on = { "dep_a" },
}

task "verify" {
    run = function()
        return false
    end,
}

task "setup" {
    run = function()
        log("dep_b setup")
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
