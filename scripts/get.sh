#!/usr/bin/env bash
# Copyright (c) 2025 Leonardo Faoro & authors
# SPDX-License-Identifier: BSD-3-Clause
set -euo pipefail

APP_NAME=ssm
REPO="lfaoro/ssm"
API_URL="https://api.github.com/repos/${REPO}/releases/latest"
DOWNLOAD_URL="https://github.com/${REPO}/releases/download"

cleanup() {
	[[ -n "${TEMP_DIR:-}" ]] && rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

error() {
	echo "error: $1" >&2
	exit 1
}

usage() {
	cat <<EOF
Usage: $0 [OPTIONS]

Install ${APP_NAME} from GitHub releases.

Options:
  -h, --help       Show this help and exit
  --debug          Enable verbose debug output (set -x)
  --dir <path>     Install to <path> instead of auto-detecting

The script auto-detects:
  /usr/local/bin  (preferred, requires write permission)
  ~/.local/bin    (fallback, created if needed)

EOF
	exit 0
}

# ---- parse flags ----

CUSTOM_DIR=""
while [[ $# -gt 0 ]]; do
	case "$1" in
		-h|--help) usage ;;
		--debug)   set -x ;;
		--dir)
			[[ -z "${2:-}" ]] && error "--dir requires a path"
			CUSTOM_DIR="$2"
			shift
			;;
		*) error "unknown option: $1 (try --help)" ;;
	esac
	shift
done

# ---- pipe-to-bash warning ----

if [[ ! -t 0 ]]; then
	echo "warning: detected piped input. consider inspecting the script first:"
	echo "  curl -fsSL https://raw.githubusercontent.com/${REPO}/main/scripts/get.sh -o get.sh" >&2
fi

# ---- detect OS ----

raw_os="$(uname -s)"
goos="$(echo "$raw_os" | tr '[:upper:]' '[:lower:]')"
case "$goos" in
	linux|darwin|freebsd|openbsd) ;;
	*) error "unsupported operating system: $raw_os" ;;
esac

# ---- detect architecture ----

raw_arch="$(uname -m)"
case "$raw_arch" in
	x86_64|amd64)                  go_arch="x86_64" ;;
	aarch64|arm64)                 go_arch="arm64" ;;
	*) error "unsupported architecture: $raw_arch" ;;
esac

# ---- resolve install directory ----

resolve_install_dir() {
	if [[ -n "${CUSTOM_DIR}" ]]; then
		mkdir -p "$CUSTOM_DIR" || error "failed to create $CUSTOM_DIR"
		return
	fi

	# try /usr/local/bin first
	CUSTOM_DIR="/usr/local/bin"
	local tmp
	tmp="$(mktemp -t install_check_XXXXXX)" || error "failed to create temp file"
	if mv "$tmp" "${CUSTOM_DIR}/" 2>/dev/null; then
		rm -f "${CUSTOM_DIR}/$(basename "$tmp")"
	else
		rm -f "$tmp"
		CUSTOM_DIR="$HOME/.local/bin"
		mkdir -p "$CUSTOM_DIR" || error "failed to create $CUSTOM_DIR"
	fi
}
resolve_install_dir

# ---- fetch latest version ----

echo "Fetching latest version..."
API_RESPONSE="$(curl -sSL "$API_URL")" || error "failed to fetch from GitHub API"
VERSION="$(grep -o '"tag_name": "[^"]*"' <<<"$API_RESPONSE" | sed 's/"tag_name": "//;s/"//')"
[[ -n "$VERSION" ]] || {
	echo "debug: raw API response:" >&2
	echo "$API_RESPONSE" >&2
	error "failed to determine latest version"
}
echo "Found version: ${VERSION}"

# ---- build archive URL ----

ARCHIVE_NAME="${APP_NAME}_${VERSION}_${goos}_${go_arch}.tar.gz"
ARCHIVE_URL="${DOWNLOAD_URL}/${VERSION}/${ARCHIVE_NAME}"
echo "Downloading ${APP_NAME} ${VERSION} for ${goos}/${go_arch}..."
echo "  ${ARCHIVE_URL}"

# verify the URL exists before downloading
HTTP_STATUS="$(curl -fsSL -o /dev/null -w '%{http_code}' "$ARCHIVE_URL")"
if [[ "$HTTP_STATUS" != "200" ]]; then
	echo "error: HTTP ${HTTP_STATUS} fetching ${ARCHIVE_URL}" >&2
	echo "Available assets:" >&2
	grep -o '"browser_download_url": "[^"]*"' <<<"$API_RESPONSE" | sed 's/"browser_download_url": "//;s/"//' >&2
	error "download URL not accessible"
fi

# ---- download and extract ----

TEMP_DIR="$(mktemp -d)" || error "failed to create temp directory"

echo "Downloading..."
curl -fsSL "$ARCHIVE_URL" -o "${TEMP_DIR}/${ARCHIVE_NAME}" || error "download failed"

echo "Extracting..."
tar -xzf "${TEMP_DIR}/${ARCHIVE_NAME}" -C "$TEMP_DIR" || error "failed to extract archive"

echo "Installing..."
mv "${TEMP_DIR}/${APP_NAME}" "${CUSTOM_DIR}/${APP_NAME}" || error "failed to install binary"
chmod +x "${CUSTOM_DIR}/${APP_NAME}" || error "failed to set executable permissions"

BINARY_PATH="${CUSTOM_DIR}/${APP_NAME}"

# ---- path check ----

echo "Installed ${APP_NAME} to: ${BINARY_PATH}"

if [[ ":$PATH:" != *":${CUSTOM_DIR}:"* ]]; then
	echo "Note: ${CUSTOM_DIR} is not in your PATH"
	case "${SHELL##*/}" in
		bash) echo "  Add: echo 'export PATH=\$PATH:${CUSTOM_DIR}' >> ~/.bashrc" ;;
		zsh)  echo "  Add: echo 'export PATH=\$PATH:${CUSTOM_DIR}' >> ~/.zshrc" ;;
		*)    echo "  Add ${CUSTOM_DIR} to your PATH" ;;
	esac
fi

echo ""
if [[ "$goos" == "darwin" ]] || command -v brew &>/dev/null; then
	echo "Or via Homebrew: brew install lfaoro/tap/ssm"
fi

# ---- verify ----

"$BINARY_PATH" --version || error "installed binary failed to run"
