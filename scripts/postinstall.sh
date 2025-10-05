#!/bin/sh

if command -v bash >/dev/null 2>&1; then
    :
fi

if command -v zsh >/dev/null 2>&1; then
    if [ -d ~/.zcompdump ]; then
        rm -f ~/.zcompdump*
    fi
fi

if command -v fish >/dev/null 2>&1; then
    :
fi

exit 0
