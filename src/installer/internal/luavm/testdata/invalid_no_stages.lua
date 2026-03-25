unit {
    id = "test_no_stages",
    name = "No Stages",
}

task "verify" {
    run = function() return true end,
}

task "setup" {
    run = function() end,
}
