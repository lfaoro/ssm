# SSM — Secure Shell Manager

A terminal interface that makes your existing `~/.ssh/config` fast, reliable, and actually pleasant to use at scale — with no changes required on any remote server.

[![Go](https://img.shields.io/github/go-mod/go-version/lfaoro/ssm?logo=go)](https://github.com/lfaoro/ssm)
[![Release](https://img.shields.io/github/v/release/lfaoro/ssm?logo=github)](https://github.com/lfaoro/ssm/releases)
[![CI](https://img.shields.io/github/actions/workflow/status/lfaoro/ssm/go-tests.yml?branch=main&label=CI&logo=github)](https://github.com/lfaoro/ssm/actions)
[![Downloads](https://img.shields.io/github/downloads/lfaoro/ssm/total?logo=github)](https://github.com/lfaoro/ssm/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)
[![License](https://img.shields.io/github/license/lfaoro/ssm)](LICENSE)

---

## You’ll recognize this

You manage infrastructure through SSH and already maintain a detailed `~/.ssh/config`.

Your hosts are organized by environment, role, region, or team. You frequently need to:

- Connect to many different servers throughout the day
- Run commands across groups of hosts
- Transfer files without leaving your terminal
- Quickly see which hosts are reachable
- Keep your configuration organized and discoverable

You want a tool that makes this workflow faster and more reliable — without installing anything on your servers.

SSM exists for exactly this situation.

## See it in action

<!-- TODO: Replace with real 30-second asciinema recording -->
[![Demo](data/demo.cast)](data/demo.cast)

> **Demo coming soon.** A 30-second asciinema recording will be placed here showing typical daily usage (tagging, filtering, remote commands, SFTP, and pings).

## Why people use SSM

- **Your SSH config is the source of truth** — SSM reads `~/.ssh/config` (including `Include` directives) and adds the organization you’ve always wanted through simple `#tag:` comments.
- **Fast, reliable discovery at fleet scale** — Filter by tags, fuzzy search, custom ordering, and live reachability checks across all your hosts.
- **One interface for the full loop** — SSH or Mosh connections, ad-hoc command execution, and two-pane SFTP file transfers — all without context switching.
- **Built for people who live in the terminal** — Clean, fast, and respectful of how experienced operators actually work.

## How it works with what you already have

Add lightweight metadata to your existing entries:

```ssh-config
Host prod-api
#tag: production,api
    User deploy
    HostName api.example.com
    ...

Host db-primary
#tag: production,database,eu
    User postgres
    HostName db.example.com
    ProxyJump prod-api
```

Then launch SSM:

```bash
ssm                  # see everything
ssm production       # filter to production hosts
ssm db               # or any tag you use
```

No changes are ever made to your remote servers.

## Key Capabilities

### Discovery & Navigation

| Capability          | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **Tag-based filtering** | Use `#tag: production,web` in your config and filter instantly with `ssm production` |
| **Fuzzy search**        | Fast search across all hosts with `/`                                       |
| **Live reachability**   | Ping one host or your entire fleet (capped at 50 concurrent) with `p` / `P` |

### Connection & Interaction

| Capability          | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **One-keystroke connect** | Tab to switch between SSH and Mosh, Enter to connect                      |
| **Remote execution**      | Run any command on the selected host with `Ctrl+r` without opening a full shell |
| **Clean exit mode**       | `--exit` flag uses `syscall.Exec` so the process is fully replaced        |

### File Management

| Capability          | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **Integrated SFTP** | Two-pane local ↔ remote file browser with batch transfers (`Ctrl+s`)      |

### Configuration & Power Tools

| Capability          | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **Live editor**         | Edit your SSH config directly from inside SSM (`Ctrl+e`)                    |
| **Config inspector**    | View a clean, sanitized version of the parsed config (`Ctrl+v`)             |
| **Advanced parsing**    | Full `Include` recursion, cycle detection, glob support, and `#tagorder`    |

### Theming & Safety

| Capability          | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| **Themes**              | `sky` (default) and `matrix`                                                |
| **Hardened defaults**   | `BatchMode=yes`, `--` anti-injection delimiter, permission checks, sensitive key filtering |

## Security & Reliability

SSM is designed for people who operate production infrastructure:

- No agents or software installed on remote servers
- SSH config permissions are checked and warned about
- Sensitive keys (`IdentityFile`, `ProxyCommand`, etc.) are filtered from the inspector
- All remote command and SFTP connections use `BatchMode=yes` and the `--` delimiter
- Stderr is sanitized and truncated
- Ping uses ordinary TCP connects — no raw sockets or privileges required

## Installation

The fastest way to try it:

```bash
curl -fsSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash
```

### Package Managers

| Platform       | Command                                              |
|----------------|------------------------------------------------------|
| Go             | `go install github.com/lfaoro/ssm@latest`            |
| macOS          | `brew install lfaoro/tap/ssm`                        |
| Arch Linux     | `yay -S ssm-bin` (AUR)                               |
| Nix            | `nix profile install github:lfaoro/ssm`              |
| Nix (run)      | `nix run github:lfaoro/ssm -- ssm`                   |
| Debian / RPM   | Download from [Releases](https://github.com/lfaoro/ssm/releases) |

Pre-built binaries for Linux, macOS, FreeBSD, and OpenBSD (amd64 + arm64) are available on the [releases page](https://github.com/lfaoro/ssm/releases).

## What people are saying

> “This is exactly what I wanted for my fleet.”  
> — [@hackerschoice](https://x.com/hackerschoice/status/1920899798837711279)

> “Finally a TUI that respects how I already manage SSH.”  
> — [@golangch](https://x.com/golangch/status/1920138613473649150)

If you’re using SSM in production or at scale, I’d love to hear about it.

## Sponsorship & Support

SSM is developed and maintained in the open.

- [GitHub Sponsors](https://github.com/sponsors/lfaoro)
- BTC: `bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs`
- XMR: `89XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c`
- FIAT: [Revolut](https://revolut.me/matrix)
- Telegram: [@leonarth](https://t.me/leonarth)

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup and contribution workflow.

The project follows a strict “never commit unless explicitly told” policy. See [AGENTS.md](AGENTS.md) for the full rules and rationale.

Releases are performed manually. The exact process is documented in [DEPLOY.md](DEPLOY.md).

## License

MIT © Leonardo Faoro and contributors
