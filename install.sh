#!/usr/bin/env bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# ── yay ──────────────────────────────────────────────────────────────
if ! command -v yay &>/dev/null; then
    print_info "Instalando yay..."
    sudo pacman -S --needed git base-devel
    git clone https://aur.archlinux.org/yay.git /tmp/yay
    cd /tmp/yay
    makepkg -si
    cd -
    rm -rf /tmp/yay
    yay -Y --gendb
    yay -Syu --devel
    yay -Y --devel --save
    print_success "yay instalado!"
else
    print_info "yay já instalado, pulando..."
fi

# ── Pacotes via yay/pacman ───────────────────────────────────────────
print_info "Instalando pacotes via yay..."
yay -S --noconfirm --needed \
    stow zsh git less wget curl unzip \
    fzf zoxide eza bat tldr \
    ghostty gh wl-clipboard xclip xsel \
    mpv yt-dlp \
    zed firefox zen-browser-bin \
    ttf-firacode-nerd ttf-jetbrains-mono-nerd \
    docker-desktop teams-for-linux-bin \
    spotify spotifyd \
    code pamac-all obsidian-bin
print_success "Pacotes instalados!"

# ── Ferramentas standalone ───────────────────────────────────────────

# mise
if ! command -v mise &>/dev/null; then
    print_info "Instalando mise..."
    curl https://mise.jdx.dev/install.sh | sh
    print_success "mise instalado!"
else
    print_info "mise já instalado, pulando..."
fi

# uv
if ! command -v uv &>/dev/null; then
    print_info "Instalando uv..."
    curl -LsSf https://astral.sh/uv/install.sh | sh
    print_success "uv instalado!"
else
    print_info "uv já instalado, pulando..."
fi

# pnpm
if ! command -v pnpm &>/dev/null; then
    print_info "Instalando pnpm..."
    curl -fsSL https://get.pnpm.io/install.sh | sh -
    print_success "pnpm instalado!"
else
    print_info "pnpm já instalado, pulando..."
fi

# oh-my-posh
if ! command -v oh-my-posh &>/dev/null; then
    print_info "Instalando oh-my-posh..."
    curl -s https://ohmyposh.dev/install.sh | bash -s -- -d "$HOME/.local/bin"
    print_success "oh-my-posh instalado!"
else
    print_info "oh-my-posh já instalado, pulando..."
fi

# claude
if ! command -v claude &>/dev/null; then
    print_info "Instalando Claude Code..."
    curl -fsSL https://cli.anthropic.com/install.sh | sh
    print_success "Claude Code instalado!"
else
    print_info "Claude Code já instalado, pulando..."
fi

# opencode
if ! command -v open-code &>/dev/null; then
    print_info "Instalando OpenCode..."
    curl -fsSL https://opencode.ai/install | bash
    print_success "OpenCode instalado!"
else
    print_info "OpenCode já instalado, pulando..."
fi

# ── Criar diretórios ─────────────────────────────────────────────────
print_info "Criando diretórios..."
mkdir -p "$HOME/.local/bin"
mkdir -p "$HOME/.config"
mkdir -p "$HOME/workspace/github.com/janozeli"
mkdir -p "$HOME/workspace/github.com/hetosoft"

# ── Git configs por workspace ────────────────────────────────────────
print_info "Configurando identidades git por workspace..."

if [ ! -f "$HOME/workspace/github.com/janozeli/.gitconfig" ]; then
    cat > "$HOME/workspace/github.com/janozeli/.gitconfig" << 'EOF'
[user]
	email = lucasmjanozeli@gmail.com
	name = janozeli
EOF
    print_success "gitconfig pessoal criado!"
fi

if [ ! -f "$HOME/workspace/github.com/hetosoft/.gitconfig" ]; then
    cat > "$HOME/workspace/github.com/hetosoft/.gitconfig" << 'EOF'
[user]
	email = lucas.janozeli@hetosoft.com.br
	name = LucasJanozeli-Hetosoft
EOF
    print_success "gitconfig trabalho criado!"
fi

# ── Criar .env com chaves vazias ──────────────────────────────────────
if [ ! -f "$HOME/.env" ]; then
    cat > "$HOME/.env" << 'EOF'
export EXA_API_KEY=""
export CONTEXT7_API_KEY=""
EOF
    print_info ".env criado em ~/.env — preencha as chaves"
else
    print_info ".env já existe, pulando..."
fi

# ── Locale e teclado ─────────────────────────────────────────────────
print_info "Configurando locale e teclado..."
sudo localectl set-locale LANG=pt_BR.UTF-8
sudo localectl set-x11-keymap us,br pc105 altgr-intl,
print_success "Locale e teclado configurados!"

# ── Aplicar symlinks com GNU Stow ────────────────────────────────────
print_info "Aplicando symlinks com GNU Stow..."
stow --dotfiles -t "$HOME" .
print_success "Symlinks criados com sucesso!"

print_success "Instalação concluída! Reinicie o terminal ou execute: chsh -s \$(which zsh)"
