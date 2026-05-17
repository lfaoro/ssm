# Secure Shell Manager 🐚

**Your SSH config on TUI-roids**

`ssm` is a tiny terminal UI that sits on top of your existing SSH config. No changes needed on your servers.

[Install](#install-30-seconds)

[![version][version-badge]](CHANGELOG.md)
[![license][license-badge]](LICENSE)
[![go report card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)

[version-badge]: https://img.shields.io/badge/version-2.0.0-blue.svg
[license-badge]: https://img.shields.io/badge/license-MIT-lue

## Demo

![demo](data/demo.png)

## Keys you'll actually use

| Key       | What it does                     |
|-----------|----------------------------------|
| `enter`   | Connect                          |
| `ctrl+e`  | Edit config live                 |
| `ctrl+r`  | Run command on host              |
| `ctrl+s`  | SFTP file browser                |
| `ctrl+v`  | Toggle config inspector          |
| `tab`     | Switch SSH ↔ MOSH                |
| `/`       | Fuzzy search                     |
| `q`       | Quit                             |

Full list in the app with `?`

## Install (30 seconds)

```bash
# macOS / Linux
curl -fsSL https://github.com/lfaoro/ssm/raw/main/scripts/get.sh | bash
# macOS quarantine workaround (I don't pay for a signing key)
xattr -d com.apple.quarantine /path/to/ssm
# or
brew install lfaoro/tap/ssm
# Arch (AUR)
yay -S ssm-bin
```

Other options: [Releases](https://github.com/lfaoro/ssm/releases) • Build from source

## Quick start

```bash
ssm                    # launch
ssm production         # filter by tag
ssm -se vpn            # show config + exit after connect
ssm --theme matrix     # green hacker vibes
ssm --theme sky        # soft blue vibes
# CTA: please add more themes
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

The more tags you use, the better it gets.

## Build / Contribute

Requires [Go](https://go.dev/doc/install).

```bash
git clone https://github.com/lfaoro/ssm.git \
  && cd ssm \
  && make build \
  && bin/ssm
```

or `go install github.com/lfaoro/ssm@latest`. PRs welcome.

**That's it.**

If you live in the terminal and manage more than a couple servers, this thing just makes life a little nicer.

Star it if it helps → https://github.com/lfaoro/ssm

Made with ❤️ and too much SSH pain.

## Shoutout

**[@hackerschoice](https://x.com/hackerschoice/status/1920899798837711279)** on X

If `ssm` actually made your life better:

- [GitHub Sponsors](https://github.com/sponsors/lfaoro)
- BTC: `bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs`
- XMR: `89XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c`
- FIAT: [Revolut](https://revolut.me/matrix)
- [message me on Telegram](https://t.me/leonarth)

or just ⭐ the repo. Appreciate it either way.
