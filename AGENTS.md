# SSM — Agent Working Instructions

This document exists to make AI coding agents (and humans) effective and safe when modifying the codebase. It prioritizes **actionable rules, invariants, and verification steps** over descriptive snapshots of the current state.

## The Prime Directive

**NEVER commit changes unless explicitly instructed.**

Prepare the staging area (`git add`), show the plan and the full diff, and wait for the user to explicitly say "commit" or "go ahead". This rule overrides all other instructions. Do not ask the user whether they want to commit.

## 1. Verification — Run These Before Claiming Work Is Complete

An agent must produce clean results from the project's canonical checks before presenting changes as done.

**Primary command (use this most often):**
```bash
make check
```

`make check` runs (see Makefile):
- `gofmt`
- `lint` (golangci-lint if present, else `go fmt + go vet`)
- `go-mod-tidy-check`
- `make test` (`go test -race -count=1 ./...`)
- `make build`

Individual commands you will use constantly:
```bash
go build ./...
go vet ./...
go test -race -count=1 ./...
golangci-lint run ./...
govulncheck ./...
make build-static
```

When touching performance-sensitive code, also consider `make bench` (and the `-cpu` / `-mem` variants).

**Rule:** If you have not run the relevant verification commands and shown the output, you are not finished.

## 2. Non-Negotiable Rules

- **Changelog discipline**: Any commit that changes code, fixes bugs, adds features, or adjusts tests **must** also update `CHANGELOG.md` with a concise entry under the appropriate keepachangelog.com section (Security, Fix, Add, Refactor, Test, Docs, etc.).
- **Injection prevention**: The `--` delimiter must appear before every hostname in all SSH, mosh, and `syscall.Exec` invocations. This is non-negotiable.
- **SFTP connection flags**: SFTP and certain remote operations deliberately use `BatchMode=yes`, `RequestTTY=no`, and `StrictHostKeyChecking=no`. These are intentional (users are connecting to their own servers). See SECURITY.md.
- **Sensitive data handling**: `IdentityFile`, `ProxyCommand`, `CertificateFile`, and similar keys must remain filtered from any configuration display/viewport.
- **Parser thread-safety contract**: See Architecture Contracts below. Do not regress the brief-publish-lock model.

## 3. Architecture Contracts (Must Remain True)

### SSH Config Parser (`pkg/sshconf`)
- `Parse()` and `ParsePath()` do **all** I/O, glob expansion, and recursive Include handling **without** holding the write lock.
- A very short critical section at the very end (on success path only) publishes the results atomically under `Lock()`.
- All readers (`GetHosts`, `GetHost`, `GetParamFor`, `GetPath`) use `RLock` and must observe a fully consistent snapshot — either the old complete state or the new complete state. Partial state must never be visible.
- `SetOrder()` must acquire the write lock.
- Include recursion is limited to depth 10 with cycle detection via the `visited` map passed through recursive calls.
- The `secondaryHosts` field no longer exists in the struct (it was an internal implementation detail removed during the lock redesign).
- `TestParseConcurrentReaders` (and the rest of `parser_test.go`) must continue to pass under `-race`.

### TUI & Sub-models (`pkg/tui`)
- The application uses Bubbletea v2 with `tea.View` struct returns (never raw strings from `View()`).
- Complex features (run command, SFTP browser, ping) are implemented as sub-models that communicate with the root model exclusively via messages.
- In-flight external processes are tracked with dedicated mutexes (`currentCmdMu` in runcmd, `connectMutex` in sftp) to allow safe cancellation on exit.
- Ping concurrency is controlled via `pingWorkerCount()` (CPU-based with hard lower/upper bounds) + semaphore. The mechanism must remain bounded.

### Security & Hardening (see also SECURITY.md)
- No software is ever installed on remote hosts.
- All remote command output is sanitized (ANSI stripped, stderr truncated).
- SSH config file permissions are checked on load; warnings are emitted for non-0600 modes.
- Debug logs are only active when the `--debug` flag is used.
- In-flight remote processes have explicit tracking + cancellation paths.

## 4. Bubbletea v2 + Charm Specifics

