FROM archlinux:base

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm --needed sudo git base-devel curl

RUN useradd -m -G wheel -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

USER testuser
WORKDIR /home/testuser

ENV YAY_PKG=yay-bin
CMD ["bash", "-c", "curl -fsSL janozeli.github.io/install.sh | bash; exec bash"]
