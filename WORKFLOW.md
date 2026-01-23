# Workflow & Release Process

This document explains the automated release workflow and how to use this package in downstream projects.

## How the Workflow Works

### Automatic Versioning with Uplift

The workflow uses **Conventional Commits** to automatically create version tags:

```
Push to main with conventional commits
    ↓
Lint & Test jobs run
    ↓
Uplift analyzes commits since last version
    ↓
Creates version tag based on commit types
    ↓
GoReleaser creates GitHub release
    ↓
pkg.go.dev indexes new version automatically
```

### Commit Message Format

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

| Commit Type | Version Bump | Example |
|-------------|--------------|---------|
| `fix: ...` | Patch (v1.0.1) | `fix: handle nil pointer in profile update` |
| `feat: ...` | Minor (v1.1.0) | `feat: add analytics endpoint support` |
| `feat!: ...` or `BREAKING CHANGE:` | Major (v2.0.0) | `feat!: change client initialization API` |
| `chore:`, `docs:`, `test:` | No version | Ignored by uplift |

**Examples:**
```bash
# Patch release (v1.0.0 → v1.0.1)
git commit -m "fix: correct request timeout handling"

# Minor release (v1.0.1 → v1.1.0)
git commit -m "feat: add logs endpoint support"

# Major release (v1.1.0 → v2.0.0)
git commit -m "feat!: redesign client authentication

BREAKING CHANGE: NewClient now requires context parameter"
```

### Required GitHub Secrets

For uplift to trigger downstream workflows, you need a GitHub App:

1. Create a GitHub App with `contents: write` permission
2. Add these secrets to your repository:
   - `APP_ID` - GitHub App ID
   - `APP_PRIVATE_KEY` - GitHub App private key (PEM format)

**Why GitHub App?** Using `GITHUB_TOKEN` won't trigger tag-based workflows (GitHub security feature). A GitHub App token bypasses this limitation.

## Using This Package in Downstream Projects

### Installation

```bash
# Get the latest version
go get github.com/jacaudi/nextdns-go/nextdns@latest

# Get a specific version
go get github.com/jacaudi/nextdns-go/nextdns@v1.2.3
```

### In your go.mod

After running `go get`, your `go.mod` will include:

```go
module github.com/yourorg/yourproject

go 1.19

require (
    github.com/jacaudi/nextdns-go v1.2.3
)
```

### Import in your code

```go
package main

import (
    "context"
    "log"

    "github.com/jacaudi/nextdns-go/nextdns"
)

func main() {
    client, err := nextdns.New(
        nextdns.WithAPIKey("your-api-key"),
    )
    if err != nil {
        log.Fatal(err)
    }

    profiles, err := client.Profiles.List(context.Background(), &nextdns.ListProfileRequest{})
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Found %d profiles", len(profiles))
}
```

### Version Selection

Go modules use **Minimal Version Selection (MVS)**:

- `@latest` - highest tagged version (excluding pre-releases)
- `@v1.2.3` - specific version
- `@v1` - highest v1.x.x version
- `@v1.2` - highest v1.2.x version
- `@main` - latest commit on main (not recommended for production)

### Updating the dependency

```bash
# Update to latest version
go get -u github.com/jacaudi/nextdns-go/nextdns

# Update to specific version
go get github.com/jacaudi/nextdns-go/nextdns@v2.0.0
```

## Workflow Behavior

### On Feature Branch Push

```
Push to feature-branch → Lint + Test ✓
```

### On Pull Request

```
Open PR → Lint + Test ✓
```

### On Main Branch Push (Automatic Release)

```
Merge to main → Lint + Test → Version (tag created) → Release ✓
```

### On Tag Push (Manual - disabled by default)

Manual tagging is disabled in favor of automatic versioning. If needed, uncomment the `release-manual` job in `.github/workflows/release.yml`.

## Viewing Releases

- **GitHub Releases**: https://github.com/jacaudi/nextdns-go/releases
- **pkg.go.dev**: https://pkg.go.dev/github.com/jacaudi/nextdns-go/nextdns (auto-indexed)
- **Tags**: https://github.com/jacaudi/nextdns-go/tags

## Troubleshooting

### "No new version created"

Uplift didn't find any releasable commits (feat/fix). Check:
- Are you using conventional commit format?
- Have you already released these commits?

### "Workflow didn't trigger on tag"

If using `GITHUB_TOKEN` instead of GitHub App, tag creation won't trigger workflows. Use GitHub App credentials.

### Downstream package not updating

```bash
# Clear module cache and retry
go clean -modcache
go get -u github.com/jacaudi/nextdns-go/nextdns@latest
```

## Development Workflow

1. **Create feature branch**
   ```bash
   git checkout -b feat/new-endpoint
   ```

2. **Make changes and commit** (using conventional commits)
   ```bash
   git add .
   git commit -m "feat: add new endpoint for analytics"
   ```

3. **Push and create PR**
   ```bash
   git push origin feat/new-endpoint
   # Open PR on GitHub
   ```

4. **Merge to main** (after review)
   - Lint and test run automatically
   - Uplift analyzes commits and creates version tag
   - GoReleaser publishes release
   - pkg.go.dev indexes new version (~10 minutes)

5. **Use in downstream project**
   ```bash
   go get -u github.com/jacaudi/nextdns-go/nextdns@latest
   ```
