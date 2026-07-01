# Contributing

Thanks for your interest in improving `nextdns-go`! This is a short pointer to the conventions already documented in the repo.

## Development workflow

The full workflow and release process — branching, conventional commits, automatic versioning, and downstream usage — is documented in **[docs/workflow.md](docs/workflow.md)**. Start there.

Common tasks are wired into [`taskfile.yml`](taskfile.yml) (install [Task](https://taskfile.dev), or `brew install go-task`; run `task --list` for the full set):

```bash
task deps      # install local dev tools (govulncheck, tparse)
task test      # go test ./...
task tparse    # race detector + coverage, pretty-printed
task lint      # golangci-lint run + govulncheck
task vet       # go vet ./...
task tidy      # go mod tidy
```

Run `task lint`, `task vet`, and `task test` clean before pushing.

## Commit messages

Releases are automated with semantic-release, so commits must follow [Conventional Commits](https://www.conventionalcommits.org/):

| Commit Type | Version Bump |
|-------------|--------------|
| `fix: ...` | Patch (v1.0.1) |
| `feat: ...` | Minor (v1.1.0) |
| `feat!: ...` / `BREAKING CHANGE:` | Major (v2.0.0) |
| `chore:`, `docs:`, `test:` | No release |

See [docs/workflow.md](docs/workflow.md#commit-message-format) for examples.

## Pull requests

1. Create a topic branch (e.g. `feat/new-endpoint`).
2. Make changes and commit using conventional commits.
3. Push and open a PR against `main`; lint and test run automatically.
4. After merge, semantic-release analyzes the commits, tags a version, and pkg.go.dev indexes it.

Include tests for behavior changes and keep changes focused.
