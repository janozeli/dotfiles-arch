#!/usr/bin/env bash
grep -q "^LANG=pt_BR" /etc/locale.conf 2>/dev/null \
    && locale -a 2>/dev/null | grep -q "pt_BR" \
    && localectl status 2>/dev/null | grep -q "X11 Layout.*us,br"
