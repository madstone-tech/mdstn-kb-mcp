# AGENTS.md - Guidelines for Agentic Code Development

## Quick Commands

**Build:** `make build` | **Test:** `go test -v ./...` | **Lint:** `golangci-lint run` | **Single test:** `go test -v ./pkg/types -run TestNote`

**Run single test file:** `go test -v ./cmd/kbvault -run TestConfig` | **Coverage:** `make test-coverage` | **Full check:** `make check`

## Build, Lint, Test

- **Full test suite:** `go test -v -race ./...`
- **Single package tests:** `go test -v ./pkg/types` (or cmd, internal)
- **Run single test:** `go test -v ./pkg/types -run TestNote`
- **Test with race detector:** `go test -race ./...` (catches concurrency bugs)
- **Rebuild binary:** `make build` (outputs to `bin/kbvault`)
- **Linting:** `golangci-lint run --timeout=5m` (enforces gofmt, go vet, style)
- **Coverage minimum:** 50% required by CI (`go tool cover -func=coverage.out`)

## Code Style & Imports

**Go version:** 1.24 | **Format:** `gofmt` (auto-enforced) | **Formatting:** `make fmt` before commits

**Imports:** Organize in 3 groups (stdlib, third-party, local), use `go fmt` to auto-sort:
```go
import (
	"fmt"
	"os"
	
	"github.com/spf13/cobra"
	"github.com/BurntSushi/toml"
	
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/madstone-tech/mdstn-kb-mcp/cmd/kbvault"
)
```

## Types, Naming, Error Handling

**Public types:** Exported (CamelCase): `type Note struct`, `func (n *Note) Save() error`

**Private types:** Unexported (camelCase): `type noteCache struct`, `func newNoteCache() *noteCache`

**Errors:** Use custom error types in `pkg/types/errors.go`, wrap with context:
```go
if err != nil {
	return fmt.Errorf("failed to save note: %w", err)
}
```

**Error handling:** Check all error returns; use `require.NoError(t, err)` in tests.

**Interfaces:** Small, focused (reader interface = 1-2 methods); use dependency injection.

## Testing Standards

**Test format:** `TestFunctionName_Scenario` using Arrange-Act-Assert, table-driven where appropriate:
```go
func TestSaveNote_Success(t *testing.T) {
	note := &types.Note{Title: "test"}
	err := note.Save()
	require.NoError(t, err)
	assert.Equal(t, "test", note.Title)
}
```

**Coverage requirement:** 50% minimum (check with `go tool cover -func=coverage.out`)

**Test dependencies:** Use `github.com/stretchr/testify` (require, assert)

## Architecture Notes

**Package structure:** `pkg/` (config, storage, types, ulid, vector, retry) | `cmd/kbvault` (CLI) | `internal/` (search, links, templates)

**Storage abstraction:** All backends implement `storage.Backend` interface (local, S3 with factory pattern)

**Config system:** TOML-based (`pkg/config/`), supports profiles via Viper

**Key dependencies:** Cobra (CLI), Viper (config), oklog/ulid (IDs), AWS SDK (S3), BurntSushi/toml (parsing)

## Conventions to Follow

- Document all public functions/types with comments
- Use meaningful names (avoid `x`, `temp`, `data`)
- Keep functions focused (<30 lines ideal)
- Use errors as values, not exceptions
- Prefer explicit over implicit
- Maintain <80 char line length where practical

## Additional Resources

See CONTRIBUTING.md for detailed contribution guidelines, CLAUDE.md for project architecture overview, and .github/workflows/ci-unified.yml for full CI pipeline.
