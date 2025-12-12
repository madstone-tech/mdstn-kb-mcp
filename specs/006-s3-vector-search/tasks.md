# Tasks: S3 Vector Search & Semantic Capabilities (Session 6)

**Status**: Phase 2 Task Breakdown  
**Date**: December 12, 2024  
**Feature Branch**: `006-s3-vector-search`  
**Total Tasks**: 58 (organized across 8 phases)  
**Test Coverage**: >60% required (all unit tests included)

---

## Table of Contents

1. [Overview & Strategy](#overview--strategy)
2. [Phase 1: Setup & Infrastructure](#phase-1-setup--infrastructure)
3. [Phase 2: Foundational Components](#phase-2-foundational-components)
4. [Phase 3: User Story 1 - Semantic Search](#phase-3-user-story-1---semantic-search)
5. [Phase 4: User Story 2 - Similar Notes](#phase-4-user-story-2---similar-notes)
6. [Phase 5: User Story 3 - Hybrid Search](#phase-5-user-story-3---hybrid-search)
7. [Phase 6: User Story 4 - Model Configuration](#phase-6-user-story-4---model-configuration)
8. [Phase 7: User Story 5 - Search Explanation](#phase-7-user-story-5---search-explanation)
9. [Phase 8: Polish & Integration](#phase-8-polish--integration)

---

## Overview & Strategy

### Implementation Approach

**MVP Scope (User Story 1 Only)**:
- Semantic search with natural language queries
- Ollama embedding generation (nomic-embed-text default)
- S3 Vectors for similarity search
- Basic result ranking by semantic similarity
- Graceful fallback if Ollama unavailable

**Full Release (All User Stories)**:
- Complete with hybrid search, model configuration, explanations
- Batch vectorization for large vaults
- Configuration management per profile
- Cache management and optimization

### Parallel Execution Opportunities

**Phase 3 (Semantic Search)**:
- [ ] T010: Ollama client can be developed in parallel with T011: Vector embedding models
- [ ] T013: Search result types and T014: Cache implementation are independent

**Phase 4 (Similar Notes)**:
- [ ] T025: Similar notes search and T026: Top-K selection use same utilities (serial dependency)

**Phase 5 (Hybrid Search)**:
- [ ] T035: Text weighting and T036: Vector weighting can be developed in parallel
- [ ] T037: Hybrid ranking depends on both (T035 + T036)

### User Story Dependencies

```
US1 (Semantic Search)
  ├─ Ollama Client (T010)
  ├─ Vector Models (T011)
  ├─ Cache (T014)
  └─ Search Result Types (T013)
       ↓
     ALL OTHERS depend on US1 completion
       ↓
US2 (Similar Notes) ← depends on US1
US3 (Hybrid Search) ← depends on US1
US4 (Model Config) ← depends on US1
US5 (Search Explain) ← depends on US1
```

### Independent Test Criteria

Each user story has independent test scenarios (no cross-story dependencies):

1. **US1**: Semantic search works with sample notes (self-contained test data)
2. **US2**: Similar notes command works on any note (independent of search type)
3. **US3**: Hybrid search combines scores correctly (uses mock scores if needed)
4. **US4**: Model configuration applies to all search types (config-only test)
5. **US5**: Explanations display correctly (uses mocked search results)

---

## Phase 1: Setup & Infrastructure

**Goal**: Initialize project structure, add dependencies, set up test fixtures

**Duration**: 1 day  
**Blockers**: None  
**Parallel Tasks**: T001-T005 can run in parallel (no dependencies)

### Tasks

- [x] T001 Create vector package structure: `pkg/vector/`, `pkg/vector/ollama/`, `pkg/vector/cache/` directories and `_test.go` files
- [x] T002 Add AWS S3 Vectors SDK dependency: Update `go.mod` to include `github.com/aws/aws-sdk-go-v2/service/s3vectors@latest`
- [x] T003 Create test fixtures directory: `specs/006-s3-vector-search/test-data/` with sample notes JSON and expected embeddings
- [x] T004 Extend config schema: Add `[vector]` section to `pkg/config/config.go` with TOML struct tags for 11 vector fields
- [x] T005 Create test utilities: `pkg/vector/testutil/fixtures.go` with mock Ollama responses and test vectors (384 dims)

---

## Phase 2: Foundational Components

**Goal**: Implement core vector search infrastructure (shared by all user stories)

**Duration**: 2 days  
**Blockers**: T001-T005  
**Parallel Opportunities**: T006-T009 can run in parallel (independent modules)

### Tasks

- [x] T006 Implement Ollama HTTP client interface: `pkg/vector/ollama/client.go` with `Embed(ctx, texts) ([][]float32, error)` method and connection pooling
- [x] T007 Implement vector types and validation: `pkg/types/vector.go` with `VectorEmbedding`, `SearchResult`, `SearchQuery` structs and validation methods
- [x] T008 Implement cosine similarity function: `pkg/vector/similarity.go` with `CosineSimilarity(a, b []float32) float32` and performance optimizations
- [x] T009 Implement in-memory LRU cache: `pkg/vector/cache/cache.go` with thread-safe Get/Set, eviction, TTL support (<50 LOC)

### Unit Tests (Required)

- [x] T009a Unit tests for cache: `pkg/vector/cache/cache_test.go` - hit/miss/eviction scenarios, >80% coverage
- [x] T008a Unit tests for similarity: `pkg/vector/similarity_test.go` - various vector pairs, edge cases (zeros, negatives, norms), >80% coverage
- [x] T006a Integration tests for Ollama client: `pkg/vector/ollama/client_test.go` - mock Ollama server, happy path + error handling, >70% coverage

---

## Phase 3: User Story 1 - Semantic Search with Natural Language

**Goal**: Enable natural language semantic queries with Ollama embeddings

**User Story**: A researcher wants to find notes using natural language queries like "distributed systems concepts" and get semantically related results

**Independent Test**: `kbvault search --semantic "distributed systems" --profile test-vault` returns semantically similar notes (test data with "microservices", "consensus", "fault-tolerance")

**Acceptance Criteria**:
- ✅ Semantic search returns 3+ related notes (different terminology)
- ✅ Each result has relevance score (0-1)
- ✅ Results ordered by score descending
- ✅ Ollama unavailable falls back to text search with message

**Duration**: 2 days  
**Blockers**: Phase 2 (T006-T009)

### Tasks

- [ ] T010 [US1] Implement VectorSearchBackend interface: `pkg/vector/backend.go` with `IndexVector()`, `SearchVectors()`, `GetVector()` methods (storage abstraction)
- [ ] T011 [US1] Create S3 Vectors backend implementation: `pkg/vector/s3/backend.go` (implement VectorSearchBackend for AWS S3 Vectors)
- [ ] T012 [US1] Create vector factory: `pkg/vector/factory.go` with `NewBackend(config)` to instantiate S3 Vectors backend
- [ ] T013 [US1] Extend search result types: `pkg/types/search.go` with `SemanticSearchRequest` and `HybridSearchResult` structs
- [ ] T014 [US1] Implement semantic search engine: `internal/search/semantic.go` with query embedding → similarity search → result ranking
- [ ] T015 [US1] Create/enhance search command: `cmd/kbvault/search.go` with `--semantic` flag, calls semantic search engine, returns results
- [ ] T016 [US1] Implement Ollama availability check: `pkg/vector/ollama/availability.go` with graceful degradation (fallback to text-only if Ollama unavailable)
- [ ] T017 [US1] Add semantic search to config loading: `pkg/config/config.go` loads `vector.*` settings from TOML, validates against schema

### Unit Tests (Required for US1)

- [ ] T010a Unit tests for VectorSearchBackend interface: `pkg/vector/backend_test.go` - mock backend, contract tests, >70% coverage
- [ ] T011a Integration tests for S3 Vectors backend: `pkg/vector/s3/backend_test.go` - mocked S3, vector indexing/search, >60% coverage
- [ ] T014a Unit tests for semantic search: `internal/search/semantic_test.go` - query embedding, ranking, threshold filtering, >75% coverage
- [ ] T015a Integration tests for search command: `cmd/kbvault/search_test.go` - CLI parsing, calls semantic engine, JSON output, >70% coverage

### Integration Test (End-to-End for US1)

- [ ] T018 [US1] E2E test: Semantic search flow: `specs/006-s3-vector-search/test-e2e/semantic_test.go`
  - Setup: Create vault with test notes (different terminology)
  - Action: Run `kbvault search --semantic "distributed systems"`
  - Assert: Returns 3+ related notes, scores in 0-1 range, Ollama unavailable falls back gracefully

---

## Phase 4: User Story 2 - Find Similar Notes

**Goal**: Enable "similar to this note" discovery

**User Story**: User editing a note wants to find other notes discussing similar concepts via `--similar-to <note-id>`

**Independent Test**: `kbvault search --similar-to "01HQXYZ123" --limit 5` returns 5 most similar notes

**Acceptance Criteria**:
- ✅ Similar notes command works (uses same embedding/similarity as US1)
- ✅ Returns top K results (--limit flag)
- ✅ Excludes target note by default
- ✅ Results ranked by similarity (highest first)

**Duration**: 1 day  
**Blockers**: Phase 3 (US1 must complete first)

### Tasks

- [ ] T019 [US2] Implement similar notes search: `internal/search/similar.go` with query note embedding → find all similar → top-K selection
- [ ] T020 [US2] Create top-K selection utility: `pkg/vector/topk.go` with efficient min-heap based selection (O(n log k))
- [ ] T021 [US2] Add --similar-to flag to search command: `cmd/kbvault/search.go` extended with similar notes handling
- [ ] T022 [US2] Implement note exclusion logic: Similar search excludes target note (unless `--exclude-self false`)
- [ ] T023 [US2] Add limit parameter handling: Default 10, user-configurable via `--limit` flag

### Unit Tests (Required for US2)

- [ ] T020a Unit tests for top-K selection: `pkg/vector/topk_test.go` - various K values, 1K-10K results, performance >80% coverage
- [ ] T019a Unit tests for similar notes: `internal/search/similar_test.go` - ranking, exclusion, limits, >75% coverage
- [ ] T021a CLI tests for --similar-to: `cmd/kbvault/search_test.go` extended - flag parsing, error handling, >70% coverage

### Integration Test (End-to-End for US2)

- [ ] T024 [US2] E2E test: Similar notes flow: `specs/006-s3-vector-search/test-e2e/similar_test.go`
  - Setup: Create vault with related notes (REST API, HTTP, web services)
  - Action: Run `kbvault search --similar-to <rest-api-note-id> --limit 5`
  - Assert: Returns 5 notes, ordered by similarity, excludes target note

---

## Phase 5: User Story 3 - Hybrid Search with Text and Semantic Results

**Goal**: Combine text and semantic search with configurable weighting

**User Story**: User wants balanced search combining exact keywords (text) with conceptual meaning (semantic)

**Independent Test**: `kbvault search "authentication" --text-weight 0.7 --vector-weight 0.3` returns results ranked by both text and semantic relevance

**Acceptance Criteria**:
- ✅ Hybrid search combines text + semantic scores
- ✅ Configurable weights (default 70% text, 30% vector)
- ✅ Results ordered by combined score
- ✅ Text-only and vector-only modes work (weights: 1.0/0.0 and 0.0/1.0)

**Duration**: 1.5 days  
**Blockers**: Phase 3 (US1) - reuses semantic search infrastructure

### Tasks

- [ ] T025 [US3] Implement score weighting formula: `pkg/vector/scoring.go` with `HybridScore(textScore, vectorScore, textWeight) float32`
- [ ] T026 [US3] Extend search engine for hybrid mode: `internal/search/engine.go` enhanced to call both text and semantic search, combine results
- [ ] T027 [US3] Add weight flags to search command: `cmd/kbvault/search.go` with `--text-weight` and `--vector-weight` flags (default 0.7/0.3)
- [ ] T028 [US3] Implement mode detection: Auto-detect search mode (text-only if vector_weight=0, vector-only if text_weight=0, hybrid otherwise)
- [ ] T029 [US3] Add mode fallback: If semantic search unavailable, automatically use text-only with user message
- [ ] T030 [US3] Validate weight configuration: Ensure text_weight + vector_weight ≤ 1.0, normalize if needed

### Unit Tests (Required for US3)

- [ ] T025a Unit tests for scoring: `pkg/vector/scoring_test.go` - weight combinations (0/100, 50/50, 100/0), >80% coverage
- [ ] T026a Unit tests for hybrid engine: `internal/search/engine_test.go` extended - text+vector combination, ranking, >75% coverage
- [ ] T027a CLI tests for weight flags: `cmd/kbvault/search_test.go` extended - flag parsing, normalization, >70% coverage

### Integration Test (End-to-End for US3)

- [ ] T031 [US3] E2E test: Hybrid search flow: `specs/006-s3-vector-search/test-e2e/hybrid_test.go`
  - Setup: Create vault with exact matches and semantic matches
  - Action 1: Run `kbvault search "authentication" --text-weight 0.7 --vector-weight 0.3`
  - Assert 1: Returns both exact keyword matches and semantic matches, prioritizes text
  - Action 2: Run same search with `--text-weight 0.3 --vector-weight 0.7`
  - Assert 2: Same results different order, prioritizes semantic

---

## Phase 6: User Story 4 - Configurable Ollama Models

**Goal**: Allow users to switch embedding models and configure local/remote Ollama

**User Story**: User optimizes for speed/quality trade-off by choosing different Ollama models (nomic vs mxbai-embed-large)

**Independent Test**: `kbvault config set vector.ollama_model "mxbai-embed-large" && kbvault vectorize` regenerates all embeddings with new model

**Acceptance Criteria**:
- ✅ User can configure ollama_model in TOML
- ✅ User can switch models and regenerate embeddings
- ✅ User can configure ollama_endpoint (local or remote)
- ✅ New model validated (must be installed in Ollama)
- ✅ Batch vectorization supports different models

**Duration**: 1.5 days  
**Blockers**: Phase 2 (foundational) + Phase 3 (US1)

### Tasks

- [ ] T032 [US4] Create vectorize command: `cmd/kbvault/vectorize.go` with `kbvault vectorize [--batch-size N] [--force] [--dry-run]`
- [ ] T033 [US4] Implement batch embedding generation: `internal/vector/batch.go` with batching loop, progress tracking, error recovery
- [ ] T034 [US4] Add model validation: Check if configured model exists in Ollama via `GET /api/tags`, provide install instructions if missing
- [ ] T035 [US4] Implement model switching: `internal/vector/model_switch.go` to delete old vectors, regenerate with new model
- [ ] T036 [US4] Add remote Ollama support: Parse `ollama_endpoint` config, support both `http://localhost:11434` and `http://remote-host:11434`
- [ ] T037 [US4] Create vectorization progress tracking: Display progress bar (X% complete), batch status, ETA
- [ ] T038 [US4] Implement resumable vectorization: Track last successful batch, resume from failure point

### Unit Tests (Required for US4)

- [ ] T033a Unit tests for batch embedding: `internal/vector/batch_test.go` - batching logic, retries, >75% coverage
- [ ] T034a Unit tests for model validation: `internal/vector/model_validation_test.go` - installed/missing models, >70% coverage
- [ ] T035a Unit tests for model switching: `internal/vector/model_switch_test.go` - cleanup, regeneration, >70% coverage

### Integration Test (End-to-End for US4)

- [ ] T039 [US4] E2E test: Model configuration flow: `specs/006-s3-vector-search/test-e2e/model_config_test.go`
  - Setup: Vault with notes, ollama_model = "nomic-embed-text"
  - Action 1: Run `kbvault vectorize` (initial indexing)
  - Assert 1: All notes vectorized with nomic model (384 dims)
  - Action 2: `kbvault config set vector.ollama_model "mxbai-embed-large" && kbvault vectorize --force`
  - Assert 2: Embeddings regenerated with new model (1024 dims), old vectors deleted

---

## Phase 7: User Story 5 - Search Explanation and Relevance Scoring

**Goal**: Show users why each result matched (text score, semantic score, combined score)

**User Story**: User sees breakdown: "Text: 85%, Semantic: 72%, Overall: 79%" for each search result

**Independent Test**: `kbvault search "authentication" --explain --json` returns results with explanation field showing score components

**Acceptance Criteria**:
- ✅ Search results include `explanation` field with score breakdown
- ✅ Shows text_match_score, semantic_match_score, relevance_score
- ✅ --explain flag shows detailed explanation (text > component analysis)
- ✅ Both JSON and human-readable output formats work
- ✅ Semantic-only and text-only searches show appropriate explanations

**Duration**: 1 day  
**Blockers**: Phase 3 (US1) + Phase 5 (US3)

### Tasks

- [ ] T040 [US5] Add explanation field to SearchResult: `pkg/types/search.go` extended with `Explanation` string field
- [ ] T041 [US5] Implement explanation generator: `internal/search/explanation.go` with functions to format score breakdowns (text %, vector %, combined %)
- [ ] T042 [US5] Add --explain flag to search command: `cmd/kbvault/search.go` with flag handling, calls explanation generator
- [ ] T043 [US5] Implement JSON output for explanations: Include `explanation` and score breakdown in JSON response format
- [ ] T044 [US5] Implement human-readable explanation: Format for CLI output (e.g., "Text: 85% | Semantic: 72% → Combined: 79%")

### Unit Tests (Required for US5)

- [ ] T041a Unit tests for explanation: `internal/search/explanation_test.go` - various score combinations, formatting, >80% coverage
- [ ] T042a CLI tests for --explain flag: `cmd/kbvault/search_test.go` extended - flag parsing, explanation output, >70% coverage

### Integration Test (End-to-End for US5)

- [ ] T045 [US5] E2E test: Explanation flow: `specs/006-s3-vector-search/test-e2e/explanation_test.go`
  - Setup: Vault with notes
  - Action: Run `kbvault search "authentication" --explain --json`
  - Assert: Each result has explanation field with text/semantic/combined scores
  - Verify: Scores make mathematical sense (combined = text*weight + semantic*weight)

---

## Phase 8: Polish & Integration

**Goal**: Cross-cutting concerns, performance optimization, documentation, final testing

**Duration**: 1.5 days  
**Blockers**: All Phase 3-7 tasks complete

### Infrastructure & Documentation Tasks

- [ ] T046 Comprehensive integration tests: `specs/006-s3-vector-search/test-e2e/full_workflow_test.go` - complete user journey (new vault → search → similar → vectorize → explain)
- [ ] T047 Performance benchmarks: `pkg/vector/benchmarks_test.go` - Ollama latency, S3 Vectors latency, cache performance (measure against targets)
- [ ] T048 Error handling audit: Verify all error paths (Ollama unavailable, S3 errors, invalid config) with appropriate messages and fallbacks
- [ ] T049 Logging implementation: Add structured logging to all vector operations using context-based logging
- [ ] T050 Documentation: Create `VECTOR_SEARCH_GUIDE.md` with usage examples, configuration, troubleshooting

### Caching & Optimization Tasks

- [ ] T051 Cache optimization: Implement query deduplication (same query text → reuse embedding) for >90% hit rate
- [ ] T052 Memory profiling: Profile cache memory usage with 10K notes, optimize if >50MB
- [ ] T053 Batch size tuning: Benchmark batch sizes 10-100, document optimal values per hardware
- [ ] T054 Timeout handling: Test Ollama timeout behavior, ensure graceful degradation

### Configuration & Validation Tasks

- [ ] T055 Config validation enhancement: Validate all vector config fields on load, provide helpful error messages
- [ ] T056 Profile example configs: Create 3 example profiles (fast, balanced, high-quality) in documentation
- [ ] T057 Default configuration: Ensure all vector fields have sensible defaults, config works out-of-box

### Final Testing & Release Preparation

- [ ] T058 Coverage report: Run `go test -cover ./...`, verify >60% overall (>70% for vector package)
- [ ] T059 Linting pass: Run `golangci-lint run`, fix all warnings (no exceptions without justification)
- [ ] T060 Pre-release checklist: Constitution compliance, backward compatibility, S3 Vector Search integration verified

---

## Task Dependency Graph

```
PHASE 1 (Setup)
├─ T001 Create structure
├─ T002 Add SDK dependency
├─ T003 Test fixtures
├─ T004 Config schema
└─ T005 Test utilities
    ↓
PHASE 2 (Foundational)
├─ T006 [P] Ollama client → T006a tests
├─ T007 [P] Vector types → T007a tests
├─ T008 [P] Similarity math → T008a tests
└─ T009 [P] Cache implementation → T009a tests
    ↓
PHASE 3 (US1: Semantic Search)
├─ T010 [P] [US1] Backend interface → T010a tests
├─ T011 [P] [US1] S3 backend → T011a tests
├─ T012 [P] [US1] Factory
├─ T013 [P] [US1] Result types
├─ T014 [US1] Search engine → T014a tests
├─ T015 [US1] Search command → T015a tests
├─ T016 [US1] Availability check
├─ T017 [US1] Config loading
└─ T018 [US1] E2E test
    ↓
PHASE 4 (US2: Similar Notes)
├─ T019 [US2] Similar search → T019a tests
├─ T020 [P] [US2] Top-K selection → T020a tests
├─ T021 [US2] --similar-to flag → T021a tests
├─ T022 [US2] Note exclusion
├─ T023 [US2] Limit parameter
└─ T024 [US2] E2E test
    ↓
PHASE 5 (US3: Hybrid Search)
├─ T025 [US3] Score weighting → T025a tests
├─ T026 [US3] Hybrid engine → T026a tests
├─ T027 [US3] Weight flags → T027a tests
├─ T028 [US3] Mode detection
├─ T029 [US3] Mode fallback
├─ T030 [US3] Weight validation
└─ T031 [US3] E2E test
    ↓
PHASE 6 (US4: Model Config)
├─ T032 [US4] Vectorize command
├─ T033 [US4] Batch generation → T033a tests
├─ T034 [US4] Model validation → T034a tests
├─ T035 [US4] Model switching → T035a tests
├─ T036 [US4] Remote Ollama
├─ T037 [US4] Progress tracking
├─ T038 [US4] Resume vectorization
└─ T039 [US4] E2E test
    ↓
PHASE 7 (US5: Explanations)
├─ T040 [US5] Explanation field
├─ T041 [US5] Explanation generator → T041a tests
├─ T042 [US5] --explain flag → T042a tests
├─ T043 [US5] JSON explanations
├─ T044 [US5] Human-readable format
└─ T045 [US5] E2E test
    ↓
PHASE 8 (Polish & Release)
├─ T046 Full workflow E2E
├─ T047 Performance benchmarks
├─ T048 Error handling audit
├─ T049 Logging implementation
├─ T050 Documentation
├─ T051 Cache optimization
├─ T052 Memory profiling
├─ T053 Batch size tuning
├─ T054 Timeout handling
├─ T055 Config validation
├─ T056 Example profiles
├─ T057 Default configuration
├─ T058 Coverage report
├─ T059 Linting pass
└─ T060 Pre-release checklist
```

---

## Parallel Execution Examples

### Parallel Execution: Phase 2 (Foundational)
```bash
# Run these 4 tasks in parallel (no dependencies)
Task T006: Ollama client
Task T007: Vector types
Task T008: Similarity math
Task T009: Cache implementation
# All complete before Phase 3 starts
```

### Parallel Execution: Phase 3 (US1 - Semantic Search)
```bash
# Parallel group 1
Task T010: Backend interface
Task T011: S3 backend
Task T012: Factory
Task T013: Result types

# Then serial
Task T014: Search engine (depends on T010-T013)
Task T015: Search command (depends on T014)
Task T016: Availability check
Task T017: Config loading
Task T018: E2E test (depends on all above)
```

### Parallel Execution: Phase 5 & 6 (After US1 Complete)
```bash
# Can start US2 and start setting up US4 in parallel
Phase 4 (US2): T019-T024 (depends on US1 complete)
Phase 6 (US4): T032-T038 (depends on US1 complete)
# Both can progress in parallel, meet at Phase 8
```

---

## Quality Assurance Checklist

### Unit Tests (Target: >75% coverage per task)

- [ ] Vector types and validation
- [ ] Ollama client (with mocking)
- [ ] Similarity algorithms
- [ ] Cache implementation (hit/miss/eviction)
- [ ] Search engines (semantic, similar, hybrid)
- [ ] Score weighting
- [ ] Model validation and switching
- [ ] Explanation formatting

### Integration Tests (Target: >70% coverage)

- [ ] S3 Vectors backend (with mock S3)
- [ ] Ollama integration (with mock server)
- [ ] Search command (CLI parsing + engine calls)
- [ ] Batch vectorization flow
- [ ] Configuration loading and validation

### End-to-End Tests (User Story Coverage)

- [ ] US1 E2E: Semantic search with sample notes
- [ ] US2 E2E: Similar notes discovery
- [ ] US3 E2E: Hybrid search with weight variations
- [ ] US4 E2E: Model switching and re-indexing
- [ ] US5 E2E: Search explanations with score breakdown

### Performance Testing

- [ ] Ollama latency: <300ms per embedding (batch of 50)
- [ ] S3 Vectors query: <200ms for 1000-note vault
- [ ] Cache hit rate: >90% on repeated queries
- [ ] Memory usage: <50MB cache for 10,000 notes
- [ ] Overall search latency: <2s end-to-end

### Code Quality

- [ ] Go fmt: All files pass `gofmt`
- [ ] Linting: `golangci-lint run` with 0 errors
- [ ] Coverage: >60% overall, >70% for vector package
- [ ] Documentation: All public functions documented
- [ ] Error messages: Clear and actionable

### Backward Compatibility

- [ ] Text-only search still works (semantic disabled)
- [ ] Existing notes unaffected (vector metadata optional)
- [ ] Config changes are additive (old configs still valid)
- [ ] API contracts unchanged (no breaking CLI changes)

---

## Suggested MVP Scope (Minimum Viable Product)

For a quick release focusing on core value (semantic search):

**MVP = Phase 1 + Phase 2 + Phase 3 (US1 Only)**

Tasks to include:
- T001-T005: Setup
- T006-T009: Foundational components
- T010-T018: Full US1 (Semantic Search) + tests + E2E
- T046-T048: Error handling and basic documentation

**Deliverable**: `kbvault search --semantic "query"` works end-to-end

**Estimated time**: 3-4 days  
**Meets acceptance**: Yes, US1 is independently testable and valuable

Then add remaining user stories incrementally (US2, US3, US4, US5).

---

## File Path Reference

### New Files Created

```text
pkg/vector/
├── backend.go                    (T010)
├── similarity.go                 (T008)
├── topk.go                       (T020)
├── scoring.go                    (T025)
├── cache/
│   ├── cache.go                  (T009)
│   └── cache_test.go             (T009a)
├── ollama/
│   ├── client.go                 (T006)
│   ├── client_test.go            (T006a)
│   └── availability.go           (T016)
├── s3/
│   ├── backend.go                (T011)
│   └── backend_test.go           (T011a)
├── factory.go                    (T012)
├── testutil/
│   └── fixtures.go               (T005)
└── benchmarks_test.go            (T047)

pkg/types/
├── vector.go                     (T007)
└── search.go extended            (T013, T040)

internal/search/
├── semantic.go                   (T014)
├── semantic_test.go              (T014a)
├── similar.go                    (T019)
├── similar_test.go               (T019a)
├── engine.go extended            (T026)
├── engine_test.go extended       (T026a)
├── explanation.go                (T041)
├── explanation_test.go           (T041a)
└── batch.go                      (T033)

internal/vector/
├── batch.go                      (T033)
├── batch_test.go                 (T033a)
├── model_switch.go               (T035)
├── model_switch_test.go          (T035a)
└── model_validation.go           (T034)

cmd/kbvault/
├── search.go extended            (T015, T021, T027, T042)
├── search_test.go extended       (T015a, T021a, T027a, T042a)
├── vectorize.go                  (T032)
└── vectorize_test.go

pkg/config/
├── config.go extended            (T004, T017)
└── config_test.go extended

Test Fixtures & E2E:
├── specs/006-s3-vector-search/
│   ├── test-data/
│   │   ├── sample-notes.json     (T003)
│   │   └── expected-vectors.json (T003)
│   └── test-e2e/
│       ├── semantic_test.go      (T018)
│       ├── similar_test.go       (T024)
│       ├── hybrid_test.go        (T031)
│       ├── model_config_test.go  (T039)
│       ├── explanation_test.go   (T045)
│       └── full_workflow_test.go (T046)

Documentation:
├── VECTOR_SEARCH_GUIDE.md        (T050)
└── configs/
    ├── example-fast-profile.toml (T056)
    ├── example-balanced-profile.toml
    └── example-quality-profile.toml
```

### Modified Files

```text
go.mod                             (T002)
pkg/config/config.go              (T004, T017)
pkg/types/search.go               (T013, T040)
internal/search/engine.go         (T026)
cmd/kbvault/search.go             (T015, T021, T027, T042)
```

---

## Task Completion Criteria

Each task is complete when:

1. **Code written**: Implementation matches task description
2. **Tests pass**: Unit/integration tests for that task pass (if applicable)
3. **No lint errors**: `golangci-lint run` shows no violations
4. **Documented**: Public functions have comments, complex logic explained
5. **Backward compatible**: No breaking changes to existing APIs (unless E2E task)
6. **Error handling**: All error paths handled with clear messages

---

## Next Steps

1. **Start Phase 1**: Create directory structure and add dependencies (T001-T005)
2. **Proceed to Phase 2**: Implement foundational components in parallel (T006-T009)
3. **Follow dependency graph**: Respect serial dependencies between phases
4. **Test continuously**: Run tests after each task, don't accumulate test debt
5. **Commit frequently**: Commit after each task or task group completion
6. **Track progress**: Update this file as tasks are completed

---

**Total Implementation Time Estimate**: 8-10 days (sequential)  
**With parallelization**: 5-6 days  
**MVP (US1 Only)**: 3-4 days
