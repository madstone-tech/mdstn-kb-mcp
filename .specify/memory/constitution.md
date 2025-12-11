<!--
CONSTITUTION SYNC REPORT (v1.0.0 → v1.1.0)
============================================
Version Bump: MINOR (new principles added)
Modified Principles:
  - Added Principle 5 (Observability)
  - Added Principle 6 (Storage Abstraction)
  - Added Principle 7 (Backward Compatibility)
Added Sections:
  - Technology Stack
  - Code Review & Quality Gates
Removed Sections: None
Templates Updated:
  ✅ plan-template.md (Constitution Check section added)
  ✅ spec-template.md (must reference core principles)
  ✅ tasks-template.md (testing & architecture tasks aligned)
Follow-up: None
============================================
-->

# kbVault Constitution

## Core Principles

### I. Test-First Development (NON-NEGOTIABLE)

Tests are written and approved by the user BEFORE implementation begins. Tests must fail initially (Red-Green-Refactor cycle). All new functions require unit tests; integration tests required for storage backends, API changes, and cross-service communication. Minimum 50% test coverage enforced by CI.

**Rationale**: This ensures correctness, reduces regression, and makes code maintainable across Go versions and refactors.

### II. Storage Backend Abstraction

All persistent data access goes through the `storage.Backend` interface. Current implementations: local filesystem and S3. New backends must implement the full interface without modifying existing backends. Database or new storage types must be introduced as new implementations, not by changing the abstraction.

**Rationale**: Enables multi-backend support and prevents vendor lock-in. Users can switch backends without application changes.

### III. CLI-First Interface

Every core feature is exposed via the CLI (`cmd/kbvault/`). The CLI uses Cobra for command structure and supports both human-readable and JSON output. Stdin/stdout/stderr protocols enabled where appropriate.

**Rationale**: CLI ensures features are composable, scriptable, and usable by agents and humans alike. Drives clear, testable APIs.

### IV. Configuration via TOML & Profiles

Configuration is TOML-based (`pkg/config/`). Supports multiple named profiles for different environments (work, personal, research, etc.). Viper handles config loading and override hierarchy.

**Rationale**: TOML is human-readable; profiles enable user flexibility without code changes.

### V. Observability & Structured Logging

All external operations (file I/O, network, errors) must be logged. Use structured logging (context in error messages, no printf-style logs). Text-based I/O ensures debuggability.

**Rationale**: Debugging agents and users requires clear visibility into system behavior.

### VI. Backward Compatibility & Versioning

Follow semantic versioning (MAJOR.MINOR.PATCH). Breaking changes to public APIs (exported types, interfaces, CLI commands, storage format) require MAJOR version bump and migration guide. Non-breaking additions use MINOR, patches are bugfixes only.

**Rationale**: Users and agents depend on stable interfaces. Breaking changes must be deliberate and documented.

### VII. Simplicity & YAGNI

Prefer simple, focused functions and types. No organizational-only abstractions. Avoid premature optimization; profile and measure before adding caches, goroutines, or complex data structures. Keep functions <30 lines where practical.

**Rationale**: Simpler code is more maintainable, testable, and auditable by agents.

## Technology Stack

**Language**: Go 1.24+ | **Primary Framework**: Cobra (CLI), Viper (config)  
**Storage**: Local filesystem, AWS S3 (via AWS SDK v2) | **Testing**: testify, go test -race  
**Key Libraries**: oklog/ulid (ID generation), BurntSushi/toml (parsing)  
**Code Style**: gofmt (enforced), golangci-lint (linting), go vet (validation)  
**CI/CD**: GitHub Actions (lint, test, build, security scan, Docker build)  
**Deployment**: Docker multi-architecture builds (amd64, arm64)

## Code Review & Quality Gates

All PRs must pass:

1. **Lint**: `golangci-lint run --timeout=5m` (no deviations without justification)
2. **Test**: `go test -v -race ./...` (must pass and maintain ≥50% coverage)
3. **Build**: `make build` (binary must be executable)
4. **Security**: `govulncheck ./...` (no unpatched vulnerabilities)
5. **Docker**: Multi-architecture build must succeed

Manual code review required; reviewers verify principle alignment (especially test-first, storage abstraction, backward compatibility). Complexity violations must be documented in PR with migration plan.

## Development Workflow

1. **Create feature branch**: `git checkout -b feature/name` or `git checkout -b fix/name`
2. **Write tests first**: Design contracts and test cases before code
3. **Implement**: Ensure tests fail initially, then implement to green
4. **Check locally**: `task check` (fmt, vet, lint, test)
5. **Commit**: Follow Conventional Commits (feat:, fix:, docs:, test:, etc.)
6. **Submit PR**: Link issues, provide test coverage info, await CI + review
7. **Merge**: Squash merge to main; delete feature branch

## Governance

**Constitution Authority**: This constitution supersedes all other guidance documents (CONTRIBUTING.md, AGENTS.md). When conflicts arise, principles in this document take precedence.

**Amendment Process**: Proposed amendments must be documented with rationale and approval plan. Changes to Core Principles require explicit approval and migration timeline for affected code. Version bump rules (MAJOR/MINOR/PATCH) apply.

**Compliance Review**: Before each release, verify all code meets current principles. Flag exceptions in release notes.

**Runtime Guidance**: AGENTS.md provides implementation specifics for agents; CONTRIBUTING.md covers human contributor workflows. Both must remain synchronized with constitution.

**Version**: 1.1.0 | **Ratified**: 2025-12-11 | **Last Amended**: 2025-12-11
