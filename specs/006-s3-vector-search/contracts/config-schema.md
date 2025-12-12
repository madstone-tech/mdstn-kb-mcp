# Configuration Schema: Vector Search

**Status**: Design Phase 1  
**Version**: 1.0.0  
**Feature**: Session 6 - S3 Vector Search & Semantic Capabilities

## Overview

TOML configuration schema for vector search settings in kbVault profiles.

---

## Configuration Structure

### TOML Section: `[profiles.<name>.vector]`

Vector search settings are scoped to individual profiles, allowing different configurations per vault.

**File Location**: `~/.config/kbvault/config.toml`

**Example**:
```toml
[profiles.research]
name = "Research Vault"
storage_backend = "s3"
storage_path = "s3://mybucket/research"

[profiles.research.vector]
semantic_enabled = true
text_weight = 0.7
vector_weight = 0.3
similarity_threshold = 0.7
ollama_model = "nomic-embed-text"
ollama_endpoint = "http://localhost:11434"
auto_regenerate = "on_save"
batch_size = 50
request_timeout = "30s"
cache_enabled = true
cache_max_size = 10000
cache_ttl = "24h"
```

---

## Field Definitions

### Core Features

#### `semantic_enabled` (boolean)
Enable or disable semantic search for this profile.

**Default**: `true`  
**Valid Values**: `true`, `false`  
**Type**: Boolean

**Behavior**:
- `true`: Semantic search available; embeddings generated on note save
- `false`: Text-only search; semantic features disabled

**Example**:
```toml
semantic_enabled = true
```

**Notes**:
- When false, all other vector settings are ignored
- Embeddings are retained (can be re-enabled later)

---

#### `text_weight` (float32)
Weight for text matching in hybrid search (0.0-1.0).

**Default**: `0.7`  
**Valid Range**: 0.0 to 1.0  
**Type**: Float

**Interpretation**:
- `0.7` = 70% text relevance, 30% semantic
- `0.0` = pure vector/semantic search
- `1.0` = pure text search (no vector component)
- Typically 0.5-0.9 for balanced hybrid search

**Example**:
```toml
text_weight = 0.7         # 70% text, 30% vector
text_weight = 0.5         # 50% text, 50% vector (equal)
text_weight = 0.0         # 0% text, 100% vector (pure semantic)
text_weight = 1.0         # 100% text, 0% vector (text-only)
```

**Validation**:
- Must be >= 0.0 and <= 1.0
- Often used with `vector_weight = 1.0 - text_weight`

---

#### `vector_weight` (float32)
Weight for semantic matching in hybrid search (0.0-1.0).

**Default**: `0.3`  
**Valid Range**: 0.0 to 1.0  
**Type**: Float

**Interpretation**:
- Combined with `text_weight`: `vector_weight = 1.0 - text_weight`
- `0.3` = 30% semantic, 70% text (default)
- Adjustable for different search priorities

**Example**:
```toml
vector_weight = 0.3       # 30% vector, 70% text
vector_weight = 0.5       # 50% vector, 50% text
vector_weight = 1.0       # 100% vector, 0% text
```

**Validation**:
- Must be >= 0.0 and <= 1.0
- Sum of `text_weight + vector_weight` should equal 1.0 (auto-normalized if not)

---

#### `similarity_threshold` (float32)
Minimum similarity score for results (0.0-1.0).

**Default**: `0.7`  
**Valid Range**: 0.0 to 1.0  
**Type**: Float

**Interpretation**:
- `0.7` = exclude results with <70% semantic similarity (industry standard)
- `0.5` = more permissive, returns more results
- `0.85+` = strict, only very similar notes
- Applied to vector similarity scores during filtering

**Common Values**:
| Threshold | Use Case |
|---|---|
| 0.85+ | High precision, narrow results |
| 0.70-0.80 | Balanced (default) |
| 0.50-0.70 | High recall, broader results |
| <0.50 | Very permissive (rare) |

