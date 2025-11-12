# Maintainer Guide

Quick reference for maintaining the fiber-multitenant package.

## 🚀 Quick Links

- **Repository**: https://github.com/1Nelsonel/fiber-multitenant
- **Issues**: https://github.com/1Nelsonel/fiber-multitenant/issues
- **Pull Requests**: https://github.com/1Nelsonel/fiber-multitenant/pulls
- **Actions**: https://github.com/1Nelsonel/fiber-multitenant/actions
- **Releases**: https://github.com/1Nelsonel/fiber-multitenant/releases

## 📋 Next Steps After Pushing

### 1. Set Up Branch Protection (REQUIRED)

Follow the guide: [.github/BRANCH_PROTECTION_SETUP.md](.github/BRANCH_PROTECTION_SETUP.md)

**Critical settings:**
- ✅ Require PR reviews (1 approval)
- ✅ Require status checks to pass
- ✅ Restrict who can push to main (only you)
- ✅ Protect tags matching `v*`

**Quick setup:**
1. Go to: https://github.com/1Nelsonel/fiber-multitenant/settings/branches
2. Click "Add rule"
3. Branch name pattern: `main`
4. Enable all protections as per BRANCH_PROTECTION_SETUP.md

### 2. Wait for First CI Run

After pushing, GitHub Actions will automatically run. Check:
- https://github.com/1Nelsonel/fiber-multitenant/actions

All 4 jobs should pass:
- ✅ Lint
- ✅ Test
- ✅ Build
- ✅ Security Scan

### 3. Set Up Codecov (Optional)

For code coverage reports:
1. Go to https://codecov.io
2. Sign in with GitHub
3. Add your repository
4. Copy the token
5. Add to GitHub: Settings > Secrets and variables > Actions > New repository secret
   - Name: `CODECOV_TOKEN`
   - Value: [your token]

## 🔄 Reviewing Pull Requests

### When a PR is Submitted:

1. **Automatic Checks:**
   - CI/CD runs automatically
   - All tests must pass
   - Linting must pass
   - Security scan must pass
   - You're auto-requested as reviewer (via CODEOWNERS)

2. **Review Process:**
   ```bash
   # Check out the PR locally
   gh pr checkout <PR_NUMBER>

   # Or manually
   git fetch origin pull/<PR_NUMBER>/head:pr-<PR_NUMBER>
   git checkout pr-<PR_NUMBER>

   # Run tests locally
   go test ./...

   # Check the changes
   git diff main
   ```

3. **Provide Feedback:**
   - Comment on specific lines in GitHub
   - Request changes if needed
   - Approve when ready

4. **Merge the PR:**
   - Option 1: **Squash and merge** (recommended - clean history)
   - Option 2: **Merge commit** (preserves all commits)
   - Option 3: **Rebase and merge** (linear history)

   ```bash
   # Or via CLI
   gh pr merge <PR_NUMBER> --squash
   ```

## 🏷️ Creating Releases

### Semantic Versioning

