# Contributing

## Requirements

- Code must pass `make check` (lint + go-mod-tidy-check + test + build)
- All commits pass CI ([Go Tests](https://github.com/lfaoro/ssm/actions/workflows/go-tests.yml), [ShellCheck](https://github.com/lfaoro/ssm/actions/workflows/shellcheck.yml), [CodeQL](https://github.com/lfaoro/ssm/actions/workflows/codeql.yml))
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

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

### direnv (auto-activate)

With either method above, direnv activates the environment automatically when you `cd` into the project:

```bash
direnv allow           # one-time trust of .envrc
```

## Workflow

1. Fork the repo
2. Create a branch: `git checkout -b feature/your-feature`
3. Make changes, run checks: `make check`
4. Push and open a pull request against `main`

Pull requests are reviewed before merging. Keep changes focused and include test coverage for new functionality.
