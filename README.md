# SSM 🐚 – Secure Shell Manager
Your SSH config on TUI-roids.

> SSM is a lightweight, open-source terminal UI (TUI) that sits seamlessly on top of your existing ~/.ssh/config (and installed ssh/mosh binaries) to eliminate SSH friction. It delivers fast, interactive host discovery, one-keystroke connections, integrated SFTP file transfers, remote command execution, tag-based filtering, live config editing, and more — with zero setup or changes required on any remote server.

[![Go](https://img.shields.io/github/go-mod/go-version/lfaoro/ssm?logo=go)](https://github.com/lfaoro/ssm)
[![Release](https://img.shields.io/github/v/release/lfaoro/ssm?logo=github)](https://github.com/lfaoro/ssm/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/lfaoro/ssm/go-tests.yml?branch=main&label=CI&logo=github)](https://github.com/lfaoro/ssm/actions)
[![Downloads](https://img.shields.io/github/downloads/lfaoro/ssm/total?logo=github)](https://github.com/lfaoro/ssm/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)
[![License](https://img.shields.io/github/license/lfaoro/ssm)](LICENSE)


## ✨ Features

SSM turns your plain SSH config into a delightful, keyboard-driven experience. Here’s what you get out of the box:

### 🔍 Discovery & Navigation
| Feature       | Description                                                                 | Shortcut      |
|---------------|-----------------------------------------------------------------------------|---------------|
| **Tags**      | Add `#tag: production` (or any label) in your SSH config and instantly filter hosts | `ssm production` |
| **Fuzzy Search** | Lightning-fast fuzzy search across all hosts                                 | `/`           |
| **Ping**      | Check reachability of one host or all hosts at once (capped at 50 concurrent for safety) | `p` / `P`     |

### 🚀 Connection & Interaction
| Feature          | Description                                                                 | Shortcut      |
|------------------|-----------------------------------------------------------------------------|---------------|
| **Connect**      | Connect via SSH or Mosh with a single keystroke                             | `tab` to toggle |
| **Remote Exec**  | Run any command on the selected host without leaving the TUI                | `ctrl+r`      |
| **Exit Mode**    | Clean exit that replaces the process entirely (`syscall.Exec`) — no leftover shell | `--exit` flag |

### 📁 File Management
| Feature     | Description                                                                 | Shortcut      |
|-------------|-----------------------------------------------------------------------------|---------------|
| **SFTP**    | Beautiful two-pane local ↔ remote file browser with batch transfer support   | `ctrl+s`      |

### ⚙️ Configuration & Power Tools
| Feature            | Description                                                                 | Shortcut      |
|--------------------|-----------------------------------------------------------------------------|---------------|
| **Live Editor**    | Edit your SSH config directly inside SSM                                     | `ctrl+e`      |
| **Config Inspector**| View sanitized, readable version of the parsed config                        | `ctrl+v`      |
| **Advanced Parser**| Full `Include` recursion (depth 10 + cycle detection) + `#tagorder` sorting | —             |

### 🎨 Theming & Security
| Feature     | Description                                                                 | Details                  |
|-------------|-----------------------------------------------------------------------------|--------------------------|
| **Themes**  | Beautiful built-in themes for different vibes                               | `sky` (default) • `matrix` |
| **Security**| Hardened defaults including `BatchMode=yes` and injection protection        | `--` anti-injection delimiter |

> **Pro tip:** Run `ssm --theme matrix` for that classic green-terminal hacker aesthetic.

## 🚀 Quick Start

SSM is designed to feel instant. Here are the most common ways to launch it:

```bash
# Launch the full TUI (recommended first command)
ssm

# Filter to only hosts tagged with "production"
ssm production

# Advanced one-liner: show config + auto-exit on connect + filter "vpn" hosts
ssm --show --exit vpn

# Choose your aesthetic
ssm --theme matrix    # classic green terminal vibes
ssm --theme sky       # soft modern blue (default)

# Run instantly without installing (Nix flake)
nix run github:lfaoro/ssm -- ssm
```

## Install

### One-liner

```bash
curl -fsSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash
```

### macOS users - remove quarantine flag
```
xattr -d com.apple.quarantine $(which ssm)
```

### Package Managers

| Platform | Command |
|---|---|
| Go | `go install github.com/lfaoro/ssm@latest` |
| macOS | `brew install lfaoro/tap/ssm` |
| Arch Linux | `yay -S ssm-bin` (AUR) |
| Nix install | `nix profile install github:lfaoro/ssm` |
| Nix run | `nix run github:lfaoro/ssm -- ssm` |
| deb / rpm | Download from [Releases](https://github.com/lfaoro/ssm/releases) |

### Pre-built Binaries

Download the latest archive for your platform from the [releases page](https://github.com/lfaoro/ssm/releases), then:

```bash
tar xzf ssm_*.tar.gz
sudo mv ssm /usr/local/bin/
```

## SSH Config Tips

Add tags like this:

```ssh-config
Host myserver
#tag: production,web
    User admin
    HostName 10.0.0.5
    ...
```

## Architecture

```
ssm
├── main.go              # CLI entry (urfave/cli v3), version check, syscall.Exec
└── pkg/
    ├── sshconf/         # SSH config parser
    │   ├── parser.go    # Thread-safe parsing, Include recursion, #tag: comments
    │   └── util.go      # Helpers, sensitive key filtering, symlink resolution
    └── tui/             # Bubbletea v2 TUI
        ├── model.go     # Root model, state management, sub-model coordination
        ├── list.go      # Host list, fuzzy search, tag filtering
        ├── runcmd.go    # Remote command execution sub-model
        ├── sftp.go      # SFTP file browser sub-model
        ├── syscmd.go    # SSH/mosh process management, signal handling
        ├── log.go       # Debug logging (debug-mode only)
        └── themes.go    # Color themes (sky, matrix)
```

## Development

Requires [Go 1.26+](https://go.dev/doc/install).

```bash
git clone https://github.com/lfaoro/ssm.git && cd ssm
make build              # compile binary to ./bin/ssm
make test               # run all tests with race detection
make lint               # golangci-lint (or go fmt + go vet)
make bench              # run benchmarks
make release-dev        # goreleaser snapshot (dry run)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for dev environment setup and guidelines.

## Security

- SSH stderr sanitized (truncated to 500 chars)
- ANSI escape sequences stripped from remote output
- Sensitive keys (identityfile, proxycommand, etc.) filtered from config viewport
- SFTP uses `BatchMode=yes` + `RequestTTY=no` to prevent interactive prompts
- `--` delimiter before hostname in all SSH/mosh/syscall invocations (anti-injection)


Star it if it helps → https://github.com/lfaoro/ssm

Made with ❤️ and too much SSH pain.

## Shoutouts

**[@hackerschoice](https://x.com/hackerschoice/status/1920899798837711279)** on X
**[@golangch](https://x.com/golangch/status/1920138613473649150)** on X

If you live in the terminal and manage more than a couple servers, this thing just makes life a little nicer.

- [GitHub Sponsors](https://github.com/sponsors/lfaoro)
- BTC: `bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs`
- XMR: `89XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c`
- FIAT: [Revolut](https://revolut.me/matrix)
- [message me on Telegram](https://t.me/leonarth)

or just ⭐ the repo. Appreciate it either way.
