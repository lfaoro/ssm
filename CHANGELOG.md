# [Unreleased]

## Add
- Cloud provider sync: `ssm sync [hetzner aws gcp azure]` CLI subcommand and `Ctrl+y` TUI
  panel to discover running servers and write them to `~/.ssh/config.d/50-ssm-{provider}`
  - Hetzner: uses `HCLOUD_TOKEN` env var
  - AWS: uses standard SDK credential chain (all regions)
  - GCP: uses `GCP_PROJECT` env var
  - Azure: uses `AZURE_SUBSCRIPTION_ID` env var + Azure SDK chain
  - `--user` and `--key` flags for default SSH user and IdentityFile
  - `--dry-run` / `-n` to preview generated config
  - Hosts named `{region}-{name}`, tagged with `#tag: {provider}` for filtering
  - Auto-injects `Include config.d/*` into `~/.ssh/config` if missing
  - CLI output shows per-provider file paths
- Tests: syncer unit tests (config generation, sanitize, file writing, include injection),
  provider interface compliance, TUI sync sub-model tests

## Fix
- `--ping` with tag filter: only ping visible/filtered hosts instead of all hosts
- `--ping` applied before tag filter: send `FilterTagMsg` before `LivenessCheckMsg` in startup sequence

# [2.3.1] May 25, 2026