Follow [Semantic Versioning](https://semver.org/):
- `v1.0.0` - Major version (breaking changes)
- `v1.1.0` - Minor version (new features, backward compatible)
- `v1.1.1` - Patch version (bug fixes)

### Release Process

#### Method 1: Via Git (Recommended)

```bash
# Make sure you're on main and up to date
git checkout main
git pull origin main

# Create an annotated tag
git tag -a v1.0.0 -m "Release v1.0.0: Initial stable release"

# Push the tag
git push origin v1.0.0
```

The release workflow will automatically:
- Run all tests
- Generate changelog from commits
- Create a GitHub release
- Publish release notes

#### Method 2: Via GitHub UI

1. Go to: https://github.com/1Nelsonel/fiber-multitenant/releases
2. Click "Create a new release"
3. Click "Choose a tag" and type: `v1.0.0` (create new)
4. Set release title: `Release v1.0.0`
5. Add release notes (or auto-generate)
6. Click "Publish release"

### Release Checklist

Before creating a release:

- [ ] All tests passing on main
- [ ] CHANGELOG or commit messages are clear
- [ ] Documentation is up to date
- [ ] Version number follows semver
- [ ] No open critical bugs

## 📝 Common Maintenance Tasks

### Update Dependencies

```bash
# Update all dependencies
go get -u ./...

# Tidy up
go mod tidy

# Run tests
go test ./...

# Commit if all good
git add go.mod go.sum
git commit -m "Update dependencies"
git push origin main
```

### Run Linter Locally

```bash
# Install golangci-lint if not already
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Run linter
golangci-lint run

# Auto-fix issues
golangci-lint run --fix
```

### Run Security Scan Locally

```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run scan
gosec ./...
```

### Check Test Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage
go tool cover -html=coverage.out
```

## 🐛 Handling Issues

### When Someone Reports a Bug:

1. **Acknowledge:** Comment thanking them for the report
2. **Reproduce:** Try to reproduce locally
3. **Label:** Add labels (bug, enhancement, question, etc.)
4. **Prioritize:** Set milestone if critical
5. **Fix or Guide:** Either fix yourself or guide contributor

### Issue Labels to Use:

- `bug` - Something isn't working
- `enhancement` - New feature request
- `documentation` - Documentation improvements
- `good first issue` - Good for newcomers
- `help wanted` - Extra attention needed
- `question` - Further information requested
- `wontfix` - This will not be worked on
- `duplicate` - This issue already exists

## 📊 Monitoring

### Check CI/CD Status

Visit: https://github.com/1Nelsonel/fiber-multitenant/actions

### Check Package Usage

```bash
# See who's importing your package
# (requires GitHub API token)
gh api repos/1Nelsonel/fiber-multitenant/dependents
```

### Monitor Issues and PRs

```bash
# List open issues
gh issue list

# List open PRs
gh pr list

# View PR status
gh pr status
```

## 🔒 Security

### Dependabot Alerts

GitHub will automatically:
- Scan for vulnerable dependencies
- Create PRs to update them
- Alert you of security issues

Enable at: https://github.com/1Nelsonel/fiber-multitenant/settings/security_analysis

### Security Policy

Create `.github/SECURITY.md`:
```markdown
# Security Policy

## Reporting a Vulnerability

Please report security vulnerabilities to: your-email@example.com

Do NOT open public issues for security vulnerabilities.
```

## 📦 Publishing Best Practices

### Commit Message Format

```
<type>: <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Example:
```
feat: add custom resolver support

Added CustomResolver function to allow users to implement
their own tenant resolution logic.

Closes #42
```

### Documentation Updates

Always update when:
- Adding new features → Update README.md
- Changing APIs → Update PACKAGE_SUMMARY.md
- Bug fixes → Update CHANGELOG (if you create one)

## 🎯 Goals

### Short-term (v1.0.0)
- [ ] Stable API
- [ ] Comprehensive tests
- [ ] Complete documentation
- [ ] CI/CD pipeline working

### Mid-term (v1.x.x)
- [ ] Performance benchmarks
- [ ] Additional tenant resolvers
- [ ] Database migration helpers
- [ ] Example applications

### Long-term (v2.x.x)
- [ ] Support for other databases
- [ ] Advanced caching strategies
- [ ] Metrics and observability
- [ ] Multi-database support

## 🆘 Getting Help

If you need help:
- Review GitHub Actions logs
- Check Go package documentation: `go doc`
- Test locally before pushing
- Ask in GitHub Discussions (if enabled)

## 📚 Resources

- [GitHub Actions Docs](https://docs.github.com/en/actions)
- [Go Modules](https://go.dev/blog/using-go-modules)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)

---

**Remember:** You're the sole maintainer, so:
- ✅ Take your time reviewing PRs
- ✅ Ask for changes if needed
- ✅ Test thoroughly before merging
- ✅ Keep the main branch stable
- ✅ Document decisions in issues/PRs
