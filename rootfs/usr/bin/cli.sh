#!/bin/ash

export TERM=xterm-256color
export HOME=/root

trap '' INT

if command -v rlwrap > /dev/null 2>&1; then
    exec rlwrap --no-warnings -H /root/.lva_history lva banner
else
    exec lva banner
fi