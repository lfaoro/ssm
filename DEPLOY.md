# SSM Release Process (Manual Only)

We publish releases from a maintainer machine. There is no CI deployment.

## Prerequisites

- `GITHUB_TOKEN` secret with `contents:write` permission
- AUR SSH key loaded: `ssh-add ~/.ssh/aur_key`

## Release Steps

Run in this exact order:

1. `make check`                    # fmt + lint + tidy-check + test + build
2. Edit `CHANGELOG.md`             # required for any code change
3. `make tag TYPE=patch|minor|major`   # creates annotated tag and pushes it
4. `make release`                  # pre + nix-lock + goreleaser (local publish)
5. `make aur-push`                 # only if you want to update the Arch package

## What the Commands Do

| Command           | Effect                                      |
|-------------------|---------------------------------------------|
| `make tag`        | Calculates next semver, `git tag`, `git push --tags` |
| `make release`    | Runs goreleaser locally (builds + publishes to GitHub + Homebrew) |
| `make aur-push`   | Copies goreleaser AUR artifacts and pushes to AUR |

## Notes

- Tags use plain semver (`2.4.0`), not `v2.4.0`.
- `make release` is the only way we ship production binaries.

## References

- `AGENTS.md` – Release section
- `Makefile` – release, tag, aur-push targets
- `.config/goreleaser.yaml` – packaging configuration
