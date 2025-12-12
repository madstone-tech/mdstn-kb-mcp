# API Contract: Search Operations

**Status**: Design Phase 1  
**Version**: 1.0.0  
**Feature**: Session 6 - S3 Vector Search & Semantic Capabilities

## Overview

Search API contracts for semantic, text, and hybrid search operations. Interfaces are CLI-based (Cobra commands) with JSON output support.

---

## 1. Semantic Search

### Command: `kbvault search --semantic <query>`

Performs semantic search using note embeddings and Ollama.

**CLI Signature**:
```bash
kbvault search --semantic <query> [flags]
```

**Flags**:
```
  --limit           Max results (default: 10)
  --threshold       Min similarity (0.0-1.0, default: 0.7)
  --explain         Show relevance breakdown
  --json            Output as JSON
  --profile         Profile name (default: "default")
```

**Request Format** (Programmatic):
```go
type SemanticSearchRequest struct {
	Query          string  `json:"query"`
	Limit          int     `json:"limit,omitempty"`
	Threshold      float32 `json:"threshold,omitempty"`
	Explain        bool    `json:"explain,omitempty"`
	ProfileName    string  `json:"profile,omitempty"`
}
```

**Response Format** (JSON):
```json
{
  "query": "distributed systems concepts",
  "search_type": "semantic",
  "total_results": 3,
  "results": [
    {
      "rank": 1,
      "note_id": "01HQXYZ123456789ABCDEF",
      "title": "Distributed Systems Fundamentals",
      "preview": "A comprehensive guide to distributed systems...",
      "relevance_score": 0.92,
      "explanation": "Semantic match: 92%"
    },
    {
      "rank": 2,
      "note_id": "01HQXYZ123456789ABCDE2",
      "title": "Consensus Algorithms",
      "preview": "Understanding Raft, Paxos, and PBFT...",
      "relevance_score": 0.78,
      "explanation": "Semantic match: 78%"
    },
    {
      "rank": 3,
      "note_id": "01HQXYZ123456789ABCDE3",
      "title": "Byzantine Fault Tolerance",
      "preview": "Exploring Byzantine-resistant systems...",
      "relevance_score": 0.71,
      "explanation": "Semantic match: 71%"
    }
  ],
  "execution": {
    "duration_ms": 145,
    "embedding_cache_hit": true
  }
}
```

**Error Responses**:
```json
{
  "error": "semantic search unavailable",
  "code": "OLLAMA_UNAVAILABLE",
  "message": "Ollama not running at http://localhost:11434",
  "fallback": "Using text-only search instead"
}
```

**Human-Readable Output** (Default CLI):
```
Semantic Search Results for: "distributed systems concepts"

1. Distributed Systems Fundamentals
   └─ Semantic: 92% | 01HQXYZ12345...
   └─ A comprehensive guide to distributed systems...

2. Consensus Algorithms
   └─ Semantic: 78% | 01HQXYZ12345...
   └─ Understanding Raft, Paxos, and PBFT...

3. Byzantine Fault Tolerance
   └─ Semantic: 71% | 01HQXYZ12345...
   └─ Exploring Byzantine-resistant systems...

(3 results in 145ms, cache hit)
```

**Implementation Notes**:
- Query is embedded using Ollama (cached if enabled)
- Vector similarity computed via cosine distance
- Results filtered by threshold
- Top K results returned, sorted by score descending

---

## 2. Find Similar Notes

### Command: `kbvault search --similar-to <note-id>`

Finds notes semantically similar to a given note.

**CLI Signature**:
```bash
kbvault search --similar-to <note-id> [flags]
```

**Flags**:
```
  --limit           Max similar notes (default: 10)
  --threshold       Min similarity (0.0-1.0, default: 0.7)
  --exclude-self    Exclude the target note (default: true)
  --json            Output as JSON
  --profile         Profile name (default: "default")
```

**Request Format** (Programmatic):
```go
type SimilarNotesRequest struct {
	NoteID       string  `json:"note_id"`
	Limit        int     `json:"limit,omitempty"`
	Threshold    float32 `json:"threshold,omitempty"`
	ExcludeSelf  bool    `json:"exclude_self,omitempty"`
	ProfileName  string  `json:"profile,omitempty"`
}
```

