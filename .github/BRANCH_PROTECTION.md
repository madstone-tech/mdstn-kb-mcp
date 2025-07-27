# GitHub Branch Protection Configuration

This document outlines the required GitHub repository settings and branch protection rules.

## Repository Settings

### General Settings

1. **Repository Features**
   - ✅ Issues
   - ✅ Projects  
   - ✅ Wiki (disabled)
   - ✅ Discussions
   - ✅ Sponsorships (disabled)

2. **Pull Requests**
   - ✅ Allow merge commits (disabled)
   - ✅ Allow squash merging (enabled)
   - ✅ Allow rebase merging (disabled)
   - ✅ Always suggest updating pull request branches
   - ✅ Allow auto-merge
   - ✅ Automatically delete head branches

3. **Archives**
   - ✅ Include Git LFS objects in archives

## Branch Protection Rules

### Main Branch (`main`)

Navigate to: **Settings** → **Branches** → **Add rule**

#### Branch name pattern: `main`

#### Protect matching branches:
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1**
  - ✅ Dismiss stale PR approvals when new commits are pushed
  - ✅ Require review from code owners
  - ✅ Restrict pushes that create new files
  - ✅ Require conversation resolution before merging

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Required status checks:**
    - `All PR checks passed` (from pr-checks.yml workflow)
    - `Lint` (from pr-checks.yml workflow)
    - `Test (ubuntu-latest, 1.23)` (from pr-checks.yml workflow)
    - `Test (macos-latest, 1.23)` (from pr-checks.yml workflow)
    - `Build` (from pr-checks.yml workflow)
    - `Security scan` (from pr-checks.yml workflow)
    - `Docker build test` (from pr-checks.yml workflow)

- ✅ **Require conversation resolution before merging**

- ✅ **Require signed commits** (optional but recommended)

- ✅ **Require linear history**

- ✅ **Require deployments to succeed before merging** (if using deployments)

- ✅ **Lock branch** (disabled - allow admins to push)

- ✅ **Do not allow bypassing the above settings**

- ✅ **Restrict pushes that create new files** (disabled for development)

#### Exceptions:
- **Allow force pushes:** ❌ Everyone (disabled)
- **Allow deletions:** ❌ (disabled)

### Develop Branch (`develop`) - Optional

If using a develop branch for feature integration:

#### Branch name pattern: `develop`

#### Protect matching branches:
- ✅ **Require a pull request before merging**
  - ✅ Require approvals: **1**
  - ✅ Dismiss stale PR approvals when new commits are pushed
  - ❌ Require review from code owners (more flexible for develop)

- ✅ **Require status checks to pass before merging**
  - ✅ Require branches to be up to date before merging
  - **Required status checks:** (same as main)

- ✅ **Require conversation resolution before merging**

## Repository Secrets and Variables

### Required Secrets

1. **GITHUB_TOKEN** (automatically provided)
   - Used for: GitHub Actions, releases, package publishing

2. **CODECOV_TOKEN** (if using Codecov)
   - Used for: Test coverage reporting

### Required Variables

1. **REGISTRY** = `ghcr.io`
   - Used for: Docker image registry

## Ruleset Configuration (Alternative)

For newer GitHub repositories, you can use Rulesets instead:

### Create Ruleset: "Main Branch Protection"

1. **Target branches:** `main`
2. **Rules:**
   - ✅ Restrict creations
   - ✅ Restrict updates  
   - ✅ Restrict deletions
   - ✅ Required linear history
   - ✅ Required pull request
     - Required approving review count: **1**
     - Dismiss stale reviews: **enabled**
     - Require review from code owners: **enabled**
     - Require conversation resolution: **enabled**
   - ✅ Required status checks
     - Require branches to be up to date: **enabled**
     - Required checks: (list all workflow jobs)

## Workflow Permissions

Ensure GitHub Actions have the correct permissions:

### Repository → Settings → Actions → General

1. **Actions permissions:**
   - ✅ Allow all actions and reusable workflows

2. **Workflow permissions:**
   - ✅ Read and write permissions
   - ✅ Allow GitHub Actions to create and approve pull requests

### Token Permissions in Workflows

```yaml
permissions:
  contents: read        # Read repository contents
  packages: write       # Push Docker images
  security-events: write # Upload security scan results
  pull-requests: write  # Comment on PRs
  actions: read         # Read workflow status
  checks: write         # Create check runs
```

## Auto-merge Configuration

To enable auto-merge for dependency updates:

### Repository → Settings → General → Pull Requests
- ✅ Allow auto-merge

### Dependabot Configuration
```yaml
# In .github/dependabot.yml
auto-merge:
  - match:
      update_type: "semver:patch"
      dependency_type: "production"
```

## Verification Commands

After setting up the protection rules, verify with:

```bash
# Check if branch protection is working
git push origin main  # Should fail

# Check required status checks
# Create a PR and verify all checks are required

# Verify auto-delete works
# Merge a PR and check if branch is deleted
```

## Troubleshooting

### Common Issues:

1. **Status checks not found**
   - Ensure workflows have run at least once
   - Check workflow names match exactly
   - Verify workflow triggers include pull_request

2. **Admin bypass not working**
   - Check "Do not allow bypassing" is disabled
   - Verify admin permissions

3. **Auto-merge not working**
   - Verify all required checks pass
   - Check auto-merge is enabled for the repository
   - Ensure PR author has sufficient permissions

### GitHub CLI Commands:

```bash
# View branch protection status
gh api repos/:owner/:repo/branches/main/protection

# Update branch protection (example)
gh api repos/:owner/:repo/branches/main/protection \
  --method PUT \
  --field required_status_checks='{"strict":true,"checks":[{"context":"lint"}]}'
```