# API Contract: Embedding Operations

**Status**: Design Phase 1  
**Version**: 1.0.0  
**Feature**: Session 6 - S3 Vector Search & Semantic Capabilities

## Overview

Embedding API contracts for generating, managing, and refreshing note embeddings using Ollama and S3 Vectors.

---

## 1. Vectorize Vault (Batch Embedding Generation)

### Command: `kbvault vectorize [flags]`

Generates or regenerates embeddings for all notes in a vault.

**CLI Signature**:
```bash
kbvault vectorize [flags]
```

**Flags**:
```
  --batch-size       Texts per Ollama batch (default: 50)
  --force            Force regeneration even if embeddings exist
  --profile          Profile name (default: "default")
  --model            Override Ollama model (default: from config)
  --dry-run          Show what would be done without executing
  --progress         Show progress updates (default: true)
  --json             Output as JSON
```

**Request Format** (Programmatic):
```go
type VectorizeRequest struct {
	BatchSize      int    `json:"batch_size,omitempty"`
	Force          bool   `json:"force,omitempty"`
	ProfileName    string `json:"profile,omitempty"`
	Model          string `json:"model,omitempty"`
	DryRun         bool   `json:"dry_run,omitempty"`
}
```

**Response Format** (JSON):
```json
{
  "operation": "vectorize_vault",
  "profile": "default",
  "model": "nomic-embed-text",
  "dimensions": 384,
  "status": "completed",
  "summary": {
    "total_notes": 1500,
    "vectorized": 1500,
    "skipped": 0,
    "failed": 0
  },
  "timing": {
    "start_time": "2024-12-12T10:00:00Z",
    "end_time": "2024-12-12T10:15:30Z",
    "total_duration_seconds": 930,
    "average_ms_per_note": 0.62
  },
  "batches": {
    "total": 30,
    "completed": 30,
    "failed": 0
  },
  "storage": {
    "backend": "s3_vectors",
    "bucket": "kbvault-vectors-123456",
    "index": "default",
    "total_vectors_stored": 1500
  },
  "performance": {
    "ollama_requests": 30,
    "ollama_total_duration_ms": 28000,
    "average_ms_per_ollama_request": 933
  }
}
```

**Human-Readable Output** (Default CLI):
```
Vectorizing vault (profile: default, model: nomic-embed-text)...

Processing 1500 notes in batches of 50...
[████████████████████████████] 100% (1500/1500)

✓ Vectorized 1500 notes
✓ Duration: 15m 30s
✓ Average: 0.62ms per note
✓ Stored in S3 Vectors index: default

Summary:
  Total:    1500 notes
  Success:  1500
  Skipped:  0
  Failed:   0
  Batches:  30/30 completed
```

**Progress Output** (Real-time during execution):
```
[▓░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] 50/1500 (3%) - Batch 1/30 - 45ms
[▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░] 450/1500 (30%) - Batch 9/30 - 850ms
[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓░░░░░] 1350/1500 (90%) - Batch 27/30 - 14m 15s
[████████████████████████████████] 1500/1500 (100%) - Complete
```

**Error Responses**:
```json
{
  "error": "vectorization_failed",
  "code": "BATCH_FAILED",
  "message": "Batch 15 failed after 3 retries",
  "details": {
    "batch_number": 15,
    "notes_in_batch": 50,
    "start_note_id": "01HQXYZ123456789ABCDE15",
    "last_error": "context deadline exceeded",
    "retries": 3
  },
  "partial_results": {
    "notes_vectorized_before_failure": 700,
    "notes_pending": 800
  },
  "recommendation": "Increase timeout or reduce batch size and retry"
}
```

**Resume on Failure**:
```bash
# If vectorization is interrupted, resume from last successful batch
$ kbvault vectorize --resume
```

**Implementation Notes**:
- Splits vault into batches for efficient processing
- Tracks progress and can resume from last successful batch
- Verifies embeddings exist in S3 Vectors
- Skips notes that are not text-based (e.g., images)
- Respects `--force` flag to overwrite existing embeddings
- Logs all successes and failures for debugging

---

## 2. Vectorize Single Note

### Command: `kbvault show <note-id> --vectorize` (implicit)

Generates embedding for a single note automatically when saving.

**Automatic Behavior**:
- On `kbvault new`: Auto-vectorizes if `auto_regenerate = "on_save"`
- On `kbvault edit`: Auto-updates embedding if content changed
- Returns immediately after embedding generation

**Programmatic Request**:
```go
type SingleNoteVectorizeRequest struct {
	NoteID      string `json:"note_id"`
	ProfileName string `json:"profile,omitempty"`
	Force       bool   `json:"force,omitempty"`
}
```

