unit {
    id = "test_contract",
    name = "Contract Test Unit",
}

task "verify" {
    output = { result = "string" },
    run = function()
        return true, { result = "hello" }
    end,
}

task "setup" {
    input = { result = "string" },
    run = function(input)
        return input.result == "hello"
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
