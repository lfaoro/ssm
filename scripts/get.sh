#!/usr/bin/env bash
# SPDX-License-Identifier: AGPL-3.0-only
# 
# Copyright (c) 2025 Leonardo Faoro. All rights reserved.
# Use of this source code is governed by the AGPL-3.0 license
# found in the LICENSE file.

set -euo pipefail

cleanup() {
    if [ -n "${TEMP_FILE:-}" ]; then
        rm -f "$TEMP_FILE"
    fi
}
trap cleanup EXIT

error() {
    echo "error: $1" >&2
    exit 1
}

is_writable() {
    local dir="$1"
    if [ ! -d "$dir" ]; then
        return 1
    fi
    local temp_check=$(mktemp -t swap_install_check_XXXXXX) || error "failed to create temp file"
    if ! mv "$temp_check" "$dir/" 2>/dev/null; then
        rm -f "$temp_check"
        return 1
    fi
    rm -f "$dir/$(basename "$temp_check")"
    return 0
}

check_permissions() {
    local dir="$1"
    TEMP_FILE=$(mktemp -t swap_install_XXXXXX) || error "failed to create temp file"
    if ! mv "$TEMP_FILE" "$dir/" 2>/dev/null; then
        echo "warning: no write permission in $dir"
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR" || error "failed to create $INSTALL_DIR"
    fi
    rm -f "$dir/$(basename "$TEMP_FILE")" 2>/dev/null
}

check_path() {
    local dir="$1"
    if [[ ":$PATH:" != *":$dir:"* ]]; then
        echo "Warning: $dir is not in your PATH"
        case "$SHELL" in
            *bash) echo "Run: echo 'export PATH=\$PATH:$dir' >> ~/.bashrc" ;;
            *zsh)  echo "Run: echo 'export PATH=\$PATH:$dir' >> ~/.zshrc" ;;
            *)     echo "Add $dir to your PATH" ;;
        esac
    fi
}

# Configuration
APP_NAME=ssm
REPO="lfaoro/ssm"
LATEST_RELEASE_URL="https://github.com/${REPO}/releases/latest"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "${ARCH}" in
    x86_64|amd64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: ${ARCH}" ;;
esac
case "${OS}" in
    linux)
        BINARY_NAME="${APP_NAME}_linux_${ARCH}"
        if is_writable "/usr/local/bin"; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="$HOME/.local/bin"
        fi
        ;;
    darwin)
        BINARY_NAME="${APP_NAME}_darwin_${ARCH}"
        if is_writable "/usr/local/bin"; then
            INSTALL_DIR="/usr/local/bin"
        else
            INSTALL_DIR="$HOME/.local/bin"
        fi
        ;;
    msys*|mingw*)
        OS="windows"
        BINARY_NAME="${APP_NAME}_windows_${ARCH}.exe"
        INSTALL_DIR="$HOME/bin"
        ;;
    *) error "Unsupported operating system: ${OS}" ;;
esac

# get latest version
echo "Fetching latest version..."
VERSION=$(curl -sSL "https://api.github.com/repos/${REPO}/releases/latest" | grep -o '"tag_name": "v[^"]*"' | sed 's/"tag_name": "//;s/"//')
[ -z "$VERSION" ] && error "failed to determine latest version"

# create installation directory
mkdir -p "${INSTALL_DIR}" || error "failed to create installation directory"

# only check permissions if we're not already in a fallback directory
if [[ "$INSTALL_DIR" != "/tmp" && "$INSTALL_DIR" != "$HOME/.local/bin" && "$INSTALL_DIR" != "$HOME/bin" ]]; then
    check_permissions "$INSTALL_DIR"
fi

# download and install binary
DOWNLOAD_BINARY_URL="${DOWNLOAD_URL}/${VERSION}/${BINARY_NAME}"
echo "Downloading ${APP_NAME} ${VERSION} for ${OS}/${ARCH}..."

# verify the download url exists before attempting to download
if ! curl --output /dev/null --silent --head --fail "${DOWNLOAD_BINARY_URL}"; then
    error "download URL not accessible: ${DOWNLOAD_BINARY_URL}"
fi

if [ "${OS}" = "windows" ]; then
    curl -fsSL "${DOWNLOAD_BINARY_URL}" -o "${INSTALL_DIR}/${APP_NAME}.exe" --progress-bar || error "Download failed"
    chmod +x "${INSTALL_DIR}/${APP_NAME}.exe" || error "failed to set executable permissions"
    BINARY_PATH="${INSTALL_DIR}/${APP_NAME}.exe"
else
    curl -fsSL "${DOWNLOAD_BINARY_URL}" -o "${INSTALL_DIR}/${APP_NAME}" --progress-bar || error "Download failed"
    chmod +x "${INSTALL_DIR}/${APP_NAME}" || error "failed to set executable permissions"
    BINARY_PATH="${INSTALL_DIR}/${APP_NAME}"
fi

echo "Successfully installed ${APP_NAME} to: ${BINARY_PATH}"
check_path "${INSTALL_DIR}"

# verify
"${BINARY_PATH}" --version || error "failed to run ${APP_NAME}"
