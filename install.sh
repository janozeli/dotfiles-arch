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

# ── sudoers NOPASSWD ─────────────────────────────────────────────────
CURRENT_USER="$(whoami)"
print_info "Configurando NOPASSWD para $CURRENT_USER..."
echo "$CURRENT_USER ALL=(ALL) NOPASSWD: ALL" | sudo tee "/etc/sudoers.d/$CURRENT_USER" >/dev/null
sudo chmod 440 "/etc/sudoers.d/$CURRENT_USER"
print_success "NOPASSWD configurado!"

# ── yay ──────────────────────────────────────────────────────────────
if ! command -v yay &>/dev/null; then
    print_info "Instalando yay..."
    sudo pacman -S --noconfirm --needed git base-devel
    YAY_PKG="${YAY_PKG:-yay}"
    git clone "https://aur.archlinux.org/${YAY_PKG}.git" /tmp/yay
    cd /tmp/yay
    makepkg -si --noconfirm
    cd -
    rm -rf /tmp/yay
    yay -Y --gendb
    yay -Syu --noconfirm --devel
    yay -Y --devel --save
    print_success "yay instalado!"
else
    print_info "yay já instalado, pulando..."
fi

# ── pacman.conf ──────────────────────────────────────────────────────
print_info "Configurando pacman.conf..."
sudo sed -i 's/^#ParallelDownloads.*/ParallelDownloads = 15\nILoveCandy/' /etc/pacman.conf
print_success "pacman.conf configurado!"

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

eval "$("$HOME/.local/bin/mise" activate bash)"
mise install node@lts uv@latest pnpm@latest
mise use --global node@lts uv@latest pnpm@latest
print_success "Node (LTS), uv e pnpm instalados via mise!"

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

# ── Claude MCP servers ───────────────────────────────────────────────
if command -v claude &>/dev/null; then
    print_info "Configurando MCP servers do Claude..."
    claude mcp add --scope user exa -e EXA_API_KEY="${EXA_API_KEY}" -- npx -y exa-mcp-server --tools=web_search_exa,get_code_context_exa,web_search_advanced_exa,crawling_exa
    claude mcp add --scope user context7 -- npx -y @upstash/context7-mcp --api-key "${CONTEXT7_API_KEY}"
    print_success "MCP servers configurados!"
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
