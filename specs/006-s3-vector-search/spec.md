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
2. **Given** semantic search is enabled with OpenAI embeddings, **When** user searches for "machine learning approaches", **Then** results include notes on "deep learning", "neural networks", "supervised learning" ranked by semantic similarity
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

### User Story 4 - Configurable Embedding Provider (Priority: P2)

A privacy-conscious team wants to use local embeddings instead of cloud providers. A team on a budget wants cheaper embeddings. Users should be able to configure which embedding provider their vault uses (OpenAI, HuggingFace, local models) and switch between them.

**Why this priority**: Supports different organizational requirements - privacy, cost, compliance. Enables flexibility in choosing trade-offs.

**Independent Test**: Can be fully tested by: configuring vault with different embedding providers, verifying semantic search works with each provider, and confirming embeddings can be regenerated when provider changes.

**Acceptance Scenarios**:

1. **Given** vault configured with OpenAI embeddings, **When** user switches to HuggingFace provider via `kbvault config set vector.provider "huggingface"`, **Then** system can regenerate embeddings with new provider
2. **Given** local model as embedding provider, **When** semantic search executes, **Then** embeddings are generated locally without external API calls
3. **Given** multiple embedding providers available, **When** user specifies provider in config, **Then** subsequent embedding generation uses specified provider

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
- **FR-002**: System MUST support multiple embedding providers (OpenAI, HuggingFace, local models) with configuration per profile
- **FR-003**: System MUST implement hybrid search combining TF-IDF text search results with vector similarity results, with configurable weighting
- **FR-004**: Users MUST be able to find similar notes using `kbvault search --similar-to <note-id>` command
- **FR-005**: System MUST support configurable similarity threshold for semantic search results (default 0.7)
- **FR-006**: System MUST store embeddings alongside notes in S3 Vector Search backend
- **FR-007**: Users MUST be able to configure embedding generation behavior (auto, manual, scheduled) per profile
- **FR-008**: System MUST provide relevance scoring for search results (0-1 scale) with explanation of score composition
- **FR-009**: System MUST support batch embedding generation for efficient processing of large vaults
- **FR-010**: System MUST handle embedding provider changes by supporting regeneration of all embeddings with new provider
- **FR-011**: System MUST support incremental embedding updates when notes are added or modified
- **FR-012**: Users MUST be able to specify search result ranking strategy (text-only, vector-only, or hybrid with configurable weights)
- **FR-013**: System MUST maintain backward compatibility with text-only search (semantic search is optional)
- **FR-014**: System MUST cache generated embeddings to avoid redundant API calls
- **FR-015**: System MUST support natural language search queries without special syntax

### Key Entities

- **Vector Embedding**: Numerical representation (vector) of note content generated by embedding model, stored alongside note metadata
  - Attributes: dimensions (1536 for OpenAI), provider (openai/huggingface/local), confidence score
  - Relationships: associated with single Note, updated when Note changes

- **Search Result**: Unified search result combining text match data with semantic match data
  - Attributes: note_id, relevance_score (0-1), text_match_score, semantic_match_score, explanation, ranking_position
  - Relationships: returned from Search Query, multiple per query

- **Search Configuration**: Per-profile settings controlling how semantic search behaves
  - Attributes: semantic_enabled (bool), text_weight (0-1), vector_weight (0-1), similarity_threshold (0-1), embedding_provider, auto_regenerate
  - Relationships: belongs to Profile

- **Embedding Provider**: Service that generates embeddings
  - Attributes: name (openai/huggingface/local), model_name, dimensions, cost_tier, api_endpoint
  - Relationships: configured in Search Configuration, used by Embedding Generation

## Success Criteria *(mandatory)*

<!--
  ACTION REQUIRED: Define measurable success criteria.
  These must be technology-agnostic and measurable.
-->

### Measurable Outcomes

- **SC-001**: Users can perform semantic searches and receive results in under 2 seconds for vaults with 1000+ notes
- **SC-002**: Semantic search returns relevant results (conceptually related notes) with minimum 80% accuracy as verified by user testing
- **SC-003**: Hybrid search with text and semantic components increases search result relevance by minimum 40% compared to text-only search (measured through user satisfaction surveys)
- **SC-004**: System can generate embeddings for 1000-note vault within 5 minutes using batch processing (including API calls)
- **SC-005**: Users can switch embedding providers and regenerate vault embeddings without data loss
- **SC-006**: System maintains >90% embedding cache hit rate during repeated searches to minimize API calls
- **SC-007**: Similarity search returns top 5 related notes in under 1 second
- **SC-008**: At least 85% of users can successfully perform semantic search without additional training or documentation
- **SC-009**: Search result explanations help users understand relevance (measured by user feedback: >80% report explanations are "clear" or "very clear")
- **SC-010**: Semantic search works correctly with text_weight/vector_weight configurations from 0/100 to 100/0 (inclusive)
- **SC-011**: System gracefully degrades to text-only search when embedding provider unavailable (no errors, meaningful fallback)
- **SC-012**: All new code maintains >60% test coverage requirement

---

## Assumptions

- S3 Vector Search backend infrastructure is available (depends on Session 5 completion)
- Multiple embedding APIs are accessible (OpenAI, HuggingFace) or local model infrastructure available
- Note content is primarily text-based (images, media handled as metadata)
- Search performance targets are for vaults up to 10,000 notes (larger vaults may require additional optimization)
- Embeddings are regenerated when configuration changes (not updated incrementally)
- Default embedding model is OpenAI text-embedding-3-small (1536 dimensions, cost-effective)
- Users have appropriate API keys/credentials for configured embedding providers

---

## Dependencies

- ✅ **Session 4**: Search & Content Management (text search engine, search interface)
- ⏳ **Session 5**: Viper Profiles & S3 Storage (profile configuration, S3 backend)
- ✅ **Existing**: CLI infrastructure (command routing, flags, configuration loading)
- ✅ **Existing**: AWS SDK (for S3 Vector Search integration)

---

## Out of Scope (Session 6 doesn't include)

- Advanced ML model fine-tuning on vault-specific data
- Multi-language embedding support beyond what base models provide
- Custom vector search algorithms (uses S3 Vector Search as-is)
- Distributed/parallel processing across multiple machines
- Explicit feedback loops for embedding model improvement
- Integration with large language models (LLMs) for query augmentation - this is Session 8 (MCP)