- Primary modules: `charm.land/bubbletea/v2`, `charm.land/bubbles/v2`, `charm.land/lipgloss/v2`.
- The charm.land/v2 packages manage their own transitive dependencies on `github.com/charmbracelet/x/ansi` and `colorprofile`. Do not introduce replaces or pins targeting the old `github.com/charmbracelet` paths unless you have a clear, documented reason.
- Important API differences from older Bubbletea (common source of agent errors):
  - `tea.KeyPressMsg` instead of `tea.KeyMsg`
  - `list.SetFilterText()` instead of `SetFilterValue`
  - `viewport.GetContent()` instead of `.Content()`
  - `tea.View` struct cannot be compared to `nil`

## 5. Linting, Suppressions, and Justifications

- `.golangci.yml` enables 18 linters (see the file). `gocyclo` minimum complexity is intentionally high (55).
- Justified `//nolint` directives that must be preserved when touching the relevant code:
  - `gosec` on `exec.Command*` / file write operations (this is an SSH TUI that legitimately spawns processes and writes to the user's `~/.ssh` area).
  - `nilerr` at `pkg/sshconf/util.go:15` (deliberate fallback to `/etc/ssh/ssh_config` when `$HOME` cannot be determined).
- Categories that are intentionally suppressed project-wide (acceptable): `errorlint`, `forcetypeassert`, `goconst`, `gocritic`, `godot`.

When adding a new `//nolint`, add a clear comment explaining why and consider whether the code should be restructured instead.

## 6. Testing Guidelines

- Prefer table-driven tests with `t.Run()` subtests and standard library assertions.
- `-race` is mandatory (`make test` enforces it).
- Tests that require external commands (ssh, cloud provider credentials, etc.) may be skipped when those commands are unavailable. Do not make the suite flaky.
- The goal is high coverage through meaningful tests, especially around the parser (recursion, cycles, globs, thread-safety, permission warnings) and TUI sub-models.
- **Do not** embed exact current test counts or per-package coverage percentages in this file or in code. They become wrong the moment anyone adds a test. Use them locally during development if helpful; do not commit them as documentation.

## 7. Release & Distribution Mechanics

- Tagging and full releases are driven by `make release`, `make release-prod`, `make tag TYPE=...`.
- goreleaser config lives in `.config/goreleaser.yaml` (multi-platform static binaries + many package formats).
- AUR publishing requires a loaded ssh-agent key (`ssh-add ~/.ssh/aur_key`) and uses the dedicated `scripts/aur-push.sh`.
- `make stats` updates the download badge data in `data/stats.json`.
- `make nix-lock` is required as part of the release process.

Any change to release-related files or scripts must be accompanied by a CHANGELOG entry.

## 8. Environment & Tooling

- Recommended: `mise` (see `.mise.toml` for exact Go + golangci-lint + goreleaser versions).
- Alternative: `nix develop` (flake.nix).
- Do not casually change the pinned tool versions in `.mise.toml` or `go.mod` without understanding the impact on reproducible builds and CI.

## 9. Non-Obvious Decisions & Historical Gotchas

- **Parser lock redesign (2026)**: The write lock used to be held for the entire parse (including all recursive Includes). This was changed to a "do the work, publish briefly at the end" model for both performance and to give stronger consistency guarantees to concurrent readers from the TUI. The new `TestParseConcurrentReaders` exists to protect this contract.
- `secondaryHosts` field removal: Was only ever an internal detail for TagOrder mode during parsing. After the lock redesign it became dead code and was deleted.
- `StrictHostKeyChecking=no` for SFTP: Intentional and documented. Users of this tool are connecting to machines they already manage.
- High `gocyclo` threshold: Some of the TUI message handling and the parser naturally have high cyclomatic complexity. The team prefers readable code over forcing artificial function splits in those areas.
- The hidden `ssm test` and `ssm generate` subcommands: Legacy placeholders. Treat them as internal and do not build user-facing features on them.

When you encounter something that looks odd, check this section and the git history before "fixing" it.

---

**Remember the Prime Directive.** When in doubt, run `make check`, show the diff, and ask.
