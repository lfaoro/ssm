# SSM | Secure Shell Manager

SSM is a TUI application which helps you manage your SSH config, filter through hosts and easily connect via SSH or MOSH.

[![MIT license](https://img.shields.io/badge/license-MIT-blue)](license)
[![Go Report Card](https://goreportcard.com/badge/github.com/lfaoro/ssm)](https://goreportcard.com/report/github.com/lfaoro/ssm)

## Features
- vim keys navigation: jkhl, ctrl+d/u, g/G
- autoreload ssh config on change
- filter through all your servers
- easily connect and return to app
- switch between ssh and mosh with a tab
- create free root servers

## Key-binds
```
enter↵       connect to selected host
ctrl+e       edit ssh configs
q            quit
/            filter hosts
tab          switch between ssh/mosh

# not yet implemented binds
ctrl+r       run command on the server without starting a pty (coming soon)
ctrl+s       sftp upload/download files to/from server (coming soon)
space␣       select multiple hosts to run a command or sftp to/from (coming soon)
```

## Install

```bash
# bash one-liner (linux/macos)
curl -sSL https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash
wget -qO- https://raw.githubusercontent.com/lfaoro/ssm/refs/heads/main/scripts/get.sh | bash

# brew (macos)
brew install lfaoro/tap/ssm

# go install
go install github.com/lfaoro/ssm@latest
```

## Build

> requires [Go](https://go.dev/doc/install)

```bash
git clone https://github.com/lfaoro/ssm.git \
  && cd ssm \
  && make build \
  && bin/ssm
```

## Road map
- [ ] add run command on host
- [ ] add multiple hosts selection
- [ ] add run commands on multiple hosts asynchronously
- [ ] add sftp with interactive files selector
- [ ] add sftp to multiple hosts async

## Contributing

I love pull&feature requests, don't hesitate. No rules, we discuss in the issue.

## Help ssm grow

> If ssm is useful to you, consider being useful to ssm.

- **star the repo**
- **tell your friends**

- [GitHub sponsor](https://github.com/sponsors/lfaoro)
- [FIAT support](https://checkout.revolut.com/pay/1122870b-1836-42e7-942b-90a99ef5e457)
- [XMR support](https://xmrchain.net/search?value=9XCyahmZiQgcVwjrSZTcJepPqCxZgMqwbABvzPKVpzC7gi8URDme8H6UThpCqX69y5i1aA81AKq57Wynjovy7g4K9MeY5c)
- [BTC support](https://mempool.space/address/bc1qzaqeqwklaq86uz8h2lww87qwfpnyh9fveyh3hs)

## License
[MIT license](license.md)
