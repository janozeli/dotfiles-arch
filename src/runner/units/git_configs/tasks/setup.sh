#!/usr/bin/env bash
set -euo pipefail

git config --global user.email "lucasmjanozeli@gmail.com"
git config --global user.name "janozeli"

cat > "$HOME/workspace/github.com/janozeli/.gitconfig" << 'EOF'
[user]
	email = lucasmjanozeli@gmail.com
	name = janozeli
EOF

cat > "$HOME/workspace/github.com/hetosoft/.gitconfig" << 'EOF'
[user]
	email = lucas.janozeli@hetosoft.com.br
	name = LucasJanozeli-Hetosoft
EOF

# Set local identity in dotfiles repo (wherever it lives)
dotfiles_root="$(git rev-parse --show-toplevel)"
git -C "$dotfiles_root" config user.name "janozeli"
git -C "$dotfiles_root" config user.email "lucasmjanozeli@gmail.com"
