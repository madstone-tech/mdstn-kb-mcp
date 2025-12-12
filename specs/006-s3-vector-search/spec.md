# Feature Specification: S3 Vector Search & Semantic Capabilities

**Feature Branch**: `006-s3-vector-search`  
**Created**: December 12, 2024  
**Status**: Draft  
**Input**: GitHub Issue #6: S3 Vector Search & Semantic Capabilities

## User Scenarios & Testing *(mandatory)*

<!--
  IMPORTANT: User stories should be PRIORITIZED as user journeys ordered by importance.
  Each user story/journey must be INDEPENDENTLY TESTABLE - meaning if you implement just ONE of them,
  you should still have a viable MVP (Minimum Viable Product) that delivers value.
  
  Assign priorities (P1, P2, P3, etc.) to each story, where P1 is the most critical.
  Think of each story as a standalone slice of functionality that can be:
  - Developed independently
  - Tested independently
  - Deployed independently
  - Demonstrated to users independently
-->

### User Story 1 - Semantic Search with Natural Language (Priority: P1)

A researcher wants to find all notes about "distributed systems concepts" without having to remember exact keywords. They should be able to type a natural language query and get semantically related notes back, even if the notes use different terminology (e.g., notes using "microservices", "consensus algorithms", "fault tolerance").

**Why this priority**: Core value proposition - enables users to find knowledge through semantic understanding, not just keyword matching. This is the fundamental feature that justifies vector search.

**Independent Test**: Can be fully tested by: setting up semantic search capability, ingesting test notes with different terminology, performing a semantic query, and verifying that conceptually related notes are returned regardless of exact wording.

**Acceptance Scenarios**:

1. **Given** a vault with notes on "microservices", "distributed consensus", "fault-tolerant systems", **When** user searches semantically for "distributed systems", **Then** all three notes are returned in results with relevance scores
2. **Given** semantic search is enabled and Ollama embeddings are generated locally, **When** user searches for "machine learning approaches", **Then** results include notes on "deep learning", "neural networks", "supervised learning" ranked by semantic similarity
3. **Given** a semantic search query returns multiple results, **When** user views results, **Then** each result shows relevance score (0-1) indicating semantic match strength

---

### User Story 2 - Find Similar Notes (Priority: P1)

A user is editing a note and wants to find other notes that discuss similar concepts. They click "More like this" or use the similar-to command to discover related knowledge they may have forgotten about or want to cross-reference.

**Why this priority**: Directly supports knowledge discovery and prevents duplicate work. Enables serendipitous discoveries of related concepts.

**Independent Test**: Can be fully tested by: taking any note in the vault, running a "similar notes" search, and verifying that conceptually related notes are returned ranked by similarity.

**Acceptance Scenarios**:

1. **Given** a note about "REST APIs", **When** user searches for similar notes using `--similar-to "note-id-123"`, **Then** system returns notes on "HTTP methods", "API design", "web services" ranked by similarity
2. **Given** similar notes command, **When** user specifies `--limit 5`, **Then** exactly 5 most similar notes are returned
3. **Given** similar notes command without limit, **When** executed, **Then** system returns reasonable default number of results (10-20) to prevent overwhelming user

---

### User Story 3 - Hybrid Search with Text and Semantic Results (Priority: P1)

A user wants best-of-both-worlds search: keyword accuracy combined with semantic understanding. They search for a term and get results ranked by both text relevance and semantic similarity, with configurable weighting based on their search mode.

**Why this priority**: Provides flexibility for different search scenarios - sometimes exact keywords matter most, sometimes conceptual relevance matters most.

**Independent Test**: Can be fully tested by: performing a hybrid search, verifying that results are ranked using both text and semantic scores, and confirming that result order reflects the configured text/vector weight split.

**Acceptance Scenarios**:

1. **Given** hybrid search mode enabled with 70% text weight, 30% vector weight, **When** user searches "authentication", **Then** results include both exact keyword matches and semantically related notes like "authorization", "security", "access control"
2. **Given** hybrid search configuration, **When** user adjusts text_weight to 90%, **Then** search results prioritize exact keyword matches over semantic similarity
3. **Given** search without mode specified, **When** executed, **Then** system defaults to hybrid search combining both text and vector results

---

### User Story 4 - Configurable Ollama Models (Priority: P2)

A user wants to optimize embeddings for their specific needs - some prefer fast/small models for speed, others prefer higher quality embeddings. Users should be able to configure which Ollama model their vault uses and easily switch between available local models.

**Why this priority**: Enables users to trade off between speed and quality based on their local hardware and use case. Supports both local and remote Ollama instances for flexibility.

