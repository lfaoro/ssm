#!/usr/bin/env bash
# Copyright (c) 2025 Leonardo Faoro & authors
# SPDX-License-Identifier: MIT
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
  -h, --help            Show this help and exit
  --debug               Enable verbose debug output (set -x)
  --dir <path>          Install to <path> instead of auto-detecting
  --no-modify-path      Do not modify shell rc files (print instructions only)
  --modify-path         Modify shell rc files without prompting (for supported shells)

The script auto-detects:
  /usr/local/bin  (preferred, requires write permission)
  ~/.local/bin    (fallback, created if needed)

For bash, zsh, and fish, the script will automatically add the install directory
to your PATH (in the appropriate rc file) when possible. Use --no-modify-path
to disable this behavior.

EOF
	exit 0
}

# ---- shell rc path helpers ----

detect_shell() {
	local shell
	shell="${SHELL:-}"
	if [[ -z "$shell" ]]; then
		shell="$(ps -p $$ -o comm= 2>/dev/null || true)"
	fi
	if [[ -z "$shell" ]]; then
		shell="${0##*/}"
	fi
	basename "$shell" 2>/dev/null || echo "unknown"
}

is_ci_environment() {
	[[ -n "${CI:-}" ]] ||
	[[ -n "${GITHUB_ACTIONS:-}" ]] ||
	[[ -n "${GITLAB_CI:-}" ]] ||
	[[ -n "${JENKINS_URL:-}" ]] ||
	[[ -n "${BUILD_NUMBER:-}" ]] ||
	[[ -n "${TF_BUILD:-}" ]] ||
	[[ -n "${TRAVIS:-}" ]]
}

get_target_rc_file() {
	local shell="$1"
	case "$shell" in
		bash) echo "$HOME/.bashrc" ;;
		zsh)  echo "$HOME/.zshrc" ;;
		fish) echo "$HOME/.config/fish/config.fish" ;;
		*)    echo "" ;;
	esac
}

get_path_line() {
	local shell="$1"
	local dir="$2"
	case "$shell" in
		bash|zsh)
			echo "export PATH=\"\$PATH:${dir}\""
			;;
		fish)
			echo "fish_add_path ${dir}"
			;;
		*)
			echo ""
			;;
	esac
}

path_already_configured() {
	local file="$1"
	local line="$2"
	[[ -f "$file" ]] && grep -Fq "$line" "$file"
}

append_to_rc_file() {
	local file="$1"
	local line="$2"
	local comment="$3"

	mkdir -p "$(dirname "$file")" 2>/dev/null || true

	if [[ ! -f "$file" ]]; then
		touch "$file" || return 1
	fi

	{
		echo ""
		echo "$comment"
		echo "$line"
	} >> "$file"
}

should_modify_path() {
	# Explicit flags take precedence
	if [[ "$MODIFY_PATH" == "no" ]]; then
		return 1
	fi
	if [[ "$MODIFY_PATH" == "yes" ]]; then
		return 0
	fi

	# Default safety rules
	if ! [[ -t 0 && -t 1 ]]; then
		return 1
	fi
	if is_ci_environment; then
		return 1
	fi

	return 0
}

# ---- parse flags ----

CUSTOM_DIR=""
MODIFY_PATH=""          # "", "yes", or "no"
while [[ $# -gt 0 ]]; do
	case "$1" in
		-h|--help) usage ;;
		--debug)   set -x ;;
		--dir)
			[[ -z "${2:-}" ]] && error "--dir requires a path"
			CUSTOM_DIR="$2"
			shift 2
			;;
		--no-modify-path)
			MODIFY_PATH="no"
			shift
			;;
		--modify-path)
			MODIFY_PATH="yes"
			shift
			;;
		*) error "unknown option: $1 (try --help)" ;;
	esac
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
if [[ -n "${GITHUB_TOKEN:-}" ]]; then
	API_RESPONSE="$(curl -sSL -H "Authorization: Bearer ${GITHUB_TOKEN}" "$API_URL")" || error "failed to fetch from GitHub API"