**Response Format** (JSON):
```json
{
  "target_note_id": "01HQXYZ123456789ABCDEF",
  "target_title": "REST API Design",
  "search_type": "similar",
  "total_similar": 5,
  "results": [
    {
      "rank": 1,
      "note_id": "01HQXYZ123456789ABCDE2",
      "title": "HTTP Methods and Status Codes",
      "preview": "Comprehensive guide to HTTP semantics...",
      "similarity_score": 0.89
    },
    {
      "rank": 2,
      "note_id": "01HQXYZ123456789ABCDE3",
      "title": "Web Service Architecture",
      "preview": "Building scalable web services...",
      "similarity_score": 0.76
    }
  ],
  "execution": {
    "duration_ms": 89,
    "embedding_cached": true
  }
}
```

**Error Responses**:
```json
{
  "error": "note not found",
  "code": "NOTE_NOT_FOUND",
  "message": "Note 01HQXYZ123456789ABCDEF does not exist in vault"
}
```

**Human-Readable Output** (Default CLI):
```
Similar Notes to: REST API Design (01HQXYZ12345...)

1. HTTP Methods and Status Codes
   └─ Similarity: 89% | 01HQXYZ12345...

2. Web Service Architecture
   └─ Similarity: 76% | 01HQXYZ12345...

3. API Authentication Patterns
   └─ Similarity: 72% | 01HQXYZ12345...

(5 similar notes in 89ms, cache hit)
```

**Implementation Notes**:
- Retrieves embedding of target note
- Compares against all note embeddings using cosine similarity
- Excludes target note by default (unless --exclude-self=false)
- Results filtered by threshold and limited to top K

---

## 3. Hybrid Search

### Command: `kbvault search <query>`

Performs hybrid search combining text and semantic results (default behavior).

**CLI Signature**:
```bash
kbvault search <query> [flags]
```

**Flags**:
```
  --text-only       Use text search only
  --vector-only     Use semantic search only
  --hybrid          Use hybrid search (default)
  --text-weight     Text weight 0.0-1.0 (default: 0.7)
  --vector-weight   Vector weight 0.0-1.0 (default: 0.3)
  --limit           Max results (default: 10)
  --threshold       Min relevance (0.0-1.0, default: 0.7)
  --explain         Show score breakdown
  --json            Output as JSON
  --profile         Profile name (default: "default")
```

**Request Format** (Programmatic):
```go
type HybridSearchRequest struct {
	Query         string  `json:"query"`
	SearchMode    string  `json:"search_mode"` // "text", "vector", "hybrid"
	TextWeight    float32 `json:"text_weight,omitempty"`
	VectorWeight  float32 `json:"vector_weight,omitempty"`
	Limit         int     `json:"limit,omitempty"`
	Threshold     float32 `json:"threshold,omitempty"`
	Explain       bool    `json:"explain,omitempty"`
	ProfileName   string  `json:"profile,omitempty"`
}
```

**Response Format** (JSON):
```json
{
  "query": "authentication",
  "search_type": "hybrid",
  "search_config": {
    "text_weight": 0.7,
    "vector_weight": 0.3
  },
  "total_results": 4,
  "results": [
    {
      "rank": 1,
      "note_id": "01HQXYZ123456789ABCDEF",
      "title": "Authentication Mechanisms",
      "preview": "Overview of auth patterns...",
      "relevance_score": 0.91,
      "text_match_score": 0.95,
      "semantic_match_score": 0.82,
      "explanation": "Text: 95% × 0.7 = 66.5% | Semantic: 82% × 0.3 = 24.6% → Combined: 91%"
    },
    {
      "rank": 2,
      "note_id": "01HQXYZ123456789ABCDE2",
      "title": "Authorization Strategies",
      "preview": "Role-based and attribute-based access control...",
      "relevance_score": 0.74,
      "text_match_score": 0.60,
      "semantic_match_score": 0.87,
      "explanation": "Text: 60% × 0.7 = 42% | Semantic: 87% × 0.3 = 26.1% → Combined: 68%"
    }
  ],
  "execution": {
    "duration_ms": 234,
    "text_search_duration_ms": 15,
    "semantic_search_duration_ms": 145,
    "ranking_duration_ms": 74,
    "cache_hits": 1
  }
}
```

**Human-Readable Output** (Default CLI):
```
Hybrid Search Results for: "authentication"
(70% Text, 30% Vector)

1. Authentication Mechanisms
   └─ Combined: 91% (Text: 95% | Semantic: 82%) | 01HQXYZ12345...
   └─ Overview of auth patterns...

2. Authorization Strategies
   └─ Combined: 74% (Text: 60% | Semantic: 87%) | 01HQXYZ12345...
   └─ Role-based and attribute-based access control...

3. OAuth and OIDC
   └─ Combined: 68% (Text: 70% | Semantic: 64%) | 01HQXYZ12345...
   └─ Understanding OAuth 2.0 and OIDC flows...

(4 results in 234ms, cache: 1 hit)
```

