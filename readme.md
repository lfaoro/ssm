# Secure Shell Manager

> Streamline SSH connections with a simple terminal UI

[![version][version-badge]](changelog.md)
[![license][license-badge]](license)
[![CI](https://github.com/lfaoro/ssm/actions/workflows/go-tests.yml/badge.svg)](https://github.com/lfaoro/ssm/actions/workflows/go-tests.yml)
[![Go report card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/lfaoro/ssm/badge)](https://scorecard.dev/viewer/?uri=github.com/lfaoro/ssm)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/12709/badge)](https://bestpractices.coreinfrastructure.org/projects/12709)
[![Downloads](https://img.shields.io/badge/dynamic/json?url=https://raw.githubusercontent.com/lfaoro/ssm/main/data/stats.json&query=$.total&label=downloads&color=blue)](https://github.com/lfaoro/ssm/releases)
[![Go version](https://img.shields.io/badge/Go-1.26.2-00ADD8.svg)](go.mod)

[version-badge]: https://img.shields.io/badge/version-1.0.0-blue.svg
[license-badge]: https://img.shields.io/badge/license-BSD3-blue

`ssm` is an SSH connection manager that works on top of your existing SSH config and installed `ssh`/`mosh` binaries. No setup required on remote systems.

**tl;dr** — [Install](#install)

---

## Features

- **Tag-based filtering** — `#tag: admin,vpn` comments in your SSH config become searchable metadata
- **Fuzzy search** — find hosts by name, hostname, user, or tag
- **SSH/MOSH dual protocol** — switch with `TAB`
- **Live config editing** — `ctrl+e` opens `$EDITOR`, auto-reloads on save
- **Run remote commands** — `ctrl+r` opens a command prompt, runs via `ssh -T`
- **Config inspection** — `ctrl+v` shows all params in a side panel
- **`--exit` flag** — connect and hand off the terminal, no lingering process
- **Theming** — `--theme sky|matrix`, extensible via `themes.go`

## Shoutout

**[@hackerschoice](https://x.com/hackerschoice/status/1920899798837711279)** on X

## Usage

```bash
ssm                    # launch the TUI
ssm admin              # filter by #tag: admin
ssm -se vpn            # --show --exit, filter by vpn tags
ssm -c ~/.ssh/other    # use a custom config file
ssm -o                 # show tagged hosts first
ssm --theme sky        # blue color scheme
ssm -d                 # debug mode with verbose logs
```

| Flag | Short | Description |
|---|---|---|
| `--show` | `-s` | show config in side panel on launch |
| `--exit` | `-e` | exit after connecting (hand off terminal) |
| `--order` | `-o` | show tagged hosts first |
| `--config` | `-c` | custom SSH config path |
| `--theme` | `-t` | color theme: `sky` or `matrix` |
| `--debug` | `-d` | debug mode with verbose log |

All flags support env vars: `SSM_SHOW`, `SSM_EXIT`, `SSM_ORDER`, `SSM_SSH_CONFIG_PATH`, `SSM_THEME`, `SSM_DEBUG`.

## Keys

| Key | Action |
|---|---|
| `enter` | connect to selected host |
| `ctrl+e` | edit SSH config in `$EDITOR` |
| `ctrl+v` | toggle config side panel |
| `ctrl+r` | run commands on host (no TTY) |
| `ctrl+c` | clear filter / quit |
| `tab` | switch between SSH and MOSH |
| `/` | filter hosts |
| `q` / `esc` | quit / exit filter |

## Quickstart

> New to SSH config? Start here. Otherwise skip to [Install](#install).

- [SSH config manual](https://man.openbsd.org/ssh_config.5)

```bash
# backup any existing config
[ -f ~/.ssh/config ] && cp ~/.ssh/config ~/.ssh/config.bak

# create a config
cat <<'EOF' >> ~/.ssh/config
#tagorder            # prioritize tagged hosts in list-view

Host myserver
#tag: production,web
    User admin
    HostName 10.0.0.5
    Port 2222
    IdentityFile ~/.ssh/id_rsa

Host terminalcoffee
#tag: shops
    User adam
    HostName terminal.shop
EOF

chmod 600 ~/.ssh/config
```

## Install

Download a binary from [releases](https://github.com/lfaoro/ssm/releases), or install via script/brew:

```bash
# shell script (linux, macos, freebsd, openbsd)
curl -sSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash

# homebrew (macos, linux)
brew install lfaoro/tap/ssm

# arch linux (AUR)
yay -S ssm-bin

# macos quarantine workaround (no paid signing key)
xattr -d com.apple.quarantine /path/to/ssm
```

Available for **4 OSes** × **2 architectures**: x86_64, arm64.

## Build

> Requires [Go](https://go.dev/doc/install) 1.26+

```bash
go install github.com/lfaoro/ssm@latest

# or clone and build
git clone https://github.com/lfaoro/ssm.git && cd ssm && make && bin/ssm

# sourcehut mirror
git clone https://git.sr.ht/~faoro/ssm && cd ssm && make && bin/ssm
```

## Development

```bash
go build ./...          # check compilation
go vet ./...            # static analysis
go test -race ./...     # tests with race detection
make build-static       # static binary (CGO_ENABLED=0)
make release-dev        # goreleaser snapshot (dry run)
```

## Resources

- [CLI flags reference](data/help)
- [SSH config example](data/config_example)
- [Changelog](changelog.md)

## Contributors

[See all](https://github.com/lfaoro/ssm/graphs/contributors)

Pull requests are welcome. Report a bug or request a feature by opening a [new issue](https://github.com/lfaoro/ssm/issues).

## Show support

> If `ssm` is useful to you, please consider giving it a ⭐.

- **star the repo**
- **tell your friends**
- **buy a VM** on [vibee.sh](https://vibee.sh)

[GitHub sponsor](https://github.com/sponsors/lfaoro) ·
[BTC](https://mempool.space/address/bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs) ·
[XMR](https://xmrchain.net/search?value=89XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c) ·
[FIAT](https://revolut.me/matrix)
