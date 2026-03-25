unit {
    id = "test_bad_contract",
    name = "Bad Contract",
}

task "verify" {
    run = function()
        return true
    end,
}

task "setup" {
    input = { missing_data = "string" },
    run = function(input) end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
