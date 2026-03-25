unit {
    id = "contract_test",
    name = "Contract Test",
}

task "verify" {
    output = { value = "string" },
    run = function()
        return false, { value = "wired_data" }
    end,
}

task "setup" {
    input = { value = "string" },
    run = function(input)
        if input.value ~= "wired_data" then
            error("wiring failed: got " .. tostring(input.value))
        end
    end,
}

task "teardown" {
    run = function() end,
}

stages { "verify", "setup" }
actions { teardown = "teardown" }
