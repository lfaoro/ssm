# Contributing

## Requirements

- Code must pass `make check` (gofmt + lint + go-mod-tidy-check + test + build)
- All commits pass CI ([Go Tests](https://github.com/lfaoro/ssm/actions/workflows/go-tests.yml), [Lint](https://github.com/lfaoro/ssm/actions/workflows/lint.yml), [ShellCheck](https://github.com/lfaoro/ssm/actions/workflows/shellcheck.yml), [CodeQL](https://github.com/lfaoro/ssm/actions/workflows/codeql.yml))
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Important Rules

**Never commit changes unless explicitly instructed.**

This project uses a strict commit policy (enforced for both humans and agents):

- Prepare your changes
- Stage them
- Show the plan (`git status`, `git diff --cached`)
- Wait for the maintainer (or user) to say **"commit"** or **"go ahead"**

See [AGENTS.md](AGENTS.md) for the full rule and rationale.

## Dev Environment

Pick one:

### mise (recommended)

```bash
mise install          # install Go + toolchain from .mise.toml
mise trust            # trust the project config
```

### Nix

```bash
nix develop            # enter shell with Go + tools from flake.nix
```

## Common Commands

| Command              | Purpose                                      |
|----------------------|----------------------------------------------|
| `make check`         | Full pre-commit check (fmt + lint + test + build) — **run this before every PR** |
| `make test`          | Run tests with race detector                 |
| `make lint`          | Run golangci-lint (or fallback to go fmt + vet) |
| `make bench`         | Run all benchmarks                           |
| `make release-dev`   | Build a local goreleaser snapshot (dry run)  |

## Workflow

1. Fork the repo
2. Create a branch: `git checkout -b feature/your-feature`
3. Make changes
4. Run `make check` and fix any issues
5. Stage your changes and show the diff
6. Wait for explicit approval before committing
7. Push and open a pull request against `main`

Pull requests are reviewed before merging. Keep changes focused and include test coverage for new functionality.

## Releasing

Releases are performed manually (there is no CI-driven publish step).

See [DEPLOY.md](DEPLOY.md) for the exact release process, including:
- How to cut a new version
- What `make tag` and `make release` actually do
- How to publish to AUR, Homebrew, etc.

## Architecture & Rules Reference

For deeper context (especially if you're an agent or doing significant work):

- [AGENTS.md](AGENTS.md) — full project rules, Bubbletea v2 constraints, security notes, benchmarking, release checklist
- [DEPLOY.md](DEPLOY.md) — release mechanics

## Questions?

Open an issue or discussion. We're happy to help.
