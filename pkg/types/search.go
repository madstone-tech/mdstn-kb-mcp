// Package types defines core data types for the knowledge base vault system.
package types

import "time"

// SemanticSearchRequest represents a semantic search query
type SemanticSearchRequest struct {
	Query     string            `json:"query"`
	Limit     int               `json:"limit,omitempty"`
	Threshold float64           `json:"threshold,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// HybridSearchRequest combines text and semantic search
type HybridSearchRequest struct {
	Query        string            `json:"query"`
	TextWeight   float64           `json:"text_weight,omitempty"`
	VectorWeight float64           `json:"vector_weight,omitempty"`
	Limit        int               `json:"limit,omitempty"`
	Threshold    float64           `json:"threshold,omitempty"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// SemanticSearchResult represents a single semantic search result
type SemanticSearchResult struct {
	NoteID    string    `json:"note_id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Score     float64   `json:"score"`
	Distance  float64   `json:"distance,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// HybridSearchResult represents a hybrid search result with both text and vector scores
type HybridSearchResult struct {
	NoteID        string    `json:"note_id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	TextScore     float64   `json:"text_score"`
	VectorScore   float64   `json:"vector_score"`
	CombinedScore float64   `json:"combined_score"`
	TextWeight    float64   `json:"text_weight"`
	VectorWeight  float64   `json:"vector_weight"`
	Explanation   string    `json:"explanation,omitempty"`
	Timestamp     time.Time `json:"timestamp"`
}

// SearchMode represents the type of search being performed
type SearchMode string

const (
	SearchModeText     SearchMode = "text"
	SearchModeSemantic SearchMode = "semantic"
	SearchModeHybrid   SearchMode = "hybrid"
)

// SearchOptions configures search behavior
type SearchOptions struct {
	Mode           SearchMode
	Limit          int
	Threshold      float64
	TextWeight     float64
	VectorWeight   float64
	IncludeScore   bool
	IncludeContent bool
	Explain        bool
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Mode:           SearchModeText,
		Limit:          10,
		Threshold:      0.7,
		TextWeight:     0.7,
		VectorWeight:   0.3,
		IncludeScore:   true,
		IncludeContent: true,
		Explain:        false,
	}
}

// ResultExplanation provides detailed explanation of search score
type ResultExplanation struct {
	Query          string    `json:"query"`
	NoteID         string    `json:"note_id"`
	TextMatchScore float64   `json:"text_match_score,omitempty"`
	SemanticScore  float64   `json:"semantic_score,omitempty"`
	CombinedScore  float64   `json:"combined_score"`
	TextWeight     float64   `json:"text_weight,omitempty"`
	VectorWeight   float64   `json:"vector_weight,omitempty"`
	Details        string    `json:"details,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
}

// ValidationError represents a search validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidateSemanticSearchRequest validates a semantic search request
func (r *SemanticSearchRequest) Validate() []ValidationError {
	var errors []ValidationError

	if r.Query == "" {
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query cannot be empty",
		})
	}

	if r.Limit <= 0 {
		r.Limit = 10
	} else if r.Limit > 1000 {
		errors = append(errors, ValidationError{
			Field:   "limit",
			Message: "limit cannot exceed 1000",
		})
	}

	if r.Threshold < 0 || r.Threshold > 1 {
		r.Threshold = 0.7 // Reset to default
	}

	return errors
}

// ValidateHybridSearchRequest validates a hybrid search request
func (r *HybridSearchRequest) Validate() []ValidationError {
	var errors []ValidationError

	if r.Query == "" {
		errors = append(errors, ValidationError{
			Field:   "query",
			Message: "query cannot be empty",
		})
	}

	if r.Limit <= 0 {
		r.Limit = 10
	} else if r.Limit > 1000 {
		errors = append(errors, ValidationError{
			Field:   "limit",
			Message: "limit cannot exceed 1000",
		})
	}

	if r.TextWeight < 0 || r.TextWeight > 1 {
		errors = append(errors, ValidationError{
			Field:   "text_weight",
			Message: "text_weight must be between 0 and 1",
		})
	}

	if r.VectorWeight < 0 || r.VectorWeight > 1 {
		errors = append(errors, ValidationError{
			Field:   "vector_weight",
			Message: "vector_weight must be between 0 and 1",
		})
	}

	totalWeight := r.TextWeight + r.VectorWeight
	if totalWeight > 1.0 && totalWeight != 0 {
		r.TextWeight = r.TextWeight / totalWeight
		r.VectorWeight = r.VectorWeight / totalWeight
	}

	if r.Threshold < 0 || r.Threshold > 1 {
		r.Threshold = 0.7
	}

	return errors
}
