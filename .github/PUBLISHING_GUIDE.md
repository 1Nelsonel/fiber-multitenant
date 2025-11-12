# Publishing Guide - Making Your Package Public

This guide walks you through publishing your Go package to the community.

## ✅ Pre-Publishing Checklist

Before publishing, ensure:
- [x] All tests passing (22/22 ✅)
- [x] CI/CD pipeline working ✅
- [x] Documentation complete ✅
- [x] Security scan passing ✅
- [x] Linting passing ✅
- [ ] Repository is public
- [ ] Branch protection enabled
- [ ] License file exists (MIT recommended)
- [ ] First release tag created

## Step 1: Make Repository Public

Your repository must be public for others to use it.

### Check Current Status

1. Go to: https://github.com/1Nelsonel/fiber-multitenant/settings
2. Scroll to the bottom to "Danger Zone"
3. Check if it shows "Change repository visibility"

### If Repository is Private:

1. Click **"Change visibility"**
2. Select **"Make public"**
3. Type the repository name to confirm: `fiber-multitenant`
4. Click **"I understand, make this repository public"**

**Note**: Once public, anyone can see and use your code.

## Step 2: Add a License (Required)

Go packages should have a license. MIT is the most common for open source.

### Add MIT License:

1. Go to: https://github.com/1Nelsonel/fiber-multitenant
2. Click **"Add file"** > **"Create new file"**
3. Name it: `LICENSE`
4. Click **"Choose a license template"**
5. Select **"MIT License"**
6. Fill in your name: `Nelson El`
7. Click **"Review and submit"**
8. Commit the file

