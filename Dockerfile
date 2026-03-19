FROM archlinux:base-devel

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm --needed sudo git curl

RUN useradd -m -G wheel -s /bin/bash testuser && \
    echo "testuser:test" | chpasswd && \
    echo "%wheel ALL=(ALL) ALL" >> /etc/sudoers

USER testuser
WORKDIR /home/testuser

CMD ["bash", "-c", "curl -fsSL janozeli.github.io/install.sh | bash; exec bash"]
