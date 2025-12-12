# Implementation Plan: S3 Vector Search & Semantic Capabilities

**Branch**: `006-s3-vector-search` | **Date**: December 12, 2024 | **Spec**: `specs/006-s3-vector-search/spec.md`
**Input**: Feature specification from `/specs/006-s3-vector-search/spec.md`

## Summary

Add semantic search capabilities to kbVault by integrating Ollama for local embedding generation and S3 Vector Search for similarity queries. Users can search notes by conceptual meaning rather than exact keywords, with configurable Ollama models optimized for standard laptop hardware (not servers). Default model: nomic-embed-text (fast, free, 384 dimensions). Support hybrid search combining text and semantic results with user-configurable weighting.

## Technical Context

**Language/Version**: Go 1.24+  
**Primary Dependencies**: AWS SDK v2 (S3 Vector Search), Ollama API client (embedding generation), existing Cobra CLI, Viper config  
**Storage**: AWS S3 Vector Search backend (Session 5 dependency) + local embedding metadata  
**Testing**: testify, `go test -race` (minimum 50% coverage required)  
**Target Platform**: Linux/macOS/Windows CLI, standard laptop hardware (8GB RAM, modern CPU)  
**Project Type**: CLI single project (existing kbVault structure)  
**Performance Goals**: Semantic search results <2s (1000+ notes), embedding generation ~10-15 min for 1000 notes (nomic-embed-text), similarity search <1s  
**Constraints**: Optimized for local Ollama on standard hardware (not high-performance servers), graceful degradation if Ollama unavailable  
**Scale/Scope**: Support vaults up to 10,000 notes; larger vaults may require optimization beyond Session 6

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

✅ **I. Test-First Development**: Session 6 requires >50% test coverage. All new search functions, embedding operations, and storage backend changes must include unit tests. Integration tests required for Ollama API client and S3 Vector Search integration.

✅ **II. Storage Backend Abstraction**: Embeddings and vector metadata use existing `storage.Backend` interface (S3 Vector Search from Session 5). No modification to storage abstraction; embeddings stored as metadata alongside notes.

✅ **III. CLI-First Interface**: All semantic search features exposed via Cobra CLI (`kbvault search --semantic`, `kbvault search --similar-to`, `kbvault vectorize`). Human-readable and JSON output supported.

✅ **IV. Configuration via TOML & Profiles**: Vector search settings configured in TOML per profile (vector section): ollama_model, ollama_endpoint, text_weight, vector_weight, similarity_threshold, auto_regenerate. Viper loads config with override hierarchy.

✅ **V. Observability & Structured Logging**: All Ollama API calls, S3 operations, and embedding generation logged. Error messages include context. Batch processing shows progress updates.

✅ **VI. Backward Compatibility**: Session 6 addition is non-breaking (semantic search optional, text-only search still works). Config changes are additive only. Vector metadata stored alongside notes without altering existing note format.

✅ **VII. Simplicity & YAGNI**: Ollama client will be a simple wrapper around HTTP API (not full SDK). Embedding cache uses map-based in-memory structure initially (not complex DB). Functions kept focused (<30 lines where practical).

**GATE STATUS**: ✅ PASS - No violations. All principles alignable with current plan.

## Project Structure

### Documentation (this feature)

```text
specs/006-s3-vector-search/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output - research findings on Ollama integration
├── data-model.md        # Phase 1 output - data model, entities, storage schema
├── quickstart.md        # Phase 1 output - developer quickstart guide
├── contracts/           # Phase 1 output - API contracts
│   ├── search-api.md
│   ├── embedding-api.md
│   └── config-schema.md
└── tasks.md             # Phase 2 output (/speckit.tasks command - created separately)
```

### Source Code (existing kbVault structure)

```text
# New packages for vector search
pkg/
├── vector/              # Vector search abstractions
│   ├── embedding.go     # Embedding data structures
│   ├── search.go        # Vector search logic
│   └── ollama/          # Ollama client implementation
│       ├── client.go    # HTTP client for Ollama API
│       └── client_test.go
├── search/              # Enhanced search engine (from Session 4)
│   ├── engine.go        # Add vector search integration
│   └── engine_test.go

# Enhanced CLI
cmd/kbvault/
├── search.go            # Enhanced with --semantic, --similar-to flags
├── search_test.go
├── vectorize.go         # New command: batch embedding generation
└── vectorize_test.go

# Configuration
pkg/config/
├── config.go            # Add vector section to TOML schema
└── config_test.go

# Storage (using existing Session 5 S3 Vector Search backend)
# No new files; embeddings stored as note metadata via storage.Backend interface
```

**Structure Decision**: Single CLI project (Option 1). Session 6 extends existing kbVault structure with new `pkg/vector/` package for Ollama integration and embedding operations. Enhanced `pkg/search/` for vector search. New CLI commands (`vectorize`) and flags (`--semantic`, `--similar-to`) in `cmd/kbvault/`. Uses existing storage abstraction (Session 5 S3 Vector Search backend) for persistence.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

| Violation | Why Needed | Simpler Alternative Rejected Because |
|-----------|------------|-------------------------------------|
| [e.g., 4th project] | [current need] | [why 3 projects insufficient] |
| [e.g., Repository pattern] | [specific problem] | [why direct DB access insufficient] |