**Response** (Embedded in note response):
```json
{
  "note": {
    "id": "01HQXYZ123456789ABCDEF",
    "title": "...",
    "content": "...",
    "vector": {
      "model": "nomic-embed-text",
      "dimensions": 384,
      "generated_at": "2024-12-12T10:30:00Z",
      "indexed": true,
      "embedding_duration_ms": 145
    }
  }
}
```

---

## 3. Embedding Status

### Command: `kbvault vectorize --status [flags]`

Shows embedding status for vault or specific notes.

**CLI Signature**:
```bash
kbvault vectorize --status [flags]
```

**Flags**:
```
  --profile          Profile name (default: "default")
  --not-vectorized   Show only non-vectorized notes
  --json             Output as JSON
```

**Response Format** (JSON):
```json
{
  "vault_status": {
    "profile": "default",
    "total_notes": 1500,
    "vectorized_notes": 1490,
    "non_vectorized_notes": 10,
    "embedding_model": "nomic-embed-text",
    "embedding_dimensions": 384,
    "index_size_mb": 22.5,
    "last_vectorization": "2024-12-12T10:00:00Z"
  },
  "non_vectorized": [
    {
      "note_id": "01HQXYZ123456789ABCDE1",
      "title": "Image-only Note",
      "reason": "No text content"
    },
    {
      "note_id": "01HQXYZ123456789ABCDE2",
      "title": "Draft Note",
      "reason": "Note created before vectorization enabled"
    }
  ]
}
```

**Human-Readable Output** (Default CLI):
```
Vault Vectorization Status (profile: default)

Model:          nomic-embed-text
Dimensions:     384
Total Notes:    1500
Vectorized:     1490 (99.3%)
Not Vectorized: 10 (0.7%)
Index Size:     22.5 MB
Last Updated:   2024-12-12T10:00:00Z

Non-Vectorized Notes:
  1. Image-only Note (01HQXYZ1234...) - No text content
  2. Draft Note (01HQXYZ1234...) - Created before vectorization enabled

To vectorize all notes: kbvault vectorize --force
```

---

## 4. Model Management

### Command: `kbvault vectorize --list-models`

Lists available Ollama embedding models.

**Response Format** (JSON):
```json
{
  "available_models": [
    {
      "name": "nomic-embed-text",
      "dimensions": 384,
      "inference_speed": "fast",
      "quality": "good",
      "status": "installed",
      "size_mb": 274
    },
    {
      "name": "mxbai-embed-large",
      "dimensions": 1024,
      "inference_speed": "medium",
      "quality": "excellent",
      "status": "not_installed",
      "size_mb": 669,
      "install_command": "ollama pull mxbai-embed-large"
    },
    {
      "name": "jina-embeddings-v2-base-en",
      "dimensions": 768,
      "inference_speed": "medium",
      "quality": "excellent",
      "status": "not_installed",
      "size_mb": 800,
      "install_command": "ollama pull jina-embeddings-v2-base-en"
    }
  ],
  "current_config": {
    "profile": "default",
    "model": "nomic-embed-text",
    "endpoint": "http://localhost:11434"
  }
}
```

**Human-Readable Output** (Default CLI):
```
Available Embedding Models

INSTALLED:
  ✓ nomic-embed-text
    └─ Dimensions: 384 | Speed: Fast | Quality: Good | Size: 274MB

NOT INSTALLED:
  ○ mxbai-embed-large
    └─ Dimensions: 1024 | Speed: Medium | Quality: Excellent | Size: 669MB
    └─ Install: ollama pull mxbai-embed-large

  ○ jina-embeddings-v2-base-en
    └─ Dimensions: 768 | Speed: Medium | Quality: Excellent | Size: 800MB
    └─ Install: ollama pull jina-embeddings-v2-base-en

Current Config (profile: default):
  Model:    nomic-embed-text
  Endpoint: http://localhost:11434
```

---

## 5. Switch Embedding Model

### Command: `kbvault vectorize --switch-model <model-name>`

Changes the embedding model and regenerates all embeddings.

**CLI Signature**:
```bash
kbvault vectorize --switch-model <model-name> [flags]
```

**Flags**:
```
  --profile          Profile name (default: "default")
  --dry-run          Show what would be done without executing
  --batch-size       Texts per batch (default: 50)
```

**Request Format** (Programmatic):
```go
type SwitchModelRequest struct {
	NewModel    string `json:"new_model"`
	ProfileName string `json:"profile,omitempty"`
	DryRun      bool   `json:"dry_run,omitempty"`
	BatchSize   int    `json:"batch_size,omitempty"`
}
```

