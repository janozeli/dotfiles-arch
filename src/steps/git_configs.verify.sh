#!/usr/bin/env bash
[ "$(git config --global user.email)" = "lucasmjanozeli@gmail.com" ] \
    && [ "$(git config --global user.name)" = "janozeli" ] \
    && [ -f "$HOME/workspace/github.com/janozeli/.gitconfig" ] \
    && [ -f "$HOME/workspace/github.com/hetosoft/.gitconfig" ]
