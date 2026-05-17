#!/usr/bin/env bash
set -euo pipefail

SSH_DIR="$HOME/.ssh"
CONFIG="$SSH_DIR/config"

if [[ ! -d "$SSH_DIR" ]]; then
	mkdir -p "$SSH_DIR"
	chmod 700 "$SSH_DIR"
	echo "Created $SSH_DIR"
else
	echo "$SSH_DIR already exists"
fi

if [[ -f "$CONFIG" ]]; then
	echo "$CONFIG already exists"
else
	touch "$CONFIG"
	echo "Created $CONFIG"
fi

chmod 600 "$CONFIG"
echo "Permissions set to 600 on $CONFIG"