**Independent Test**: Can be fully tested by: configuring vault with different Ollama models, verifying semantic search works with each model, and confirming embeddings can be regenerated when model changes.

**Acceptance Scenarios**:

1. **Given** vault configured with `nomic-embed-text` model, **When** user switches to another fast/small model via `kbvault config set vector.model "mxbai-embed-large"`, **Then** system can regenerate embeddings with new model
2. **Given** local Ollama instance running, **When** semantic search executes, **Then** embeddings are generated locally without external API calls
3. **Given** user wants to use remote Ollama instance, **When** user configures `kbvault config set vector.ollama_endpoint "http://remote-host:11434"`, **Then** subsequent embedding generation uses remote instance

---

### User Story 5 - Search Explanation and Relevance Scoring (Priority: P2)

A user wants to understand why a note was returned in their search results. They should see relevance scores and explanations showing how the result matches their query - whether through exact keywords, semantic similarity, or metadata.

**Why this priority**: Builds user confidence in search results and helps refine search techniques. Supports better search debugging.

**Independent Test**: Can be fully tested by: performing searches, displaying result explanations, and verifying that users can understand and potentially dispute relevance scores.

**Acceptance Scenarios**:

1. **Given** search with `--explain` flag, **When** results displayed, **Then** each result shows explanation indicating "Text match: 85%, Semantic match: 72%, Overall: 78%"
2. **Given** semantic search result, **When** user requests explanation, **Then** system shows which concepts matched and their similarity confidence
3. **Given** hybrid search result, **When** displayed, **Then** user can see breakdown of text vs semantic contribution to final score

---

### User Story 6 - Batch Embedding Generation for Large Vaults (Priority: P3)

For vaults with hundreds or thousands of notes, users want embeddings to be generated efficiently in batches rather than one at a time, with progress tracking and ability to resume if interrupted.

**Why this priority**: Necessary for production use with large knowledge bases. Prevents system degradation during initial embedding generation.

**Independent Test**: Can be fully tested by: ingesting large number of notes, triggering batch embedding generation, tracking progress, and verifying all notes receive embeddings.

**Acceptance Scenarios**:

1. **Given** vault with 500+ notes, **When** user runs `kbvault vectorize --batch-size 50`, **Then** embeddings are generated 50 at a time with progress updates
2. **Given** batch embedding process, **When** batch completes, **Then** system persists progress and can resume from last successful batch if interrupted
3. **Given** batch operation in progress, **When** user monitors execution, **Then** system displays "Generated 150/500 embeddings (30%)"

### Edge Cases

- What happens when embedding provider API is unavailable? (System should degrade to text-only search gracefully)
- What if user vault contains notes with no readable text content (e.g., images, metadata only)? (Skip embedding generation, use metadata/filename for search)
- What if user has inconsistent embedding dimensions across notes (e.g., mixed OpenAI and HuggingFace)? (System should handle gracefully or alert user to regenerate)
- What happens when semantic search returns no results above configured similarity threshold? (Return empty results with helpful message, suggest lowering threshold)
- How does system handle updates to notes - should embeddings be automatically regenerated? (Configurable: auto-regenerate on save, manual refresh, scheduled updates)

## Requirements *(mandatory)*

<!--
  ACTION REQUIRED: The content in this section represents placeholders.
  Fill them out with the right functional requirements.
-->

### Functional Requirements

- **FR-001**: System MUST support semantic search that returns notes based on conceptual similarity, not just keyword matching
- **FR-002**: System MUST support Ollama as the embedding provider with configurable model selection and support for both local and remote Ollama instances (default: local `http://localhost:11434`)
- **FR-003**: System MUST implement hybrid search combining TF-IDF text search results with vector similarity results, with configurable weighting
- **FR-004**: Users MUST be able to find similar notes using `kbvault search --similar-to <note-id>` command
- **FR-005**: System MUST support configurable similarity threshold for semantic search results (default 0.7)
- **FR-006**: System MUST store embeddings alongside notes in S3 Vector Search backend
- **FR-007**: Users MUST be able to configure embedding generation behavior (auto, manual, scheduled) per profile
- **FR-008**: System MUST provide relevance scoring for search results (0-1 scale) with explanation of score composition
- **FR-009**: System MUST support batch embedding generation for efficient processing of large vaults
- **FR-010**: System MUST handle Ollama model changes by supporting regeneration of all embeddings with new model
- **FR-011**: System MUST support incremental embedding updates when notes are added or modified
- **FR-012**: Users MUST be able to specify search result ranking strategy (text-only, vector-only, or hybrid with configurable weights)
- **FR-013**: System MUST maintain backward compatibility with text-only search (semantic search is optional)
- **FR-014**: System MUST cache generated embeddings to avoid redundant API calls
- **FR-015**: System MUST support natural language search queries without special syntax

