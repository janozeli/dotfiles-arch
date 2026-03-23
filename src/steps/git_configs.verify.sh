#!/usr/bin/env bash
dotfiles_root="$(git rev-parse --show-toplevel)"
[ "$(git config --global user.email)" = "lucasmjanozeli@gmail.com" ] \
    && [ "$(git config --global user.name)" = "janozeli" ] \
    && [ -f "$HOME/workspace/github.com/janozeli/.gitconfig" ] \
    && [ -f "$HOME/workspace/github.com/hetosoft/.gitconfig" ] \
    && [ "$(git -C "$dotfiles_root" config user.name)" = "janozeli" ]
