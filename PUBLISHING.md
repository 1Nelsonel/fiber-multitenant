# Publishing Guide

This guide walks you through publishing the Fiber Multitenant package to GitHub and making it available for the Go community.

## Prerequisites

- GitHub account
- Git installed locally
- Go 1.21+ installed

## Step 1: Prepare the Package

### 1.1 Update Module Path

Replace `github.com/1Nelsonel/fiber-multitenant` with your actual GitHub username in these files:

- `go.mod`
- `README.md`
- `examples/basic/main.go`
- `examples/chained/main.go`
- `examples/provisioning/main.go`

Example for username `johndoe`:
```bash
# In go.mod
module github.com/johndoe/fiber-multitenant

# In examples
import "github.com/johndoe/fiber-multitenant/middleware"
```

### 1.2 Test Locally

```bash
# Run tests if you've added any
go test ./...

# Try running examples
cd examples/basic
go mod init example-app
go mod edit -replace github.com/1Nelsonel/fiber-multitenant=../..
go mod tidy
go run main.go
```

## Step 2: Create GitHub Repository

### 2.1 Create Repository on GitHub

1. Go to https://github.com/new
2. Repository name: `fiber-multitenant`
3. Description: "Production-ready multitenancy for Go Fiber with PostgreSQL schema isolation"
4. Public repository
5. Don't initialize with README (we already have one)
6. Click "Create repository"

### 2.2 Initialize Local Git Repository

```bash
cd /path/to/fiber-multitenant
git init
git add .
git commit -m "Initial commit: Fiber multitenancy package

- Schema-based multitenancy with PostgreSQL
- Flexible tenant resolution (subdomain, header, path, query)
- Connection pooling and health checks
- Auto-migration support
- Complete documentation and examples"
```

### 2.3 Push to GitHub

```bash
git remote add origin https://github.com/1Nelsonel/fiber-multitenant.git
git branch -M main
git push -u origin main
```

## Step 3: Create Release

### 3.1 Tag Initial Version

```bash
git tag v1.0.0
git push origin v1.0.0
```

### 3.2 Create GitHub Release

1. Go to your repository on GitHub
2. Click "Releases" → "Create a new release"
3. Tag: `v1.0.0`
4. Release title: `v1.0.0 - Initial Release`
5. Description:

```markdown
# Fiber Multitenant v1.0.0

Initial release of production-ready multitenancy for Go Fiber with PostgreSQL.

## Features

✨ **Schema-based multitenancy** - Each tenant gets isolated PostgreSQL schema
🔄 **Connection pooling** - Cached tenant database connections with health checks
🌐 **Flexible tenant resolution** - Subdomain, header, path prefix, query param, or custom
🚀 **Auto-migration** - Automatic schema creation and model migration
🔒 **Secure isolation** - DSN-level `search_path` prevents cross-tenant data leaks
⚡ **Zero configuration** - Sensible defaults with full customization

## Installation

```bash
go get github.com/1Nelsonel/fiber-multitenant@v1.0.0
```

## Quick Start

See the [README](https://github.com/1Nelsonel/fiber-multitenant#readme) for documentation and examples.

## Examples

- [Basic Setup](./examples/basic)
- [Chained Resolvers](./examples/chained)
- [Tenant Provisioning](./examples/provisioning)
```

6. Click "Publish release"

## Step 4: Register with Go Module Proxy

The package is automatically indexed by `proxy.golang.org` when someone first downloads it. To trigger indexing immediately:

```bash
GOPROXY=proxy.golang.org go list -m github.com/1Nelsonel/fiber-multitenant@v1.0.0
```

## Step 5: Add Repository Topics

On GitHub repository page:

1. Click the gear icon next to "About"
2. Add topics:
   - `go`
   - `golang`
   - `fiber`
   - `multitenancy`
   - `multi-tenant`
   - `postgresql`
   - `gorm`
   - `saas`
   - `middleware`
   - `schema-isolation`

## Step 6: Documentation

### 6.1 Add Badges to README

