package types

import (
	"time"
)

// Note represents a single knowledge base note
type Note struct {
	// ID is the ULID-based unique identifier
	ID string `json:"id" toml:"id"`

	// Title is the human-readable note title
	Title string `json:"title" toml:"title"`

	// Content is the markdown body of the note
	Content string `json:"content" toml:"content"`

	// Frontmatter contains structured metadata
	Frontmatter Frontmatter `json:"frontmatter" toml:"frontmatter"`

	// FilePath is the relative path from vault root
	FilePath string `json:"file_path" toml:"file_path"`

	// StorageBackend indicates where this note is stored
	StorageBackend StorageType `json:"storage_backend" toml:"storage_backend"`

	// CreatedAt is when the note was first created
	CreatedAt time.Time `json:"created_at" toml:"created_at"`

	// UpdatedAt is when the note was last modified
	UpdatedAt time.Time `json:"updated_at" toml:"updated_at"`

	// Size is the content size in bytes
	Size int64 `json:"size" toml:"size"`
}

// Frontmatter represents the YAML/TOML metadata at the top of notes
type Frontmatter struct {
	// ID mirrors the note ID for consistency
	ID string `json:"id" yaml:"id" toml:"id"`

	// Title mirrors the note title
	Title string `json:"title" yaml:"title" toml:"title"`

	// Tags for categorization and filtering
	Tags []string `json:"tags" yaml:"tags" toml:"tags"`

	// Type indicates the kind of note (note, daily, template, etc.)
	Type string `json:"type" yaml:"type" toml:"type"`

	// Created timestamp in ISO format
	Created string `json:"created" yaml:"created" toml:"created"`

	// Updated timestamp in ISO format
	Updated string `json:"updated" yaml:"updated" toml:"updated"`

	// Storage backend for this note
	Storage string `json:"storage" yaml:"storage" toml:"storage"`

	// Template used to create this note
	Template string `json:"template,omitempty" yaml:"template,omitempty" toml:"template,omitempty"`

	// Custom metadata fields
	Custom map[string]interface{} `json:"custom,omitempty" yaml:"custom,omitempty" toml:"custom,omitempty"`
}

// NoteMetadata represents lightweight note information for listings
type NoteMetadata struct {
	ID             string      `json:"id"`
	Title          string      `json:"title"`
	Tags           []string    `json:"tags"`
	Type           string      `json:"type"`
	FilePath       string      `json:"file_path"`
	StorageBackend StorageType `json:"storage_backend"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
	Size           int64       `json:"size"`
}

// CreateNoteRequest represents a request to create a new note
type CreateNoteRequest struct {
	// Title for the new note (required)
	Title string `json:"title" validate:"required,max=200"`

	// Content for the note body (optional)
	Content string `json:"content,omitempty" validate:"max=10485760"` // 10MB limit

	// Tags to assign (optional)
	Tags []string `json:"tags,omitempty"`

	// Template to use (optional, defaults to "default")
	Template string `json:"template,omitempty"`

	// Type of note (optional, defaults to "note")
	Type string `json:"type,omitempty"`

	// Custom frontmatter fields (optional)
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// UpdateNoteRequest represents a request to update an existing note
type UpdateNoteRequest struct {
	// ID of the note to update (required)
	ID string `json:"id" validate:"required"`

	// Title update (optional)
	Title *string `json:"title,omitempty" validate:"omitempty,max=200"`

	// Content update (optional)
	Content *string `json:"content,omitempty" validate:"omitempty,max=10485760"` // 10MB limit

	// Tags update (optional)
	Tags []string `json:"tags,omitempty"`

	// Type update (optional)
	Type *string `json:"type,omitempty"`

	// Custom frontmatter updates (optional)
	Custom map[string]interface{} `json:"custom,omitempty"`
}

// Link represents a connection between notes
type Link struct {
	// SourceID is the note containing the link
	SourceID string `json:"source_id"`

	// TargetID is the note being linked to
	TargetID string `json:"target_id"`

	// TargetTitle is the resolved title of the target note
	TargetTitle string `json:"target_title"`

	// LinkText is the display text of the link
	LinkText string `json:"link_text"`

	// Position is the character position in the source note
	Position int `json:"position"`

	// IsValid indicates if the target exists
	IsValid bool `json:"is_valid"`

	// Type of link (wiki, markdown, etc.)
	Type LinkType `json:"type"`
}

// LinkType represents the style of link
type LinkType string

const (
	LinkTypeWiki     LinkType = "wiki"     // [[Note Title]]
	LinkTypeMarkdown LinkType = "markdown" // [Link Text](note-id)
	LinkTypeID       LinkType = "id"       // [[note-id]]
)

// ToMetadata converts a Note to NoteMetadata
func (n *Note) ToMetadata() NoteMetadata {
	return NoteMetadata{
		ID:             n.ID,
		Title:          n.Title,
		Tags:           n.Frontmatter.Tags,
		Type:           n.Frontmatter.Type,
		FilePath:       n.FilePath,
		StorageBackend: n.StorageBackend,
		CreatedAt:      n.CreatedAt,
		UpdatedAt:      n.UpdatedAt,
		Size:           n.Size,
	}
}

// HasTag checks if the note has a specific tag
func (n *Note) HasTag(tag string) bool {
	for _, t := range n.Frontmatter.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

// AddTag adds a tag to the note if it doesn't already exist
func (n *Note) AddTag(tag string) {
	if !n.HasTag(tag) {
		n.Frontmatter.Tags = append(n.Frontmatter.Tags, tag)
	}
}

// RemoveTag removes a tag from the note
func (n *Note) RemoveTag(tag string) {
	for i, t := range n.Frontmatter.Tags {
		if t == tag {
			n.Frontmatter.Tags = append(n.Frontmatter.Tags[:i], n.Frontmatter.Tags[i+1:]...)
			break
		}
	}
}

// IsDaily returns true if this is a daily note
func (n *Note) IsDaily() bool {
	return n.Frontmatter.Type == "daily"
}

// IsTemplate returns true if this is a template note
func (n *Note) IsTemplate() bool {
	return n.Frontmatter.Type == "template"
}

// Validate performs basic validation on the note
func (n *Note) Validate() error {
	if n.ID == "" {
		return NewValidationError("note ID cannot be empty")
	}
	if n.Title == "" {
		return NewValidationError("note title cannot be empty")
	}
	if len(n.Content) > 10*1024*1024 { // 10MB limit
		return NewValidationError("note content exceeds 10MB limit")
	}
	return nil
}
