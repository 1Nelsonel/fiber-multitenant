# Branch Protection Setup Guide

This guide will help you set up branch protection rules on GitHub to ensure only you can merge PRs and create releases.

## Step 1: Enable Branch Protection for `main` Branch

1. Go to your repository on GitHub
2. Click on **Settings** tab
3. Click on **Branches** in the left sidebar
4. Under "Branch protection rules", click **Add rule**

## Step 2: Configure Protection Rules

### Branch name pattern
```
main
```

### Protect matching branches - Enable these settings:

#### Required Status Checks
- ✅ **Require status checks to pass before merging**
  - ✅ **Require branches to be up to date before merging**
  - Select these status checks (they will appear after your first CI run):
    - `Lint`
    - `Test`
    - `Build`
    - `Security Scan`

#### Pull Request Reviews
- ✅ **Require a pull request before merging**
  - ✅ **Require approvals**: Set to `1`
  - ✅ **Dismiss stale pull request approvals when new commits are pushed**
  - ✅ **Require review from Code Owners** (optional)

#### Conversation Resolution
- ✅ **Require conversation resolution before merging**

#### Commit Restrictions
- ✅ **Do not allow bypassing the above settings**
- ⚠️ **Allow force pushes**: Leave UNCHECKED
- ⚠️ **Allow deletions**: Leave UNCHECKED

#### Rules Applied to Administrators
- ✅ **Include administrators** (This ensures even you follow the PR process)
  - OR leave unchecked if you want to push directly to main (not recommended)

#### Restrict Who Can Push
- ✅ **Restrict who can push to matching branches**
  - Add yourself as the only user who can push
  - This prevents others from pushing directly to main

## Step 3: Set Up CODEOWNERS (Optional but Recommended)

Create a file `.github/CODEOWNERS` to automatically request your review on all PRs:

```
# Default owner for everything in the repo
* @1Nelsonel

# Specific paths (optional)
/middleware/ @1Nelsonel
/tenantstore/ @1Nelsonel
```

## Step 4: Repository Settings for Merging

1. Go to **Settings** > **General**
2. Scroll to **Pull Requests** section
3. Configure merge options:
   - ✅ **Allow merge commits** (recommended)
   - ✅ **Allow squash merging** (recommended for clean history)
   - ❌ **Allow rebase merging** (optional)
   - ✅ **Automatically delete head branches** (cleanup after merge)

## Step 5: Set Up Repository Permissions

1. Go to **Settings** > **Collaborators and teams**
2. Set base permissions to:
   - **Read** for public repositories
   - This ensures contributors can fork and submit PRs but cannot push directly

## Step 6: Protected Tags (for Releases)

1. Go to **Settings** > **Tags**
2. Click **Add rule**
3. Tag name pattern: `v*` (matches v1.0.0, v2.1.3, etc.)
4. Enable:
   - ✅ **Require signed commits**
   - ✅ **Restrict who can push to matching tags**
     - Add yourself as the only user

## Workflow After Setup

### For Contributors:
1. Fork the repository
2. Create a feature branch
3. Submit a Pull Request
4. Wait for CI checks to pass
5. Wait for your review and approval
6. You merge the PR (they cannot)

### For You (Maintainer):
1. Review the Pull Request
2. Check that all CI/CD checks pass
3. Review the code changes
4. Approve the PR
5. Merge the PR (squash, merge commit, or rebase)
6. Create release tags when ready

## Creating Release Tags

After merging PRs, you can create version tags:

```bash
# Create an annotated tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push the tag
git push origin v1.0.0
```

Or use the GitHub Releases UI:
1. Go to **Releases** > **Create a new release**
2. Choose a tag (e.g., `v1.0.0`)
3. Add release notes
4. Publish release

## Testing the Setup

1. Try to push directly to main - should be blocked
2. Create a test PR - should require passing checks
3. Try to merge without approval - should be blocked
4. Approve and merge - should succeed

## Recommended: Enable Required Signed Commits

For extra security:
1. Go to branch protection rule for `main`
2. Enable: ✅ **Require signed commits**
3. Set up GPG signing locally:

```bash
# Configure Git to sign commits
git config --global commit.gpgsign true
git config --global user.signingkey YOUR_GPG_KEY_ID
```

## Summary Checklist

- [ ] Branch protection enabled for `main`
- [ ] Required status checks configured
- [ ] PR approval required (1 reviewer)
- [ ] Force pushes disabled
- [ ] Only you can push to main branch
- [ ] Tag protection enabled for `v*`
- [ ] CODEOWNERS file created
- [ ] Auto-delete merged branches enabled
- [ ] Tested with a sample PR

Your repository is now protected and only you can merge PRs and create release tags!
