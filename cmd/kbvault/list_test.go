package main

import (
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestHasAnyTag(t *testing.T) {
	tests := []struct {
		name       string
		noteTags   []string
		filterTags []string
		want       bool
	}{
		{
			name:       "exact_match",
			noteTags:   []string{"work", "meeting", "urgent"},
			filterTags: []string{"work"},
			want:       true,
		},
		{
			name:       "case_insensitive_match",
			noteTags:   []string{"Work", "Meeting"},
			filterTags: []string{"work"},
			want:       true,
		},
		{
			name:       "multiple_filter_tags_one_match",
			noteTags:   []string{"personal", "todo"},
			filterTags: []string{"work", "todo", "urgent"},
			want:       true,
		},
		{
			name:       "no_match",
			noteTags:   []string{"personal", "diary"},
			filterTags: []string{"work", "meeting"},
			want:       false,
		},
		{
			name:       "empty_note_tags",
			noteTags:   []string{},
			filterTags: []string{"work"},
			want:       false,
		},
		{
			name:       "empty_filter_tags",
			noteTags:   []string{"work"},
			filterTags: []string{},
			want:       false,
		},
		{
			name:       "both_empty",
			noteTags:   []string{},
			filterTags: []string{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasAnyTag(tt.noteTags, tt.filterTags)
			if got != tt.want {
				t.Errorf("hasAnyTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilterNotesByTags(t *testing.T) {
	// Create test notes
	note1 := &types.Note{
		ID:    "note1",
		Title: "Work Meeting",
		Frontmatter: types.Frontmatter{
			Tags: []string{"work", "meeting"},
		},
	}
	note2 := &types.Note{
		ID:    "note2",
		Title: "Personal Todo",
		Frontmatter: types.Frontmatter{
			Tags: []string{"personal", "todo"},
		},
	}
	note3 := &types.Note{
		ID:    "note3",
		Title: "Urgent Work Task",
		Frontmatter: types.Frontmatter{
			Tags: []string{"work", "urgent"},
		},
	}
	note4 := &types.Note{
		ID:    "note4",
		Title: "No Tags Note",
		Frontmatter: types.Frontmatter{
			Tags: []string{},
		},
	}

	notes := []*types.Note{note1, note2, note3, note4}

	tests := []struct {
		name       string
		notes      []*types.Note
		filterTags []string
		wantCount  int
		wantIDs    []string
	}{
		{
			name:       "filter_by_work",
			notes:      notes,
			filterTags: []string{"work"},
			wantCount:  2,
			wantIDs:    []string{"note1", "note3"},
		},
		{
			name:       "filter_by_personal",
			notes:      notes,
			filterTags: []string{"personal"},
			wantCount:  1,
			wantIDs:    []string{"note2"},
		},
		{
			name:       "filter_by_multiple_tags",
			notes:      notes,
			filterTags: []string{"meeting", "todo"},
			wantCount:  2,
			wantIDs:    []string{"note1", "note2"},
		},
		{
			name:       "filter_by_nonexistent_tag",
			notes:      notes,
			filterTags: []string{"nonexistent"},
			wantCount:  0,
			wantIDs:    []string{},
		},
		{
			name:       "empty_filter",
			notes:      notes,
			filterTags: []string{},
			wantCount:  0,
			wantIDs:    []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := filterNotesByTags(tt.notes, tt.filterTags)

			if len(filtered) != tt.wantCount {
				t.Errorf("filterNotesByTags() returned %d notes, want %d", len(filtered), tt.wantCount)
				return
			}

			// Check that the correct notes were returned
			gotIDs := make(map[string]bool)
			for _, note := range filtered {
				gotIDs[note.ID] = true
			}

			for _, wantID := range tt.wantIDs {
				if !gotIDs[wantID] {
					t.Errorf("filterNotesByTags() missing expected note ID: %s", wantID)
				}
			}
		})
	}
}

func TestSortNotes(t *testing.T) {
	// Create test notes with different timestamps
	now := time.Now()
	yesterday := now.AddDate(0, 0, -1)
	lastWeek := now.AddDate(0, 0, -7)

	note1 := &types.Note{
		ID:        "note1",
		Title:     "Zebra Note",
		CreatedAt: lastWeek,
		UpdatedAt: yesterday,
	}
	note2 := &types.Note{
		ID:        "note2",
		Title:     "Alpha Note",
		CreatedAt: yesterday,
		UpdatedAt: now,
	}
	note3 := &types.Note{
		ID:        "note3",
		Title:     "Beta Note",
		CreatedAt: now,
		UpdatedAt: lastWeek,
	}

	tests := []struct {
		name      string
		notes     []*types.Note
		sortBy    string
		reverse   bool
		wantOrder []string // Expected order of note IDs
	}{
		{
			name:      "sort_by_title",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "title",
			reverse:   false,
			wantOrder: []string{"note2", "note3", "note1"}, // Alpha, Beta, Zebra
		},
		{
			name:      "sort_by_title_reverse",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "title",
			reverse:   true,
			wantOrder: []string{"note1", "note3", "note2"}, // Zebra, Beta, Alpha
		},
		{
			name:      "sort_by_created",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "created",
			reverse:   false,
			wantOrder: []string{"note1", "note2", "note3"}, // lastWeek, yesterday, now
		},
		{
			name:      "sort_by_created_reverse",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "created",
			reverse:   true,
			wantOrder: []string{"note3", "note2", "note1"}, // now, yesterday, lastWeek
		},
		{
			name:      "sort_by_updated",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "updated",
			reverse:   false,
			wantOrder: []string{"note3", "note1", "note2"}, // lastWeek, yesterday, now
		},
		{
			name:      "sort_by_updated_reverse",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "updated",
			reverse:   true,
			wantOrder: []string{"note2", "note1", "note3"}, // now, yesterday, lastWeek
		},
		{
			name:      "sort_by_default",
			notes:     []*types.Note{note1, note2, note3},
			sortBy:    "invalid",
			reverse:   false,
			wantOrder: []string{"note3", "note1", "note2"}, // defaults to updated
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Make a copy of the notes slice to avoid modifying the original
			notesCopy := make([]*types.Note, len(tt.notes))
			copy(notesCopy, tt.notes)

			sortNotes(notesCopy, tt.sortBy, tt.reverse)

			// Check the order
			for i, expectedID := range tt.wantOrder {
				if notesCopy[i].ID != expectedID {
					t.Errorf("sortNotes() position %d: got %s, want %s", i, notesCopy[i].ID, expectedID)
				}
			}
		})
	}
}

func TestFormatRelativeTime(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name string
		time time.Time
		want string
	}{
		{
			name: "just_now",
			time: now.Add(-30 * time.Second),
			want: "just now",
		},
		{
			name: "5_minutes_ago",
			time: now.Add(-5 * time.Minute),
			want: "5 minute(s) ago",
		},
		{
			name: "2_hours_ago",
			time: now.Add(-2 * time.Hour),
			want: "2 hour(s) ago",
		},
		{
			name: "3_days_ago",
			time: now.Add(-3 * 24 * time.Hour),
			want: "3 day(s) ago",
		},
		{
			name: "2_weeks_ago",
			time: now.Add(-14 * 24 * time.Hour),
			want: now.Add(-14 * 24 * time.Hour).Format("2006-01-02"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatRelativeTime(tt.time)
			if got != tt.want {
				t.Errorf("formatRelativeTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatTagsJSON(t *testing.T) {
	tests := []struct {
		name string
		tags []string
		want string
	}{
		{
			name: "multiple_tags",
			tags: []string{"work", "meeting", "urgent"},
			want: `"work", "meeting", "urgent"`,
		},
		{
			name: "single_tag",
			tags: []string{"personal"},
			want: `"personal"`,
		},
		{
			name: "empty_tags",
			tags: []string{},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatTagsJSON(tt.tags)
			if got != tt.want {
				t.Errorf("formatTagsJSON() = %v, want %v", got, tt.want)
			}
		})
	}
}
