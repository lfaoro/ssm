# Contributing

## Requirements

- Code must pass `go vet ./...` and `go test -race -count=1 ./...`
- Code must be formatted with `go fmt ./...`
- All commits pass CI ([Go Tests](https://github.com/lfaoro/ssm/actions/workflows/go-tests.yml), [ShellCheck](https://github.com/lfaoro/ssm/actions/workflows/shellcheck.yml), [CodeQL](https://github.com/lfaoro/ssm/actions/workflows/codeql.yml))
- Follow [Effective Go](https://go.dev/doc/effective_go) and the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

## Workflow

1. Fork the repo
2. Create a branch: `git checkout -b feature/your-feature`
3. Make changes, run tests: `go test -race ./...`
4. Push and open a pull request against `main`

Pull requests are reviewed before merging. Keep changes focused and include test coverage for new functionality.
