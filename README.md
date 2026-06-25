# SSM — Secure Shell Manager

> A fast, keyboard-driven TUI that makes your existing `~/.ssh/config` delightful to use at fleet scale.

[![Go](https://img.shields.io/github/go-mod/go-version/lfaoro/ssm?logo=go)](https://github.com/lfaoro/ssm)
[![Release](https://img.shields.io/github/v/release/lfaoro/ssm?logo=github)](https://github.com/lfaoro/ssm/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/lfaoro/ssm/go-tests.yml?branch=main&label=CI&logo=github)](https://github.com/lfaoro/ssm/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)
[![Downloads](https://img.shields.io/github/downloads/lfaoro/ssm/total?logo=github)](https://github.com/lfaoro/ssm/releases)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/lfaoro/ssm)

---

**No agents. No changes on your servers. Just a better way to navigate the SSH config you already have.**

![SSM TUI](data/demo.png)

## 30-second quick start

```bash
# 1. Install
curl -fsSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash

# 2. Tag a few hosts in ~/.ssh/config
Host prod-api
#tag: production,api
    HostName 10.0.0.42
    User deploy

# 3. Launch
ssm production          # filter to production hosts
ssm                     # or see everything
```

That's it. SSM reads your existing config (including `Include` directives), adds powerful navigation on top, and never touches your remote machines.

## What you get

- **Instant filtering** by tags, names, or fuzzy search (`/`)
- **Live reachability** — `p` pings the selected host, `P` pings everything visible (bounded concurrency based on CPU cores)
- **One-keystroke connect** — `Enter` (toggle SSH/Mosh with `Tab`)
- **Batch commands** — `ssm [tag] -r 'uptime && whoami'` runs across any filtered set, non-interactively
- **Integrated SFTP** — `Ctrl+s` opens a two-pane file browser with batch transfers
- **Remote execution** — `Ctrl+r` for interactive commands on the selected host
- **Cloud discovery** — `ssm sync hetzner|aws|gcp|azure` pulls running instances into your config automatically
- **Live config editing** — `Ctrl+e` opens your SSH config in `$EDITOR`, then reloads
- **Everything else** — copy host name with `y`/`Y`, config inspector (`Ctrl+v`), themes, `--ping` at startup, `--exit` for clean handoff, and more

## How tagging works

Add lightweight comments to your existing entries:

```ssh-config
#tagorder                 # optional: show tagged hosts first

Host web-01
#tag: production,web,eu
    HostName 203.0.113.10
    User deploy
    IdentityFile ~/.ssh/prod_ed25519

Host db-primary
#tag: production,database
    HostName 203.0.113.20
    ProxyJump web-01
```

Then use those tags as filters:

```bash
ssm production          # all production hosts
ssm web                 # anything tagged "web"
ssm eu,production       # combine tags
```

SSM supports full `Include` recursion, globs, `#tagorder`, and cycle detection.

## Cloud provider sync

Discover running servers and write them into your SSH config with zero manual work:

```bash
ssm sync                    # all configured providers
ssm sync aws hetzner        # specific providers
ssm sync --dry-run          # preview only
ssm sync --user deploy --key ~/.ssh/id_ed25519
```

- Each provider gets its own file under `~/.ssh/config.d/50-ssm-{provider}`
- `Include config.d/*` is added to your main config automatically
- Hosts are tagged with the provider name so you can filter with `ssm aws`

Supported providers and credentials:
- **Hetzner**: `HCLOUD_TOKEN`
- **AWS**: standard SDK chain (`AWS_PROFILE`, env vars, IAM role)
- **GCP**: `GCP_PROJECT` + Application Default Credentials
- **Azure**: `AZURE_SUBSCRIPTION_ID` + Azure auth chain

## Key bindings (TUI)

| Key          | Action                              |
|--------------|-------------------------------------|
| `Enter`      | Connect (SSH or Mosh)               |
| `Tab`        | Toggle SSH ↔ Mosh                   |
| `p`          | Ping selected host                  |
| `P`          | Ping all visible hosts              |
| `y` / `Y`    | Copy host name to clipboard         |
| `/`          | Fuzzy search / filter               |
| `q` or `Ctrl+c` | Quit                             |
| `Ctrl+e`     | Edit `~/.ssh/config` in `$EDITOR`   |
| `Ctrl+r`     | Run command on selected host        |
| `Ctrl+s`     | Open SFTP file browser              |
| `Ctrl+v`     | Toggle parsed config inspector      |
| `Ctrl+y`     | Open cloud sync panel               |
| `Esc`        | Clear filter / close panels         |

Emacs navigation (`Ctrl+p/n/b/f`) also works.

## CLI flags & batch usage

```bash
ssm [tag]                           # launch TUI, optionally filtered
ssm --backend mosh                  # launch with mosh as default for direct connections (Tab toggles; Ctrl+r and batch stay on ssh)
ssm exec prod 'uptime'              # run command on matching hosts and exit (recommended)
ssm e web --delay 150ms 'nginx -t'  # with pacing and concurrency control
ssm [tag] -r 'uptime'               # legacy form (still works; see --help)
ssm --exit prod-api           # connect and fully replace the ssm process
ssm --order production        # show tagged hosts first
ssm -t matrix                 # use a different theme (sky | matrix)
ssm -c ~/.ssh/work_config     # use a custom config file
# env var also works: SSM_BACKEND=mosh ssm
```

The `-r` / `--command` flag (and the newer `ssm exec` subcommand) are fully scriptable and exit non-zero if any host fails. Use `ssm exec --help` for the current options including `--delay`, `--threads`, and jitter control.

Note: `--backend` / `SSM_BACKEND` (and the Tab toggle) only affect direct Enter connections in the TUI. Batch execution, the legacy `-r` path, and the Ctrl+r "run command" submodel always use ssh.

## Security & hardening

SSM is built for production infrastructure teams:

- **Zero software** is ever installed on remote hosts
- Every SSH, mosh, and SFTP invocation uses the `--` delimiter to prevent flag injection
- `BatchMode=yes` + `RequestTTY=no` for all non-interactive operations
- Sensitive keys (`IdentityFile`, `ProxyCommand`, etc.) are filtered from the config viewer
- SSH config file permissions are checked (warns if not 0600)
- All remote output is sanitized (ANSI stripped, stderr truncated)
- Ping uses ordinary TCP connects — no raw sockets or elevated privileges

See [SECURITY.md](SECURITY.md) for the full model.

## Installation

**Fastest:**

```bash
curl -fsSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash
```

**Package managers:**

| Platform       | Command |
|----------------|---------|
| Go             | `go install github.com/lfaoro/ssm@latest` |
| macOS          | `brew install lfaoro/tap/ssm` |
| Arch Linux     | `yay -S ssm-bin` (AUR) |
| Nix            | `nix profile install github:lfaoro/ssm` |
| Nix (run)      | `nix run github:lfaoro/ssm` |
| Debian         | `sudo apt install ./ssm_*_linux_*.deb` download from [Releases](https://github.com/lfaoro/ssm/releases) |
| RPM            | `sudo rpm -i ssm_*_linux_*.rpm` download from [Releases](https://github.com/lfaoro/ssm/releases) |

Pre-built static binaries for Linux, macOS, FreeBSD, and OpenBSD (amd64 + arm64) are available on the [releases page](https://github.com/lfaoro/ssm/releases).

**From source:**

```bash
# requires Go
git clone https://github.com/lfaoro/ssm.git \
  && cd ssm \
  && make \
  && bin/ssm
```

## Development

See [CONTRIBUTING.md](CONTRIBUTING.md).

The project follows a strict “never commit unless explicitly told” policy for AI agents and careful humans. Full rules are in [AGENTS.md](AGENTS.md).

Release process is documented in [DEPLOY.md](DEPLOY.md).

## Sponsorship

SSM is developed in the open.

- [GitHub Sponsors](https://github.com/sponsors/lfaoro)
- BTC: `bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs`
- XMR: `89XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c`
- FIAT: [Revolut](https://revolut.me/matrix)
- Telegram: [@leonarth](https://t.me/leonarth)

## License

MIT © Leonardo Faoro and contributors