**Response Format** (JSON):
```json
{
  "operation": "switch_model",
  "profile": "default",
  "old_model": "nomic-embed-text",
  "new_model": "mxbai-embed-large",
  "status": "completed",
  "migration": {
    "notes_to_reindex": 1500,
    "old_dimensions": 384,
    "new_dimensions": 1024,
    "action": "delete_old_vectors_and_reindex"
  },
  "summary": {
    "total_reindexed": 1500,
    "failed": 0
  },
  "timing": {
    "start_time": "2024-12-12T10:00:00Z",
    "end_time": "2024-12-12T10:30:00Z",
    "total_duration_seconds": 1800
  }
}
```

**Process**:
1. Validate new model is installed
2. Show what will change (--dry-run)
3. Delete old vectors from S3 Vectors
4. Regenerate embeddings with new model
5. Update configuration

---

## 6. Configure Ollama Endpoint

### Command: `kbvault config set vector.ollama_endpoint <url>`

Changes Ollama endpoint (local or remote).

**Examples**:
```bash
# Use local Ollama (default)
$ kbvault config set vector.ollama_endpoint "http://localhost:11434"

# Use remote Ollama instance
$ kbvault config set vector.ollama_endpoint "http://remote-host.example.com:11434"

# HTTPS with auth (if supported)
$ kbvault config set vector.ollama_endpoint "https://user:pass@ollama.example.com"
```

**Validation**:
- Endpoint must be reachable (tested on first search)
- Protocol must be http or https
- Port must be valid (1-65535)

**Response** (Configuration updated):
```json
{
  "updated": true,
  "field": "vector.ollama_endpoint",
  "old_value": "http://localhost:11434",
  "new_value": "http://remote-host:11434",
  "verification": {
    "endpoint_reachable": true,
    "response_time_ms": 45,
    "available_models": ["nomic-embed-text", "mxbai-embed-large"]
  }
}
```

---

## 7. Embedding Cache Management

### Command: `kbvault cache --clear-embeddings [flags]`

Clears in-memory embedding cache.

**CLI Signature**:
```bash
kbvault cache --clear-embeddings [flags]
```

**Flags**:
```
  --profile          Profile name (default: "default")
  --stats            Show cache statistics before clearing
```

**Response Format** (JSON):
```json
{
  "operation": "clear_cache",
  "profile": "default",
  "before": {
    "cache_size": 5000,
    "memory_usage_mb": 7.5
  },
  "after": {
    "cache_size": 0,
    "memory_usage_mb": 0
  },
  "status": "cleared"
}
```

**Cache Statistics** (--stats):
```bash
$ kbvault cache --clear-embeddings --stats
```

**Output**:
```json
{
  "cache_stats": {
    "total_entries": 5000,
    "memory_usage_mb": 7.5,
    "hit_rate": 0.92,
    "hits": 4600,
    "misses": 400,
    "oldest_entry_age": "24h",
    "newest_entry_age": "2m"
  }
}
```

---

## Common Response Structure

All embedding API responses follow this structure:

```json
{
  "operation": "vectorize_vault|switch_model|status|...",
  "profile": "default",
  "status": "pending|in_progress|completed|failed",
  "summary": {...},
  "timing": {...},
  "error": null
}
```

---

## Error Handling

### Common Errors

| Scenario | Code | Message | Action |
|---|---|---|---|
| Ollama not running | `OLLAMA_UNAVAILABLE` | "Cannot connect to Ollama" | Prompt to start Ollama |
| Model not installed | `MODEL_NOT_FOUND` | "Model 'xyz' not found" | Suggest install command |
| No text content | `NO_TEXT` | "Note has no text content" | Skip note, continue |
| Batch timeout | `BATCH_TIMEOUT` | "Batch processing timed out" | Retry with smaller batch |
| S3 write failed | `S3_ERROR` | "Failed to store vectors" | Retry or check AWS credentials |

### Retry Strategy

- Transient errors (timeout, network): Retry with exponential backoff (3 retries)
- Permanent errors (not found, permission): Fail immediately with helpful message
- Partial failures: Log failed notes and continue (don't block entire operation)

---

## Implementation Checklist

- [ ] Parse vectorize command and flags in Cobra
- [ ] Validate batch size (1-1000) and timeout (>0)
- [ ] Check Ollama connectivity before starting
- [ ] Support batch processing with progress tracking
- [ ] Handle Ollama unavailability gracefully
- [ ] Store embeddings in S3 Vectors with metadata
- [ ] Cache embeddings in memory with LRU eviction
- [ ] Support model switching with re-indexing
- [ ] Provide resumable operations for large vaults
- [ ] Return consistent JSON and human-readable formats
- [ ] Log all operations for debugging
- [ ] Support --dry-run for preview before execution