**Example**:
```toml
similarity_threshold = 0.7        # Default: balanced
similarity_threshold = 0.85       # Strict: high precision
similarity_threshold = 0.5        # Loose: high recall
```

**Validation**:
- Must be >= 0.0 and <= 1.0
- Typical range 0.5-0.9

---

### Ollama Configuration

#### `ollama_model` (string)
Ollama embedding model to use.

**Default**: `"nomic-embed-text"`  
**Valid Values**: Installed Ollama models (see `kbvault vectorize --list-models`)  
**Type**: String

**Supported Models**:
- `nomic-embed-text` (384 dims, fast, default)
- `mxbai-embed-large` (1024 dims, high quality)
- `jina-embeddings-v2-base-en` (768 dims, good quality)
- Any other installed Ollama embedding model

**Example**:
```toml
ollama_model = "nomic-embed-text"        # Fast, lightweight (default)
ollama_model = "mxbai-embed-large"       # High quality, slower
ollama_model = "jina-embeddings-v2-base-en"  # Medium quality
```

**Validation**:
- Must be installed in local Ollama
- Error if not found: run `ollama pull <model-name>`
- Changing model triggers full re-indexing (expensive operation)

**Notes**:
- Model dimensions vary (384, 768, 1024, etc.)
- Switching models requires regenerating all embeddings
- Each model has different speed/quality trade-offs

---

#### `ollama_endpoint` (string, URL)
HTTP endpoint for Ollama service.

**Default**: `"http://localhost:11434"`  
**Valid Format**: Valid HTTP(S) URL  
**Type**: String (URL)

**Examples**:
```toml
ollama_endpoint = "http://localhost:11434"           # Local (default)
ollama_endpoint = "http://192.168.1.100:11434"       # Local network
ollama_endpoint = "https://ollama.example.com:443"   # Remote HTTPS
ollama_endpoint = "http://remote-host:11434"         # Remote HTTP
```

**Validation**:
- Must start with `http://` or `https://`
- Must have valid hostname/IP and port
- Port must be 1-65535
- Tested on first search (error if unreachable)

**Behavior**:
- If unreachable: Semantic search falls back to text-only (graceful degradation)
- Supports both local and remote Ollama instances
- Remote instances useful for shared infrastructure

---

### Embedding Generation

#### `auto_regenerate` (string, enum)
When and how to regenerate embeddings automatically.

**Default**: `"on_save"`  
**Valid Values**: `"on_save"`, `"manual"`, `"scheduled"`  
**Type**: String (enum)

**Modes**:

| Mode | Behavior | Use Case |
|---|---|---|
| `"on_save"` | Generate embedding when note is saved/modified | Default, real-time updates |
| `"manual"` | Only regenerate via `kbvault vectorize` command | Large vaults, batch processing |
| `"scheduled"` | Regenerate on schedule (e.g., nightly) | Planned indexing |

**Example**:
```toml
auto_regenerate = "on_save"          # Immediate (default)
auto_regenerate = "manual"           # User-triggered
auto_regenerate = "scheduled"        # Planned updates
```

**Behavior**:
- `"on_save"`: 50-500ms latency per save (Ollama time)
- `"manual"`: No latency on save; user controls batch timing
- `"scheduled"`: Run nightly/weekly; separate operation

**Notes**:
- Only applies to individual note saves
- Vault-wide re-indexing always uses `kbvault vectorize`

---

#### `batch_size` (integer)
Texts per Ollama API request during batch embedding.

**Default**: `50`  
**Valid Range**: 1 to 1000  
**Type**: Integer

**Interpretation**:
- `50` = embed 50 note contents per Ollama request
- Larger batches: better throughput, higher latency per batch
- Smaller batches: lower latency, more API calls

**Trade-offs**:
| Size | Latency | Throughput | Notes |
|---|---|---|---|
| 10 | Low (~100ms) | Slow (~10/sec) | For interactive use |
| 50 | Medium (~500ms) | Medium (~100/sec) | Balanced (default) |
| 100 | High (~1000ms) | Fast (~200/sec) | For batch operations |