else
	API_RESPONSE="$(curl -sSL "$API_URL")" || error "failed to fetch from GitHub API"
fi
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

echo "Downloading archive..."
curl -fsSL "$ARCHIVE_URL" -o "${TEMP_DIR}/${ARCHIVE_NAME}" || error "download failed"

# ---- verify checksum ----

CHECKSUM_FILE="ssm_${VERSION}_checksums.txt"
CHECKSUM_URL="${DOWNLOAD_URL}/${VERSION}/${CHECKSUM_FILE}"
echo "Downloading checksums..."
curl -fsSL "$CHECKSUM_URL" -o "${TEMP_DIR}/${CHECKSUM_FILE}" || {
	echo "warning: checksum file not available, skipping verification"
}

if [[ -f "${TEMP_DIR}/${CHECKSUM_FILE}" ]]; then
	echo "Verifying checksum..."
	cd "$TEMP_DIR"
	if command -v sha256sum &>/dev/null; then
		expected=$(grep "$ARCHIVE_NAME" "${CHECKSUM_FILE}" | awk '{print $1}')
		actual=$(sha256sum "${ARCHIVE_NAME}" | awk '{print $1}')
	elif command -v shasum &>/dev/null; then
		expected=$(grep "$ARCHIVE_NAME" "${CHECKSUM_FILE}" | awk '{print $1}')
		actual=$(shasum -a 256 "${ARCHIVE_NAME}" | awk '{print $1}')
	else
		echo "warning: no SHA256 tool found, skipping verification"
		expected=""
		actual=""
	fi

	if [[ -n "$expected" && -n "$actual" ]]; then
		if [[ "$expected" == "$actual" ]]; then
			echo "Checksum verified ✓"
		else
			error "checksum mismatch! expected $expected, got $actual"
		fi
	fi
	cd - >/dev/null
fi

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

	shell=$(detect_shell)
	rc_file=$(get_target_rc_file "$shell")
	line=$(get_path_line "$shell" "$CUSTOM_DIR")
	comment="# Added by ssm install script on $(date +%Y-%m-%d)"

	case "$shell" in
		bash|zsh|fish)
			if should_modify_path && [[ -n "$rc_file" ]] && [[ -n "$line" ]]; then
				# Preferred path: interactive + not CI
				if path_already_configured "$rc_file" "$line"; then
					echo "  (already present in $rc_file)"
				elif append_to_rc_file "$rc_file" "$line" "$comment"; then
					echo "  Added to $rc_file"
				else
					echo "  Could not write to $rc_file"
					echo "  Please add this line manually:"
					echo "    $line"
				fi
			elif [[ "$shell" =~ ^(bash|zsh|fish)$ ]] && ! is_ci_environment && [[ "$MODIFY_PATH" != "no" ]] && [[ -n "$rc_file" ]] && [[ -n "$line" ]]; then
				# Aggressive path for supported shells (e.g. curl | bash in interactive terminal)
				if path_already_configured "$rc_file" "$line"; then
					echo "  (already present in $rc_file)"
				elif append_to_rc_file "$rc_file" "$line" "$comment"; then
					echo "  Added to $rc_file"
				else
					echo "  Could not write to $rc_file"
					echo "  Please add this line manually:"
					echo "    $line"
				fi
			else
				if [[ -n "$rc_file" ]]; then
					echo "  Add this line to $rc_file:"
				else
					echo "  Add this line to your $shell config:"
				fi
				echo "    $line"
			fi
			;;
		*)
			echo "  Add ${CUSTOM_DIR} to your PATH"
			;;
	esac
fi

echo ""
if [[ "$goos" == "darwin" ]] || command -v brew &>/dev/null; then
	echo "Or via Homebrew: brew install lfaoro/tap/ssm"
fi

# ---- verify ----

"$BINARY_PATH" --version || error "installed binary failed to run"