### Key Entities

- **Vector Embedding**: Numerical representation (vector) of note content generated by Ollama embedding model, stored alongside note metadata
  - Attributes: dimensions (varies by model: 384 for nomic-embed-text), ollama_model, confidence_score, generated_at
  - Relationships: associated with single Note, updated when Note changes or model changes

- **Search Result**: Unified search result combining text match data with semantic match data
  - Attributes: note_id, relevance_score (0-1), text_match_score, semantic_match_score, explanation, ranking_position
  - Relationships: returned from Search Query, multiple per query

- **Search Configuration**: Per-profile settings controlling how semantic search behaves
  - Attributes: semantic_enabled (bool), text_weight (0-1), vector_weight (0-1), similarity_threshold (0-1), ollama_model (default: nomic-embed-text), ollama_endpoint (default: http://localhost:11434), auto_regenerate
  - Relationships: belongs to Profile

- **Ollama Model**: Embedding model available in Ollama runtime
  - Attributes: model_name (nomic-embed-text, mxbai-embed-large, etc.), dimensions, inference_speed (relative), quality_tier, local_only (bool)
  - Relationships: configured in Search Configuration, used for Embedding Generation

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: Users can perform semantic searches and receive results in under 2 seconds for vaults with 1000+ notes (on standard laptop hardware)
- **SC-002**: Semantic search returns relevant results (conceptually related notes) with minimum 80% accuracy as verified by user testing
- **SC-003**: Hybrid search with text and semantic components increases search result relevance by minimum 40% compared to text-only search (measured through user satisfaction surveys)
- **SC-004**: System can generate embeddings for 1000-note vault efficiently using batch processing (local Ollama, optimized for standard hardware, expected ~10-15 minutes for nomic-embed-text)
- **SC-005**: Users can switch Ollama models and regenerate vault embeddings without data loss
- **SC-006**: System maintains >90% embedding cache hit rate during repeated searches to minimize Ollama inference calls
- **SC-007**: Similarity search returns top 5 related notes in under 1 second (on standard laptop hardware)
- **SC-008**: At least 85% of users can successfully perform semantic search without additional training or documentation
- **SC-009**: Search result explanations help users understand relevance (measured by user feedback: >80% report explanations are "clear" or "very clear")
- **SC-010**: Semantic search works correctly with text_weight/vector_weight configurations from 0/100 to 100/0 (inclusive)
- **SC-011**: System gracefully degrades to text-only search when embedding provider unavailable (no errors, meaningful fallback)
- **SC-012**: All new code maintains >60% test coverage requirement

---

## Assumptions

- S3 Vector Search backend infrastructure is available (depends on Session 5 completion)
- Ollama is installed and running locally by default on `http://localhost:11434` (configurable for remote instances)
- Default Ollama model is `nomic-embed-text` (384 dimensions, fast, free, small footprint suitable for standard hardware)
- Note content is primarily text-based (images, media handled as metadata)
- Search performance targets are optimized for standard laptop hardware (e.g., 8GB RAM, modern CPU), not high-end servers
- Embeddings are regenerated when model/configuration changes (not updated incrementally)
- Users will have Ollama installed as a prerequisite; installation instructions provided in setup docs
- Supported fast/small/free Ollama models include: nomic-embed-text, mxbai-embed-large, and similar lightweight models
- Remote Ollama instance support is available but local instance is the primary/recommended use case

---

## Dependencies

- ✅ **Session 4**: Search & Content Management (text search engine, search interface)
- ⏳ **Session 5**: Viper Profiles & S3 Storage (profile configuration, S3 backend)
- ✅ **Existing**: CLI infrastructure (command routing, flags, configuration loading)
- ✅ **Existing**: AWS SDK (for S3 Vector Search integration)

---

## Out of Scope (Session 6 doesn't include)

- Support for cloud embedding providers (OpenAI, HuggingFace, etc.) - planned for future session
- Advanced ML model fine-tuning on vault-specific data
- Multi-language embedding support beyond what base models provide
- Custom vector search algorithms (uses S3 Vector Search as-is)
- Distributed/parallel processing across multiple machines
- Explicit feedback loops for embedding model improvement
- Integration with large language models (LLMs) for query augmentation - this is Session 8 (MCP)
- GPU-specific optimizations or high-performance computing scenarios