**Weighting Examples**:

| Config | Use Case | Result |
|---|---|---|
| 100% text, 0% vector | Exact keyword matching | Only text matches |
| 70% text, 30% vector | Balanced (default) | Both exact and conceptual |
| 50% text, 50% vector | Equal weighting | No bias |
| 0% text, 100% vector | Pure semantic | Only conceptual match |

**Implementation Notes**:
- Text search uses existing full-text engine (Session 4)
- Semantic search uses embeddings and cosine similarity
- Scores normalized to 0-1 range before combination
- Final rank determined by combined score
- Can operate in text-only, vector-only, or hybrid mode
- Default uses profile configuration

---

## 4. Search with Explanation

### Flag: `--explain`

When used with any search command, provides detailed relevance scoring breakdown.

**Example** (Semantic Search with Explanation):
```bash
$ kbvault search --semantic "consensus algorithms" --explain --json
```

**Response** (with explanation details):
```json
{
  "query": "consensus algorithms",
  "results": [
    {
      "note_id": "01HQXYZ123456789ABCDEF",
      "title": "Raft Consensus",
      "relevance_score": 0.88,
      "semantic_match_score": 0.88,
      "explanation": {
        "query_embedding_dims": 384,
        "query_embedding_normalized": true,
        "similarity_metric": "cosine",
        "score_breakdown": "Cosine similarity: 0.88 (88%)",
        "confidence": "High"
      }
    },
    {
      "note_id": "01HQXYZ123456789ABCDE2",
      "title": "Byzantine Fault Tolerance",
      "relevance_score": 0.74,
      "semantic_match_score": 0.74,
      "explanation": {
        "score_breakdown": "Cosine similarity: 0.74 (74%)",
        "confidence": "Medium"
      }
    }
  ]
}
```

**Example** (Hybrid Search with Explanation):
```bash
$ kbvault search "authentication" --explain --json
```

**Response** (detailed score composition):
```json
{
  "results": [
    {
      "note_id": "01HQXYZ123456789ABCDEF",
      "title": "Authentication Mechanisms",
      "relevance_score": 0.91,
      "explanation": {
        "text_score": 0.95,
        "text_weight": 0.7,
        "text_contribution": 0.665,
        "semantic_score": 0.82,
        "vector_weight": 0.3,
        "semantic_contribution": 0.246,
        "final_score": 0.911,
        "components": [
          "Text: 95% (exact keyword match for 'authentication')",
          "Semantic: 82% (related to auth concepts)",
          "Weighted sum: (0.95 × 0.7) + (0.82 × 0.3) = 0.91"
        ]
      }
    }
  ]
}
```

---

## Common Response Structure

All search API responses follow this structure:

```json
{
  "query": "...",
  "search_type": "semantic|hybrid|similar|text",
  "total_results": 0,
  "results": [...],
  "execution": {
    "duration_ms": 0,
    "cache_hits": 0,
    "cache_misses": 0
  },
  "fallback": "..." // Optional: if Ollama unavailable
}
```

---

## Error Handling

### Common Errors

| Scenario | Code | Message | Action |
|---|---|---|---|
| Ollama unavailable | `OLLAMA_UNAVAILABLE` | "Cannot connect to Ollama" | Fallback to text search |
| Invalid note ID | `NOTE_NOT_FOUND` | "Note does not exist" | Return error |
| Invalid threshold | `INVALID_THRESHOLD` | "Threshold must be 0.0-1.0" | Reject request |
| Vault empty | `NO_NOTES` | "Vault contains no notes" | Return empty results |
| Embedding failed | `EMBEDDING_ERROR` | "Failed to generate embedding" | Fallback or retry |

### Graceful Degradation

If Ollama is unavailable:
1. Semantic search falls back to text search
2. Response includes `"fallback": "text_search"` field
3. User is informed via message
4. Search completes with text results only

```json
{
  "query": "...",
  "search_type": "hybrid",
  "fallback": "text_search (Ollama unavailable at http://localhost:11434)",
  "warning": "Semantic search unavailable; using text-only results",
  "results": [...]
}
```

---

## Implementation Checklist

- [ ] Parse search flags in Cobra command
- [ ] Validate request parameters (threshold, weights, limit)
- [ ] Check if semantic search enabled in profile config
- [ ] Handle Ollama unavailability gracefully
- [ ] Cache embeddings to avoid redundant Ollama calls
- [ ] Format responses as JSON or human-readable text
- [ ] Return error codes and messages consistently
- [ ] Support all weighting combinations (0-100%)
- [ ] Explain relevance scores when requested