**Example**:
```toml
batch_size = 10            # Small: lower latency
batch_size = 50            # Medium: balanced (default)
batch_size = 100           # Large: higher throughput
```

**Validation**:
- Must be >= 1 and <= 1000
- Recommended 10-100 range
- Too large (>200): may cause Ollama to slow down

**Notes**:
- Affects only batch operations (`kbvault vectorize`)
- Single note saves use batch_size=1 implicitly

---

#### `request_timeout` (duration)
Timeout per Ollama API request.

**Default**: `"30s"` (30 seconds)  
**Valid Format**: Go duration string (e.g., "30s", "1m", "500ms")  
**Type**: Duration

**Examples**:
```toml
request_timeout = "30s"              # Default: 30 seconds
request_timeout = "1m"               # 1 minute (for slow hardware)
request_timeout = "500ms"            # 500 milliseconds (fast machine)
```

**Typical Values**:
| Duration | Hardware | Batch Size |
|---|---|---|
| 30s | Standard laptop | 50 (default) |
| 60s | Slow/old hardware | 25 |
| 10s | High-performance CPU | 100 |

**Validation**:
- Must be valid Go duration (e.g., "500ms", "30s", "1m30s")
- Minimum recommended: 5s (Ollama first load)
- Too short (<5s): requests will timeout frequently

**Notes**:
- First Ollama request in session slower (model load)
- Subsequent requests faster (model cached in RAM)
- Adjust based on hardware and batch size

---

### Caching

#### `cache_enabled` (boolean)
Enable in-memory embedding cache.

**Default**: `true`  
**Valid Values**: `true`, `false`  
**Type**: Boolean

**Behavior**:
- `true`: Cache embeddings in memory; >90% hit rate typical
- `false`: No caching; every query generates new embedding

**Example**:
```toml
cache_enabled = true         # Enable cache (default)
cache_enabled = false        # Disable cache (debugging)
```

**Performance Impact**:
- With cache: First search ~200ms, repeated searches <1ms
- Without cache: Every search ~150-300ms (Ollama call)

**Notes**:
- Cache is in-memory (not persisted)
- Cleared on kbvault restart
- Query deduplication combined with cache yields >90% hit rate

---

#### `cache_max_size` (integer)
Maximum number of embeddings to cache.

**Default**: `10000`  
**Valid Range**: 100 to 1000000  
**Type**: Integer

**Interpretation**:
- `10000` = cache up to 10,000 unique query embeddings
- Uses LRU eviction (least recently used removed first)
- Each entry ~1.6KB (384 dims × 4 bytes + overhead)

**Memory Usage**:
| Size | Memory | Duration |
|---|---|---|
| 1,000 | ~1.6MB | Small vaults |
| 10,000 | ~16MB | Medium vaults (default) |
| 100,000 | ~160MB | Large vaults |

**Example**:
```toml
cache_max_size = 1000          # Small: ~1.6MB
cache_max_size = 10000         # Medium: ~16MB (default)
cache_max_size = 100000        # Large: ~160MB
```

**Validation**:
- Must be >= 100
- Recommended 1000-100000
- Too large: wastes memory; too small: low hit rate

**Notes**:
- LRU eviction: oldest unused entries removed first
- Hit rate typically >90% for typical vault workflows

---

#### `cache_ttl` (duration)
Time-to-live for cache entries (auto-invalidation).

**Default**: `"24h"` (24 hours)  
**Valid Format**: Go duration string  
**Type**: Duration

**Examples**:
```toml
cache_ttl = "24h"              # 1 day (default)
cache_ttl = "0"                # No expiry (keep indefinitely)
cache_ttl = "1h"               # 1 hour (frequent refresh)
```

**Use Cases**:
| TTL | Scenario |
|---|---|
| `"0"` (never) | Static vault, don't change embeddings |
| `"1h"` | Frequently changing notes |
| `"24h"` | Daily refresh (default) |
| `"7d"` | Weekly refresh for large vaults |

