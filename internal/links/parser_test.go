package links

import (
	"testing"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockNoteResolver implements NoteResolver for testing
type mockNoteResolver struct {
	notes   map[string]*types.Note
	byTitle map[string]*types.Note
	byPath  map[string]*types.Note
}

func newMockResolver() *mockNoteResolver {
	return &mockNoteResolver{
		notes:   make(map[string]*types.Note),
		byTitle: make(map[string]*types.Note),
		byPath:  make(map[string]*types.Note),
	}
}

func (m *mockNoteResolver) addNote(note *types.Note) {
	m.notes[note.ID] = note
	m.byTitle[note.Title] = note
	m.byPath[note.FilePath] = note
}

func (m *mockNoteResolver) ResolveByTitle(title string) (*types.Note, error) {
	if note, ok := m.byTitle[title]; ok {
		return note, nil
	}
	return nil, types.NewValidationError("note not found")
}

func (m *mockNoteResolver) ResolveByID(id string) (*types.Note, error) {
	if note, ok := m.notes[id]; ok {
		return note, nil
	}
	return nil, types.NewValidationError("note not found")
}

func (m *mockNoteResolver) ResolveByPath(path string) (*types.Note, error) {
	if note, ok := m.byPath[path]; ok {
		return note, nil
	}
	return nil, types.NewValidationError("note not found")
}

func TestParser_ParseWikiLinks(t *testing.T) {
	resolver := newMockResolver()

	// Add target notes
	targetNote1 := &types.Note{
		ID:       "note1",
		Title:    "Target Note 1",
		FilePath: "note1.md",
	}
	targetNote2 := &types.Note{
		ID:       "note2",
		Title:    "Target Note 2",
		FilePath: "note2.md",
	}

	resolver.addNote(targetNote1)
	resolver.addNote(targetNote2)

	parser := New(resolver)

	// Test note with wiki links
	sourceNote := &types.Note{
		ID:      "source",
		Title:   "Source Note",
		Content: "This links to [[Target Note 1]] and [[note2]] and [[Invalid Note]].",
	}

	links, err := parser.ParseLinks(sourceNote)
	require.NoError(t, err)

	// Should find 3 links total
	assert.Len(t, links, 3)

	// Check first link (by title)
	link1 := findLinkByText(links, "Target Note 1")
	require.NotNil(t, link1)
	assert.Equal(t, "source", link1.SourceID)
	assert.Equal(t, "note1", link1.TargetID)
	assert.Equal(t, "Target Note 1", link1.TargetTitle)
	assert.Equal(t, types.LinkTypeWiki, link1.Type)
	assert.True(t, link1.IsValid)

	// Check second link (by ID)
	link2 := findLinkByText(links, "note2")
	require.NotNil(t, link2)
	assert.Equal(t, "note2", link2.TargetID)
	assert.True(t, link2.IsValid)

	// Check third link (invalid)
	link3 := findLinkByText(links, "Invalid Note")
	require.NotNil(t, link3)
	assert.False(t, link3.IsValid)
	assert.Equal(t, "", link3.TargetID)
}

func TestParser_ParseMarkdownLinks(t *testing.T) {
	resolver := newMockResolver()

	// Add target note
	targetNote := &types.Note{
		ID:       "note1",
		Title:    "Target Note",
		FilePath: "notes/target.md",
	}
	resolver.addNote(targetNote)

	parser := New(resolver)

	// Test note with markdown links
	sourceNote := &types.Note{
		ID:      "source",
		Title:   "Source Note",
		Content: `This has [link by ID](note1) and [link by path](notes/target.md) and [external](https://example.com) and [invalid](nonexistent).`,
	}

	links, err := parser.ParseLinks(sourceNote)
	require.NoError(t, err)

	// Should find 2 valid internal links (external URL is skipped)
	internalLinks := filterInternalLinks(links)
	assert.Len(t, internalLinks, 3) // 2 valid + 1 invalid

	// Check link by ID
	linkByID := findLinkByText(internalLinks, "link by ID")
	require.NotNil(t, linkByID)
	assert.Equal(t, "note1", linkByID.TargetID)
	assert.Equal(t, types.LinkTypeMarkdown, linkByID.Type)
	assert.True(t, linkByID.IsValid)

	// Check link by path
	linkByPath := findLinkByText(internalLinks, "link by path")
	require.NotNil(t, linkByPath)
	assert.Equal(t, "note1", linkByPath.TargetID)
	assert.True(t, linkByPath.IsValid)

	// Check invalid link
	invalidLink := findLinkByText(internalLinks, "invalid")
	require.NotNil(t, invalidLink)
	assert.False(t, invalidLink.IsValid)
}

func TestParser_ParseHashtags(t *testing.T) {
	parser := New(nil) // No resolver needed for hashtag parsing

	content := "This has #golang and #web-dev tags, plus #AI_research and duplicate #golang."

	tags := parser.ParseHashtags(content)

	expectedTags := []string{"golang", "web-dev", "AI_research"}
	assert.ElementsMatch(t, expectedTags, tags)
}

func TestParser_FindBrokenLinks(t *testing.T) {
	resolver := newMockResolver()
	parser := New(resolver)

	// Note with only broken links
	note := &types.Note{
		ID:      "source",
		Title:   "Source Note",
		Content: "Links to [[Nonexistent Note]] and [another](invalid-id).",
	}

	brokenLinks, err := parser.FindBrokenLinks(note)
	require.NoError(t, err)

	assert.Len(t, brokenLinks, 2)
	for _, link := range brokenLinks {
		assert.False(t, link.IsValid)
	}
}

func TestParser_ValidateLinks(t *testing.T) {
	resolver := newMockResolver()

	// Add one valid target
	targetNote := &types.Note{
		ID:    "valid",
		Title: "Valid Target",
	}
	resolver.addNote(targetNote)

	parser := New(resolver)

	// Note with mix of valid and invalid links
	note := &types.Note{
		ID:      "source",
		Title:   "Source Note",
		Content: "Valid: [[Valid Target]], Invalid: [[Invalid Target]], Another invalid: [broken](missing).",
	}

	validation, err := parser.ValidateLinks(note)
	require.NoError(t, err)

	assert.Equal(t, "source", validation.NoteID)
	assert.Equal(t, 3, validation.TotalLinks)
	assert.Equal(t, 1, validation.ValidLinks)
	assert.Len(t, validation.BrokenLinks, 2)
	assert.False(t, validation.IsValid())
	assert.Equal(t, 2, validation.BrokenCount())
	assert.InDelta(t, 33.33, validation.ValidPercentage(), 0.01)
}

func TestWikiLinkRegex(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{
			input:    "Simple [[Note Title]] link",
			expected: []string{"Note Title"},
		},
		{
			input:    "Multiple [[First Note]] and [[Second Note]] links",
			expected: []string{"First Note", "Second Note"},
		},
		{
			input:    "With [[note-id-123]] ID style",
			expected: []string{"note-id-123"},
		},
		{
			input:    "No links here",
			expected: nil,
		},
		{
			input:    "Malformed [Note Title] link",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := WikiLinkRegex.FindAllStringSubmatch(tt.input, -1)

			var actual []string
			for _, match := range matches {
				if len(match) > 1 {
					actual = append(actual, match[1])
				}
			}

			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestMarkdownLinkRegex(t *testing.T) {
	tests := []struct {
		input          string
		expectedText   []string
		expectedTarget []string
	}{
		{
			input:          "Simple [Link Text](target) link",
			expectedText:   []string{"Link Text"},
			expectedTarget: []string{"target"},
		},
		{
			input:          "Multiple [First](target1) and [Second](target2) links",
			expectedText:   []string{"First", "Second"},
			expectedTarget: []string{"target1", "target2"},
		},
		{
			input:          "External [Google](https://google.com) link",
			expectedText:   []string{"Google"},
			expectedTarget: []string{"https://google.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			matches := MarkdownLinkRegex.FindAllStringSubmatch(tt.input, -1)

			var actualText, actualTarget []string
			for _, match := range matches {
				if len(match) > 2 {
					actualText = append(actualText, match[1])
					actualTarget = append(actualTarget, match[2])
				}
			}

			assert.Equal(t, tt.expectedText, actualText)
			assert.Equal(t, tt.expectedTarget, actualTarget)
		})
	}
}

// Helper functions

func findLinkByText(links []types.Link, text string) *types.Link {
	for _, link := range links {
		if link.LinkText == text {
			return &link
		}
	}
	return nil
}

func filterInternalLinks(links []types.Link) []types.Link {
	var internal []types.Link
	for _, link := range links {
		if link.Type == types.LinkTypeWiki || link.Type == types.LinkTypeMarkdown {
			internal = append(internal, link)
		}
	}
	return internal
}