Or use the command line (we'll do this):

```bash
# Create LICENSE file
cat > LICENSE << 'EOF'
MIT License

Copyright (c) 2025 Nelson El

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

git add LICENSE
git commit -m "Add MIT license"
git push origin main
```

## Step 3: Create Your First Release (v1.0.0)

This is the most important step - it makes your package available via `go get`.

### Using Git Tags (Recommended):

```bash
# Make sure you're on main and up to date
git checkout main
git pull origin main

# Create an annotated tag for v1.0.0
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release

Features:
- Schema-based multitenancy for PostgreSQL
- Connection pooling and caching
- Multiple tenant resolution strategies (subdomain, header, path, query)
- Auto-migration support
- Comprehensive test coverage
- Production-ready with CI/CD

Full documentation: https://github.com/1Nelsonel/fiber-multitenant"

# Push the tag to GitHub
git push origin v1.0.0
```

### What This Does:

1. Creates a release tag `v1.0.0`
2. GitHub Actions automatically creates a release
3. Go's module proxy picks it up
4. Users can install with: `go get github.com/1Nelsonel/fiber-multitenant@v1.0.0`

### Verify Release Created:

- Check: https://github.com/1Nelsonel/fiber-multitenant/releases
- You should see "Release v1.0.0" with auto-generated notes

## Step 4: Submit to pkg.go.dev

Go's official package documentation site (pkg.go.dev) indexes public packages automatically.

### Automatic Indexing:

Once you push your v1.0.0 tag, pkg.go.dev will automatically index your package within 15-30 minutes.

### Manual Trigger (Optional):

1. Go to: https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant
2. If it shows "Request 'github.com/1Nelsonel/fiber-multitenant'", click it
3. This triggers immediate indexing

### Add Package Documentation:

Add package-level documentation to your main files:

```go
// Package middleware provides Fiber middleware for multitenancy support.
//
// This package enables schema-based multitenancy in Go Fiber applications
// using PostgreSQL. It provides automatic tenant resolution, connection
// pooling, and seamless integration with GORM.
//
// Example usage:
//
//	import (
//		"github.com/1Nelsonel/fiber-multitenant/middleware"
//		"github.com/1Nelsonel/fiber-multitenant/tenantstore"
//		"github.com/gofiber/fiber/v2"
//	)
//
//	func main() {
//		app := fiber.New()
//
//		// Configure tenant store
//		config := tenantstore.DefaultConfig("postgres://...")
//		store, _ := tenantstore.New(config)
//
//		// Add middleware
//		app.Use(middleware.New(middleware.Config{
//			Store: store,
//		}))
//
//		// Your routes here
//		app.Listen(":3000")
//	}
package middleware
```

## Step 5: Enhance README with Badges

Add status badges to show package quality at a glance.

### Badges to Add:

```markdown
# Fiber Multitenant

[![Go Version](https://img.shields.io/github/go-mod/go-version/1Nelsonel/fiber-multitenant)](https://go.dev/)
[![Go Report Card](https://goreportcard.com/badge/github.com/1Nelsonel/fiber-multitenant)](https://goreportcard.com/report/github.com/1Nelsonel/fiber-multitenant)
[![CI](https://github.com/1Nelsonel/fiber-multitenant/workflows/CI/badge.svg)](https://github.com/1Nelsonel/fiber-multitenant/actions)
[![GoDoc](https://pkg.go.dev/badge/github.com/1Nelsonel/fiber-multitenant)](https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Release](https://img.shields.io/github/v/release/1Nelsonel/fiber-multitenant)](https://github.com/1Nelsonel/fiber-multitenant/releases)

A production-ready multitenancy solution for Go Fiber using PostgreSQL schema isolation.
```

## Step 6: Add GitHub Topics

Topics make your package discoverable on GitHub.

1. Go to: https://github.com/1Nelsonel/fiber-multitenant
2. Click the gear icon ⚙️ next to "About"
3. Add these topics:
   - `go`
   - `golang`
   - `fiber`
   - `multitenancy`
   - `multitenant`
   - `postgresql`
   - `postgres`
   - `saas`
   - `middleware`
   - `gorm`
   - `fiber-middleware`
   - `schema-isolation`

4. Add a description: "Production-ready multitenancy solution for Go Fiber with PostgreSQL schema isolation"
5. Add website: `https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant`
6. Click **"Save changes"**

## Step 7: Announce Your Package

### Share on:

1. **Reddit**:
   - r/golang - "Show HN: fiber-multitenant - A multitenancy package for Go Fiber"
   - r/webdev

2. **Twitter/X**:
   - Tweet about your package with hashtags: #golang #gofiber #opensource

3. **Dev.to**:
   - Write a blog post: "Building a Multitenancy Package for Go Fiber"

4. **Golang Weekly**:
   - Submit to: https://golangweekly.com/

5. **Fiber Discord/Community**:
   - Share in Fiber's community channels

### Announcement Template:

```
🚀 Introducing fiber-multitenant v1.0.0

A production-ready multitenancy solution for Go Fiber using PostgreSQL schema isolation.

✨ Features:
- Schema-based tenant isolation
- Connection pooling & caching
- Multiple resolution strategies
- Auto-migration support
- Full test coverage
- CI/CD pipeline

📦 Install: go get github.com/1Nelsonel/fiber-multitenant
📖 Docs: https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant
⭐ Star: https://github.com/1Nelsonel/fiber-multitenant

Feedback welcome!
```

## Step 8: Monitor Package Usage

### Track Metrics:

1. **GitHub Stars**: https://github.com/1Nelsonel/fiber-multitenant/stargazers
2. **Go Module Proxy Stats**: Wait ~1 week, then check download stats
3. **pkg.go.dev**: Shows imports and documentation views
4. **Issues/PRs**: Monitor community engagement

### Set Up Repository Insights:

1. Go to: https://github.com/1Nelsonel/fiber-multitenant/pulse
2. View weekly/monthly activity
3. Track contributors and traffic

## Step 9: Enable Discussions (Optional)

For community questions and feedback:

1. Go to: https://github.com/1Nelsonel/fiber-multitenant/settings
2. Scroll to "Features"
3. Check ✅ "Discussions"
4. Set up categories:
   - Q&A (for questions)
   - Ideas (for feature requests)
   - Show and tell (for user showcases)

## Step 10: Submit to Awesome Lists

Add your package to curated lists:

1. **Awesome Fiber**: https://github.com/gofiber/awesome-fiber
   - Fork the repo
   - Add your package under "Middleware"
   - Submit a PR

2. **Awesome Go**: https://github.com/avelino/awesome-go
   - Check contribution guidelines
   - Add under "Database" or "Web Frameworks"
   - Submit a PR

## Post-Publication Checklist

After publishing:

- [ ] Package visible at pkg.go.dev
- [ ] Can install with `go get github.com/1Nelsonel/fiber-multitenant`
- [ ] Badges working on README
- [ ] GitHub topics added
- [ ] License visible on GitHub
- [ ] Release v1.0.0 exists
- [ ] CI/CD pipeline passing
- [ ] Documentation complete

## Testing Installation

Ask a friend or use a different machine to test:

```bash
# Create a test project
mkdir test-fiber-multitenant
cd test-fiber-multitenant
go mod init test

# Install your package
go get github.com/1Nelsonel/fiber-multitenant

# Create a simple test
cat > main.go << 'EOF'
package main

import (
    "github.com/1Nelsonel/fiber-multitenant/middleware"
    "github.com/gofiber/fiber/v2"
)

func main() {
    app := fiber.New()

    // Test that package imports work
    _ = middleware.SubdomainResolver

    println("Package imports successfully!")
}
EOF

go run main.go
```

## Versioning Guidelines

Follow Semantic Versioning (semver):

- **v1.0.x** - Bug fixes (patch)
- **v1.x.0** - New features, backward compatible (minor)
- **v2.0.0** - Breaking changes (major)

### Creating Future Releases:

```bash
# Bug fix release
git tag -a v1.0.1 -m "Fix: [description]"
git push origin v1.0.1

# Feature release
git tag -a v1.1.0 -m "Add: [description]"
git push origin v1.1.0

# Breaking change
git tag -a v2.0.0 -m "BREAKING: [description]"
git push origin v2.0.0
```

## Support and Maintenance

### Responding to Issues:

1. **Acknowledge quickly** (within 24-48 hours)
2. **Label appropriately** (bug, enhancement, question)
3. **Ask for details** if needed
4. **Close stale issues** after 30 days of inactivity

### Handling PRs:

1. **Thank contributors** first
2. **Check CI passes**
3. **Review code carefully**
4. **Request changes** if needed
5. **Merge and release** new version

## Success Metrics

Track these over time:
- GitHub stars
- go get downloads (via go module proxy)
- Issues/PRs from community
- pkg.go.dev views
- Mentions on social media/blogs

## Congratulations! 🎉

Your package is now part of the Go ecosystem!

**Package URL**: https://pkg.go.dev/github.com/1Nelsonel/fiber-multitenant

**Installation**:
```bash
go get github.com/1Nelsonel/fiber-multitenant
```

---

For questions about publishing, refer to:
- Go Modules: https://go.dev/doc/modules/publishing
- pkg.go.dev: https://pkg.go.dev/about