## Fix
- Mosh connection failure on macOS: replace fragile `--ssh=` quoting with `SSH_CONFIG` env var (closes #45)

# [2.3.0] May 22, 2026

## Fix
- Filtering bug after ping cleared host list
- Fix changelog versions (2.1.2 renamed to 2.2.0, add missing 2.2.1 section)

## Add
- `aur-push` target and script for AUR package publishing
- Nix package output via flake.nix (`nix run`, `nix profile install`)
- `nix-lock` target: regenerate flake.lock, auto-staged for release
- `gofmt` in `make check` target for formatting consistency

## Docs
- Nix run and profile install in quick-start section
- Fix Nix table entries in readme
- Add Ailton Baúque to AUTHORS

## Build
- Bump `actions/checkout` from 4 to 6
- Bump `actions/setup-go` from 5 to 6
- Bump `golangci/golangci-lint-action` from 8 to 9
- Bump `dependabot/fetch-metadata` from 2 to 3
- Bump `github/codeql-action` from 4.35.4 to 4.35.5
- Disable AUR upload in goreleaser (`skip_upload: true`), defer to `aur-push.sh`

# [2.2.1] May 19, 2026

## Add
- Per-theme terminal background: solid dark via `tea.View.BackgroundColor`

# [2.2.0] May 18, 2026

## Security
- Add `--` delimiter before hostname in SFTP connection (anti-injection, matches SSH/mosh)
- Shell-escape mosh `--ssh` config path (prevents metacharacter injection)
- `filepath.Clean` on Include paths to prevent directory traversal

## Add
- 9 new linters: `misspell`, `unconvert`, `bodyclose`, `noctx`, `nilnil`, `prealloc`, `dupword`, `intrange`, `perfsprint`
- `make check` target: lint + go-mod-tidy-check + test + build (pre-commit)
- `go-mod-tidy-check` target: verify go.mod/go.sum consistency via `go mod tidy -diff`
- Mise and Nix dev environments: `.mise.toml` (Go + toolchain), `flake.nix`, `.envrc` (direnv)
- Nix: package via flake.nix (`nix run github:lfaoro/ssm`, `nix profile install github:lfaoro/ssm`)
- SFTP: Esc clears active pane selections, second Esc/q exits to main
- SFTP: auto-deselect after batch transfer
- README: feature list section

## Fix
- SftpModel: return base on bad cast instead of crashing with panic
- SftpModel: bounds-check `GlobalIndex()` before indexing hosts, use `GetHosts()` API
- SFTP: use remote `Getwd()` root directory instead of hardcoded `/tmp`
- `cmdModel.currentCmd` data race fixed with `sync.Mutex`
- Elm architecture: `vp.Style` moved from `View()` to `syncViewportStyle()` in `Update()`
- Zero-value theme on initial render — `skyTheme()` applied before `listFrom()`
- Version check goroutine guarded with `atomic.Bool` shutdown flag
- SFTP: `handleEnter` remote directory nil `sftpClient` crash guard
- SFTP: `batchTransfer` nil `sftpClient` crash guard
- SFTP: per-pane selection maps prevent cross-pane clearing
- SFTP: `close()` no longer calls `Process.Wait()` under `connectMutex`
- SFTP: `sftpConnectMsg` error path clears stale `s.connecting` reference

## Refactor
- `sync.Mutex` → `sync.RWMutex` in SSH config parser (concurrent reads no longer block)
- Loop-invariant `c.order == TagOrder` check hoisted out of scanner loop
- `fileExists()` simplified — removed dead nested return
- `%v` → `%w` in `latestTag` and `connect` error wrapping
- `resolvePingTarget` explicit return (no naked return)
- Shadowed `err` → `statErr` in parser permission check
- `pingAllCmd` capped at 50 concurrent TCP dials via semaphore
- `exec.Command` → `exec.CommandContext(context.Background())` — all packages
- `fmt.Errorf` → `errors.New` for constant-string calls (perfsprint)
- `fmt.Sprintf` single placeholder → string concatenation (perfsprint)
- SFTP: removed unimplemented `handleMkdir` handler, `modeMkdir` constant, `m` keybinding

## Test
- Parser: include depth limit exceeded and within limit
- Parser: cyclic include detection (A→B→A)
- Parser: include glob pattern expansion
- Parser: include relative path resolution
- Parser: path traversal with `filepath.Clean` (`sub/../sub/safe`)
- Parser: `Parse()` default config path via `$HOME/.ssh/config`

# [2.1.1] May 18, 2026

## Fix
- Preserve filter state after ping — `p`/`P` no longer clears filtered host list

# [2.1.0] May 18, 2026

## Add
- `p`/`P` key bindings to ping servers via TCP dial to the SSH port
  - `p` pings the selected host, `P` (Shift+p) pings all hosts in parallel
  - Shows latency next to each host (e.g. `42ms`, `timeout`, `unreachable`)
  - No privileges needed — uses `net.Dialer` (same permissions as SSH)
- `--ping` CLI flag triggers liveness check on all hosts at startup

# [2.0.0] May 17, 2026

## Security
- Add `--` delimiter before hostname in SSH command to prevent flag injection
- Fix mosh `--ssh` flag quoting (remove single quotes that broke parsing)

## Fix
- Fix `tea.Batch` type mismatch in `ExecProcess` callback (returned `tea.Cmd` instead of `tea.Msg`)
- Replace magic number `1` with `list.Filtering` constant for filter state comparison
- Fix editor lookup to respect `EDITOR` env var (was overwritten by first editor found in PATH)
- Remove hardcoded black background color (broke light terminal themes)
- Fix tag order test threshold (8 tagged hosts)

## Refactor
- Move `sensitiveKeys` map and `isSensitiveKey` from `pkg/tui/model.go` to `pkg/sshconf/parser.go`
- Add exported `sshconf.RemoveComments()` and `sshconf.IsSensitiveKey()` wrappers for benchmarking
- Simplify debug log joining with `strings.Join` instead of manual loop
- Replace custom `contains`/`searchSubstring` with `strings.Contains` in tests

## Add
- SFTP file browser (`ctrl+s`) — dual-pane local ↔ remote file transfer using `github.com/pkg/sftp`
  - Upload/download with file size display and transfer history buffer (50 entries)
  - Overwrite confirmation dialog and delete operations on both panes
  - Symlink detection in directory listings
  - SSH batch mode (`StrictHostKeyChecking=no`, `BatchMode=yes`, `RequestTTY=no`)
  - Theme-consistent UI with solid background bar headers and adaptive colors
  - Remote pane starts in `/tmp` by default
- `make lint` target (golangci-lint if available, fallback to go fmt + go vet)
- `make bench`, `make bench-cpu`, `make bench-mem`, `make bench-compare` targets
- Benchmark suite for `sshconf` (ParsePath, GetHost, GetHosts, GetParamFor, RemoveComments, IsSensitiveKey)
- Benchmark suite for `tui` (setConfig, formatHost, sanitizeOutput, sanitizeStderr)

## Test
- Add comprehensive TUI test suite (132+ tests, 85%+ coverage)
- Add `testdata/test_config` fixture and shared test helpers
- Add SFTP test suite (`sftp_test.go`, 40+ tests covering file items, panes, transfers, confirmations, history limit)

# [1.0.2] May 15, 2026

## Fix
- Fix Escape key quitting app when filter not active
- Update readme keys table split `q` (quit) and `esc` (exit filter)
- Expand `data/config_example` with real-world host groups

# [1.0.1] May 14, 2026

## Release
- Reduce release matrix to 4 OSes × 2 arches (drop netbsd, solaris, 386, arm)
- Add AUR package support
- Add download stats badge to readme
- Add `make stats` target and `scripts/stats.go`
- Add copyright headers to all `.go` files

# [1.0.0] Apr 29, 2026

## Security Audit & Hardening
- audit all injection vectors, file access, secrets, concurrency, and logging
- filter sensitive SSH keys (identityfile, proxycommand, etc.) from config viewport
- add `--` delimiter before hostname in all SSH/mosh/syscall invocations
- resolve symlinks on config paths, add Include recursion depth limit + cycle detection
- check config file permissions, warn if not 0600
- strip ANSI escape sequences from remote command output
- sanitize SSH stderr before displaying (truncate at 500 chars)
- don't collect debug log entries when debug mode is disabled
- track version-check goroutine in WaitGroup to prevent leak
- lock all Config read paths (GetHost, GetParamFor, GetHosts)
- migrate google/go-github v17 → v69 (8yr stale dep)
- upgrade Go to 1.26.2 (resolves 7 stdlib CVEs)

## All 0.4.x changes
- migrate charmbracelet bubbletea/bubbles/lipgloss from github.com to charm.land
- remove segfault.net hardcoded password and special-case logic
- add table-driven tests: parser, log, themes
- replace O(n²) list insertion with O(n) SetItems
- add CI workflow: go vet, go test -race, go build
- update dependencies, fix bugs, remove dead code

# [0.4.2] Apr 29, 2026
- migrate charmbracelet bubbletea/bubbles/lipgloss from github.com to charm.land
- upgrade View() from string to tea.View, rewrite Init() commands
- update Go to 1.26, upgrade all dependencies to latest
- add ci workflow: go vet, go test -race, go build
- add AGENTS.md

# [0.4.1] Apr 29, 2026
- remove segfault.net hardcoded password and special-case logic
- fix relative path fallback in config discovery
- fix inverted debug view (messages now shown when debug active)
- remove dead code: state.go, run.go, keywords.go, styles.go
- clean up commented-out code blocks
- add table-driven tests: parser (9 subtests), log (11 subtests), themes (3)
- replace O(n²) list insertion with O(n) SetItems
- add 5s context timeout to GitHub API version check
- only fire tick loop in debug mode
- fix runcmd window resize to account for bar height
- update Go to 1.26, upgrade dependencies

# [0.4.0] Jul 29, 2025
- add run command feature (ctrl+r)
- add themes `--theme matrix` editable from themes.go

# [0.3.5] May 14, 2025
- add use ENV variables to configure FLAGS
- fix bug causing high cpu usage

# [0.3.4] May 9, 2025
- resize list dynamically when error
- add ctrl+c to quit the app
- remove segfault auto add

# [0.3.3] May 5, 2025
- add #tagorder key to show `#tag` hosts first
- add `--order` flag to show `#tag` hosts first
- add clear filter using `backspace`
- fix could not resolve hostname bug
- fix show Host when HostName is missing
- refactor ssh config parser
- add when EDITOR not set search for vim,vi,nano,ed

# [0.3.2] April 30, 2025
- fix error/log message alignment
- add emacs keys: ctrl+p/n/b/f(up/down/left/right)
- update libraries

# [0.3.1] April 30, 2025
- fix exithost invalid character ssh error
- exit filtering on enter key if prompt is empty
- remove watcher from parser
- skip parsing wildcard(*) hosts

# [0.3.0] April 30, 2025
- add cursor while filtering
- restructure codebase

# [0.2.2] April 29, 2025
- return error when EDITOR env is not set
- add version check
- inform via msg when new version is released
- fix custom config path
- improve codebase

# [0.2.1] April 25, 2025
- fix crash on segfault
- remove windows release
- upgrade deps
- general improvements

# [0.2.0] April 24, 2025
- add exit flag `--exit / -e`: ssm will exit after connecting to a host
- add `ctrl+v`: view full config for selected host
- add ordered map for config options
- fix filtering hosts
- improve cli helpfile
- improve readme

# [0.1.2] - April 21, 2025
- fix parsing of tag keys

# [0.1.1] - April 21, 2025
- fix parsing comments on same line as config keys
- move segfault free server at the bottom
- resolve absolute path from custom --config
- add help section to readme
- add ssh config example in data/config_example

# [0.1.0] - April 20, 2025
- extend pkg/sshconf to support #tag: keys e.g. #tag: admin,vpn
- add arg for tags e.g. `ssm admin` will show only admin tagged hosts
- add `--config, -c` flag to provide custom config location other than default search paths

# [0.0.1] - April 18, 2025
- initial release
- pkg/sshconf: parse, watch logic 
- pkg/tui: bubbletea UI implementation
- main.go: initilization logic, args & flags handling

[0.0.1]: https://github.com/lfaoro/ssm/releases/tag/0.0.1
[0.1.0]: https://github.com/lfaoro/ssm/compare/0.0.1...0.1.0
[0.1.1]: https://github.com/lfaoro/ssm/compare/0.1.0...0.1.1
[0.1.2]: https://github.com/lfaoro/ssm/compare/0.1.1...0.1.2
[0.2.0]: https://github.com/lfaoro/ssm/compare/0.1.2...0.2.0
[0.2.1]: https://github.com/lfaoro/ssm/compare/0.2.0...0.2.1
[0.2.2]: https://github.com/lfaoro/ssm/compare/0.2.1...0.2.2
[0.3.0]: https://github.com/lfaoro/ssm/compare/0.2.2...0.3.0
[0.3.1]: https://github.com/lfaoro/ssm/compare/0.3.0...0.3.1
[0.3.2]: https://github.com/lfaoro/ssm/compare/0.3.1...0.3.2
[0.3.3]: https://github.com/lfaoro/ssm/compare/0.3.2...0.3.3
[0.3.4]: https://github.com/lfaoro/ssm/compare/0.3.3...0.3.4
[0.3.5]: https://github.com/lfaoro/ssm/compare/0.3.4...0.3.5
[0.4.0]: https://github.com/lfaoro/ssm/compare/0.3.5...0.4.0
[0.4.1]: https://github.com/lfaoro/ssm/compare/0.4.0...0.4.1
[0.4.2]: https://github.com/lfaoro/ssm/compare/0.4.1...0.4.2
[1.0.0]: https://github.com/lfaoro/ssm/compare/0.4.2...1.0.0
[1.0.1]: https://github.com/lfaoro/ssm/compare/1.0.0...1.0.1
[1.0.2]: https://github.com/lfaoro/ssm/compare/1.0.1...1.0.2
[2.0.0]: https://github.com/lfaoro/ssm/compare/1.0.2...2.0.0
[Unreleased]: https://github.com/lfaoro/ssm/compare/2.0.0...HEAD