Add these badges at the top of README.md:

```markdown
[![Go Reference](https://pkg.go.dev/badge/github.com/1Nelsonel/fiber-multitenant.svg)](https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant)
[![Go Report Card](https://goreportcard.com/badge/github.com/1Nelsonel/fiber-multitenant)](https://goreportcard.com/report/github.com/1Nelsonel/fiber-multitenant)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
```

### 6.2 Verify pkg.go.dev

After a few minutes, your package should appear on https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant

## Step 7: Share with Community

### 7.1 Reddit

Post to:
- r/golang
- r/golang_jobs (if hiring)

Example post:
```
Title: [Project] Fiber Multitenant - Production-ready multitenancy for Go Fiber

I built a multitenancy package for Go Fiber with PostgreSQL schema isolation.

Features:
- Schema-based multitenancy
- Multiple tenant resolution strategies
- Auto-migration and connection pooling
- Complete examples and documentation

GitHub: https://github.com/1Nelsonel/fiber-multitenant

Would love feedback from the community!
```

### 7.2 Social Media

- Twitter/X with hashtags: #golang #gofiber #saas
- LinkedIn
- Dev.to article

### 7.3 Fiber Community

- Post in Fiber Discord: https://gofiber.io/discord
- Create discussion in Fiber repo

### 7.4 Awesome Lists

Submit PR to add your package:
- [awesome-go](https://github.com/avelino/awesome-go)
- [awesome-fiber](https://github.com/gofiber/awesome-fiber)

## Step 8: Maintenance

### 8.1 Issue Template

Create `.github/ISSUE_TEMPLATE/bug_report.md`:

```markdown
---
name: Bug report
about: Create a report to help us improve
---

**Describe the bug**
A clear description of what the bug is.

**To Reproduce**
Steps to reproduce the behavior.

**Expected behavior**
What you expected to happen.

**Environment:**
- Go version: [e.g. 1.21.5]
- Fiber version: [e.g. 2.52.0]
- PostgreSQL version: [e.g. 15.4]
- OS: [e.g. Ubuntu 22.04]

**Additional context**
Add any other context about the problem here.
```

### 8.2 Setup GitHub Actions

Create `.github/workflows/test.yml`:

```yaml
name: Tests

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest

    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: postgres
          POSTGRES_DB: test_db
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
        ports:
          - 5432:5432

    steps:
    - uses: actions/checkout@v3

    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'

    - name: Run tests
      run: go test -v ./...
      env:
        DATABASE_URL: postgres://postgres:postgres@localhost:5432/test_db?sslmode=disable

    - name: Run go vet
      run: go vet ./...
```

### 8.3 Versioning

Follow [Semantic Versioning](https://semver.org/):

- **v1.0.x** - Bug fixes
- **v1.x.0** - New features (backwards compatible)
- **vx.0.0** - Breaking changes

## Example: Publishing Updates

```bash
# Make changes
git add .
git commit -m "feat: Add custom context key support"

# Create new tag
git tag v1.1.0

# Push changes and tag
git push origin main
git push origin v1.1.0

# Create release on GitHub
```

## Tips for Success

1. **Good Documentation** - Clear README with examples
2. **Examples** - Working code examples users can run
3. **Tests** - Add unit tests (shows reliability)
4. **Responsive** - Reply to issues promptly
5. **Changelog** - Keep CHANGELOG.md updated
6. **Semantic Versioning** - Follow semver strictly
7. **License** - Include MIT license (most permissive)
8. **Code of Conduct** - Optional but recommended

## Getting Help

If you need help publishing or have questions:

1. Check [Go Modules documentation](https://go.dev/doc/modules/publishing)
2. Ask in [Go Forum](https://forum.golangbridge.org/)
3. Check [Fiber Discord](https://gofiber.io/discord)

## Monitoring Success

Track your package adoption:

- **Stars** on GitHub
- **Downloads** via pkg.go.dev
- **Issues/PRs** from community
- **Dependents** shown on GitHub

Good luck with your open-source contribution! 🚀
