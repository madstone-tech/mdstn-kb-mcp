# Contributing to kbVault

Thank you for your interest in contributing to kbVault! This document provides guidelines and information for contributors.

## Code of Conduct

This project adheres to a code of conduct. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

## Development Process

### Prerequisites

- Go 1.23 or later
- Git
- Make
- Docker or Podman (for container testing)
- golangci-lint (for linting)

### Setting Up Development Environment

1. **Fork and Clone**
   ```bash
   git clone https://github.com/YOUR_USERNAME/mdstn-kb-mcp.git
   cd mdstn-kb-mcp
   ```

2. **Install Dependencies**
   ```bash
   make setup
   ```

3. **Verify Setup**
   ```bash
   make quick
   ```

### Development Workflow

1. **Create a Feature Branch**
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b fix/your-bug-fix
   ```

2. **Make Changes**
   - Write code following Go best practices
   - Add tests for new functionality
   - Update documentation as needed

3. **Test Your Changes**
   ```bash
   # Run all checks
   make check
   
   # Run specific tests
   go test ./pkg/your-package -v
   
   # Test with race detection
   go test -race ./...
   
   # Check test coverage
   make test-coverage
   ```

4. **Lint Your Code**
   ```bash
   make lint
   ```

5. **Build and Test Docker Image**
   ```bash
   make docker-build
   make docker-shell
   ```

6. **Commit Your Changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

### Commit Message Guidelines

We follow [Conventional Commits](https://conventionalcommits.org/) specification:

- `feat:` New features
- `fix:` Bug fixes
- `docs:` Documentation changes
- `style:` Code style changes (formatting, etc.)
- `refactor:` Code refactoring
- `test:` Adding or updating tests
- `chore:` Maintenance tasks
- `ci:` CI/CD changes
- `perf:` Performance improvements

Examples:
```
feat: add ULID validation for note IDs
fix: resolve concurrent access issue in local storage
docs: update configuration examples
test: add integration tests for storage backends
```

### Pull Request Process

1. **Before Creating PR**
   - Ensure all tests pass: `make test`
   - Ensure linting passes: `make lint`
   - Ensure build succeeds: `make build`
   - Update documentation if needed

2. **Create Pull Request**
   - Use the provided PR template
   - Link related issues
   - Provide clear description of changes
   - Include test coverage information

3. **PR Requirements**
   - All CI checks must pass
   - Code coverage must be maintained (≥80%)
   - Code review approval required
   - Branch protection rules enforced

4. **Merge Process**
   - Squash and merge for feature branches
   - Linear history maintained on main branch
   - Delete feature branch after merge

## Code Standards

### Go Code Style

- Follow `gofmt` formatting
- Use `golangci-lint` configuration
- Write idiomatic Go code
- Include package documentation
- Use meaningful variable names

### Testing Standards

- Write unit tests for all new functions
- Maintain test coverage ≥80%
- Use table-driven tests where appropriate
- Include integration tests for major features
- Mock external dependencies

Example test structure:
```go
func TestFunction_Success(t *testing.T) {
    // Arrange
    input := "test input"
    expected := "expected output"
    
    // Act
    result, err := Function(input)
    
    // Assert
    require.NoError(t, err)
    assert.Equal(t, expected, result)
}
```

### Documentation Standards

- Document all public functions and types
- Include usage examples
- Update README.md for user-facing changes
- Write clear commit messages
- Update CHANGELOG.md for releases

## Architecture Guidelines

### Package Structure

```
pkg/
├── config/     # Configuration management
├── storage/    # Storage backends
├── retry/      # Error handling and retry logic
├── types/      # Core data types
└── ulid/       # ULID generation and validation
```

### Interface Design

- Design interfaces for testability
- Keep interfaces small and focused
- Use dependency injection
- Follow Go interface conventions

### Error Handling

- Use structured error types
- Implement retry mechanisms where appropriate
- Provide meaningful error messages
- Include context in error chains

## Security Guidelines

- Never commit secrets or credentials
- Validate all inputs
- Use secure defaults
- Follow security best practices
- Report security issues privately

## Performance Guidelines

- Profile performance-critical code
- Use benchmarks for optimization
- Consider memory allocations
- Implement efficient algorithms
- Test with realistic data sizes

## Release Process

1. **Version Numbering**
   - Follow Semantic Versioning (SemVer)
   - Format: `vMAJOR.MINOR.PATCH`

2. **Release Checklist**
   - Update CHANGELOG.md
   - Update version in documentation
   - Create release notes
   - Tag release with proper version

3. **Automated Release**
   - CI/CD handles binary builds
   - Docker images published automatically
   - GitHub releases created automatically

## Getting Help

- **Questions**: Open a [Discussion](https://github.com/madstone-tech/mdstn-kb-mcp/discussions)
- **Bugs**: Create a [Bug Report](https://github.com/madstone-tech/mdstn-kb-mcp/issues/new?template=bug_report.yml)
- **Features**: Create a [Feature Request](https://github.com/madstone-tech/mdstn-kb-mcp/issues/new?template=feature_request.yml)
- **Security**: Use [Security Advisories](https://github.com/madstone-tech/mdstn-kb-mcp/security/advisories/new)

## Useful Commands

```bash
# Development
make dev              # Run in development mode
make quick            # Quick checks (format, vet, test)
make check            # Full checks (format, vet, lint, test)

# Testing
make test             # Run all tests
make test-coverage    # Generate coverage report
make benchmark        # Run benchmarks

# Building
make build            # Build binary
make build-all        # Build for all platforms
make docker-build     # Build Docker image

# Quality
make lint             # Run linter
make fmt              # Format code
make vet              # Run go vet

# Cleanup
make clean            # Clean build artifacts
```

## License

By contributing to kbVault, you agree that your contributions will be licensed under the MIT License.