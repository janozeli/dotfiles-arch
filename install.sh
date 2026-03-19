#!/usr/bin/env bash
set -e

# Cores para saída
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para imprimir mensagens
print_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Verificar se yay está instalado
if ! command -v yay &> /dev/null; then
    print_info "Instalando yay..."
    sudo pacman -S --needed git base-devel
    git clone https://aur.archlinux.org/yay.git
    cd yay
    makepkg -si
    cd ..
    rm -rf yay
    yay -Y --gendb
    yay -Syu --devel
    yay -Y --devel --save
    print_success "yay instalado com sucesso!"
else
    print_info "yay já instalado, pulando..."
fi

# Instalar todas as dependências
print_info "Instalando dependências..."
yay -S --noconfirm --needed stow zsh git fzf zoxide eza bat ghostty gh wl-clipboard mpv less wget curl unzip tldr yt-dlp zed firefox zen-browser-bin oh-my-posh-bin asdf-vm ttf-firacode-nerd pamac-all cursor-bin obsidian-bin spotify
print_success "Dependências instaladas com sucesso!"

# Criar Pastas se não existirem
print_info "Criando pastas..."
mkdir -p $HOME/.local/bin
mkdir -p $HOME/.config

# Aplicar symlinks com GNU Stow
print_info "Aplicando symlinks com GNU Stow..."
stow --dotfiles -t $HOME .
print_success "Symlinks criados com sucesso!"
