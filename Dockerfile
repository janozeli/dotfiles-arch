FROM archlinux:base

RUN pacman -Syu --noconfirm && \
    pacman -S --noconfirm --needed sudo git base-devel

RUN useradd -m -G wheel -s /bin/bash testuser && \
    echo "testuser ALL=(ALL) NOPASSWD: ALL" >> /etc/sudoers

COPY --chown=testuser:testuser . /home/testuser/dotfiles-arch

USER testuser
WORKDIR /home/testuser/dotfiles-arch

CMD ["bash", "install.sh"]