**Validation**:
- Must be valid Go duration
- `"0"` = no expiry
- Common values: 1h, 24h, 7d

**Notes**:
- Expired entries still take up cache space (not freed immediately)
- Cache cleared on Ollama model change
- TTL prevents stale embeddings if vault changes

---

## Full Example Configuration

```toml
[profiles.research]
name = "Research Vault"
storage_backend = "s3"
storage_path = "s3://mybucket/research"

[profiles.research.vector]
# Feature toggle
semantic_enabled = true

# Hybrid search weighting
text_weight = 0.7
vector_weight = 0.3

# Filtering
similarity_threshold = 0.7

# Ollama configuration
ollama_model = "nomic-embed-text"
ollama_endpoint = "http://localhost:11434"

# Embedding generation
auto_regenerate = "on_save"
batch_size = 50
request_timeout = "30s"

# Caching
cache_enabled = true
cache_max_size = 10000
cache_ttl = "24h"
```

---

## Alternative Profiles

### High-Performance Profile (Fast Hardware)

```toml
[profiles.work]
storage_backend = "s3"

[profiles.work.vector]
semantic_enabled = true
text_weight = 0.5              # Equal weighting
vector_weight = 0.5
similarity_threshold = 0.75    # Stricter filtering
ollama_model = "mxbai-embed-large"  # Higher quality
ollama_endpoint = "http://localhost:11434"
auto_regenerate = "on_save"
batch_size = 100               # Larger batches
request_timeout = "10s"        # Faster hardware
cache_max_size = 100000        # Larger cache
cache_ttl = "12h"              # Shorter TTL
```

### Low-Latency Profile (Text-Heavy)

```toml
[profiles.quick]
storage_backend = "s3"

[profiles.quick.vector]
semantic_enabled = true
text_weight = 0.9              # 90% text, 10% vector
vector_weight = 0.1
similarity_threshold = 0.5     # Loose filtering (more results)
ollama_model = "nomic-embed-text"  # Fast model
ollama_endpoint = "http://localhost:11434"
auto_regenerate = "manual"     # Batch vectorization only
batch_size = 10                # Small batches for speed
request_timeout = "5s"
cache_enabled = true
cache_max_size = 50000         # Large cache
cache_ttl = "24h"
```

### Conservative Profile (Large Vaults)

```toml
[profiles.archive]
storage_backend = "s3"

[profiles.archive.vector]
semantic_enabled = true
text_weight = 0.8              # Text-primary
vector_weight = 0.2
similarity_threshold = 0.8     # Strict filtering
ollama_model = "nomic-embed-text"  # Lightweight
ollama_endpoint = "http://localhost:11434"
auto_regenerate = "scheduled"  # Weekly regeneration
batch_size = 25                # Small batches
request_timeout = "60s"        # Conservative timeout
cache_enabled = true
cache_max_size = 5000          # Small cache
cache_ttl = "7d"               # Weekly refresh
```

---

## Configuration Validation

All configurations are validated on load:

| Field | Validation | Error |
|---|---|---|
| `semantic_enabled` | Boolean | Type error |
| `text_weight` | 0.0-1.0 | Value out of range |
| `vector_weight` | 0.0-1.0 | Value out of range |
| `similarity_threshold` | 0.0-1.0 | Value out of range |
| `ollama_model` | Model exists in Ollama | Model not found |
| `ollama_endpoint` | Valid URL | URL parse error |
| `auto_regenerate` | enum (on_save, manual, scheduled) | Invalid value |
| `batch_size` | 1-1000 | Value out of range |
| `request_timeout` | Valid duration | Duration parse error |
| `cache_max_size` | 100+ | Value too small |
| `cache_ttl` | Valid duration | Duration parse error |

---

## Implementation Notes

- Configuration loaded via existing Viper system (pkg/config/)
- Defaults applied if keys missing
- Profile-scoped settings (can differ per vault)
- Changes take effect on next search/operation (no reload needed)
- Changing model triggers full re-indexing
- Cache cleared on semantic_enabled toggle or model change
