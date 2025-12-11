package links

import (
	"context"
	"testing"
	"time"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestGraph_AddAndRemoveLinks(t *testing.T) {
	graph := NewGraph()

	// Add notes
	note1 := &types.NoteMetadata{
		ID:    "note1",
		Title: "First Note",
	}
	note2 := &types.NoteMetadata{
		ID:    "note2",
		Title: "Second Note",
	}

	graph.AddNote(note1)
	graph.AddNote(note2)

	// Add link
	link := types.Link{
		SourceID: "note1",
		TargetID: "note2",
		LinkText: "Second Note",
		Type:     types.LinkTypeWiki,
		IsValid:  true,
	}

	graph.AddLink(link)

	// Test outgoing links
	outgoing := graph.GetOutgoingLinks("note1")
	assert.Len(t, outgoing, 1)
	assert.Equal(t, "note2", outgoing[0].TargetID)

	// Test incoming links
	incoming := graph.GetIncomingLinks("note2")
	assert.Len(t, incoming, 1)
	assert.Equal(t, "note1", incoming[0].SourceID)

	// Test connected notes
	connected := graph.GetConnectedNotes("note1")
	assert.Contains(t, connected, "note2")

	connected = graph.GetConnectedNotes("note2")
	assert.Contains(t, connected, "note1")

	// Remove link
	graph.RemoveLink("note1", "note2")

	// Verify removal
	outgoing = graph.GetOutgoingLinks("note1")
	assert.Len(t, outgoing, 0)

	incoming = graph.GetIncomingLinks("note2")
	assert.Len(t, incoming, 0)
}

func TestGraph_FindPath(t *testing.T) {
	graph := NewGraph()

	// Create a chain: note1 -> note2 -> note3
	notes := []*types.NoteMetadata{
		{ID: "note1", Title: "Note 1"},
		{ID: "note2", Title: "Note 2"},
		{ID: "note3", Title: "Note 3"},
		{ID: "note4", Title: "Note 4"}, // isolated
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Add links to create chain
	links := []types.Link{
		{SourceID: "note1", TargetID: "note2", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "note2", TargetID: "note3", Type: types.LinkTypeWiki, IsValid: true},
	}

	for _, link := range links {
		graph.AddLink(link)
	}

	// Test direct path
	path := graph.FindPath("note1", "note2")
	assert.Equal(t, []string{"note1", "note2"}, path)

	// Test longer path
	path = graph.FindPath("note1", "note3")
	assert.Equal(t, []string{"note1", "note2", "note3"}, path)

	// Test path to self
	path = graph.FindPath("note1", "note1")
	assert.Equal(t, []string{"note1"}, path)

	// Test no path
	path = graph.FindPath("note1", "note4")
	assert.Nil(t, path)
}

func TestGraph_GetOrphanNotes(t *testing.T) {
	graph := NewGraph()

	// Add notes
	notes := []*types.NoteMetadata{
		{ID: "connected1", Title: "Connected 1"},
		{ID: "connected2", Title: "Connected 2"},
		{ID: "orphan1", Title: "Orphan 1"},
		{ID: "orphan2", Title: "Orphan 2"},
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Connect only some notes
	link := types.Link{
		SourceID: "connected1",
		TargetID: "connected2",
		Type:     types.LinkTypeWiki,
		IsValid:  true,
	}
	graph.AddLink(link)

	// Get orphans
	orphans := graph.GetOrphanNotes()
	assert.ElementsMatch(t, []string{"orphan1", "orphan2"}, orphans)
}

func TestGraph_GetMostLinkedNotes(t *testing.T) {
	graph := NewGraph()

	// Add notes
	notes := []*types.NoteMetadata{
		{ID: "popular", Title: "Popular Note"},
		{ID: "source1", Title: "Source 1"},
		{ID: "source2", Title: "Source 2"},
		{ID: "source3", Title: "Source 3"},
		{ID: "unpopular", Title: "Unpopular Note"},
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Multiple links to popular note
	links := []types.Link{
		{SourceID: "source1", TargetID: "popular", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "source2", TargetID: "popular", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "source3", TargetID: "popular", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "source1", TargetID: "unpopular", Type: types.LinkTypeWiki, IsValid: true},
	}

	for _, link := range links {
		graph.AddLink(link)
	}

	// Get most linked notes
	ranked := graph.GetMostLinkedNotes(2)
	assert.Len(t, ranked, 2)

	// Popular note should be first
	assert.Equal(t, "popular", ranked[0].NoteID)
	assert.Equal(t, "Popular Note", ranked[0].Title)
	assert.Equal(t, 3, ranked[0].Score)

	// Unpopular note should be second
	assert.Equal(t, "unpopular", ranked[1].NoteID)
	assert.Equal(t, 1, ranked[1].Score)
}

func TestGraph_GetMostConnectedNotes(t *testing.T) {
	graph := NewGraph()

	// Add notes
	notes := []*types.NoteMetadata{
		{ID: "hub", Title: "Hub Note"},
		{ID: "connected1", Title: "Connected 1"},
		{ID: "connected2", Title: "Connected 2"},
		{ID: "isolated", Title: "Isolated"},
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Hub has both incoming and outgoing connections
	links := []types.Link{
		{SourceID: "connected1", TargetID: "hub", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "connected2", TargetID: "hub", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "hub", TargetID: "connected1", Type: types.LinkTypeWiki, IsValid: true},
	}

	for _, link := range links {
		graph.AddLink(link)
	}

	// Get most connected notes
	ranked := graph.GetMostConnectedNotes(0) // no limit

	// Hub should be most connected (3 total connections)
	assert.True(t, len(ranked) > 0)
	assert.Equal(t, "hub", ranked[0].NoteID)
	assert.Equal(t, 3, ranked[0].Score)
}

func TestGraph_GetStatistics(t *testing.T) {
	graph := NewGraph()

	// Add notes
	notes := []*types.NoteMetadata{
		{ID: "note1", Title: "Note 1"},
		{ID: "note2", Title: "Note 2"},
		{ID: "note3", Title: "Note 3"},
		{ID: "orphan", Title: "Orphan"},
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Add some links
	links := []types.Link{
		{SourceID: "note1", TargetID: "note2", Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "note2", TargetID: "note3", Type: types.LinkTypeWiki, IsValid: true},
	}

	for _, link := range links {
		graph.AddLink(link)
	}

	stats := graph.GetStatistics()

	assert.Equal(t, 4, stats.TotalNotes)
	assert.Equal(t, 2, stats.TotalLinks)
	assert.Equal(t, 1, stats.OrphanNotes) // orphan note
	assert.Greater(t, stats.AvgConnectionsPerNote, 0.0)
}

func TestGraph_SortingOrder(t *testing.T) {
	graph := NewGraph()

	// Add notes with specific titles for sorting test
	notes := []*types.NoteMetadata{
		{ID: "z", Title: "Z Note"},
		{ID: "a", Title: "A Note"},
		{ID: "m", Title: "M Note"},
	}

	for _, note := range notes {
		graph.AddNote(note)
	}

	// Create links from z to others
	links := []types.Link{
		{SourceID: "z", TargetID: "a", Position: 100, Type: types.LinkTypeWiki, IsValid: true},
		{SourceID: "z", TargetID: "m", Position: 50, Type: types.LinkTypeWiki, IsValid: true},
	}

	for _, link := range links {
		graph.AddLink(link)
	}

	// Test outgoing links are sorted by position
	outgoing := graph.GetOutgoingLinks("z")
	assert.Len(t, outgoing, 2)
	assert.Equal(t, "m", outgoing[0].TargetID) // Position 50 comes first
	assert.Equal(t, "a", outgoing[1].TargetID) // Position 100 comes second

	// Test incoming links are sorted by source title
	graph.AddLink(types.Link{SourceID: "m", TargetID: "z", Type: types.LinkTypeWiki, IsValid: true})
	graph.AddLink(types.Link{SourceID: "a", TargetID: "z", Type: types.LinkTypeWiki, IsValid: true})

	incoming := graph.GetIncomingLinks("z")
	assert.Len(t, incoming, 2)
	assert.Equal(t, "a", incoming[0].SourceID) // "A Note" comes before "M Note"
	assert.Equal(t, "m", incoming[1].SourceID)
}

func TestBuilder_BuildFromNotes(t *testing.T) {
	resolver := newMockResolver()

	// Create test notes
	note1 := &types.Note{
		ID:        "note1",
		Title:     "First Note",
		Content:   "Links to [[Second Note]] and [[note3]].",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	note2 := &types.Note{
		ID:        "note2",
		Title:     "Second Note",
		Content:   "Links back to [[First Note]].",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	note3 := &types.Note{
		ID:        "note3",
		Title:     "Third Note",
		Content:   "No links here.",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add to resolver
	resolver.addNote(note1)
	resolver.addNote(note2)
	resolver.addNote(note3)

	parser := New(resolver)
	builder := NewBuilder(parser)

	notes := []*types.Note{note1, note2, note3}
	graph, err := builder.BuildFromNotes(context.TODO(), notes)

	assert.NoError(t, err)
	assert.NotNil(t, graph)

	// Verify notes were added
	stats := graph.GetStatistics()
	assert.Equal(t, 3, stats.TotalNotes)

	// Verify links were created
	outgoing1 := graph.GetOutgoingLinks("note1")
	assert.Len(t, outgoing1, 2)

	outgoing2 := graph.GetOutgoingLinks("note2")
	assert.Len(t, outgoing2, 1)

	// Verify bidirectional relationship
	incoming1 := graph.GetIncomingLinks("note1")
	assert.Len(t, incoming1, 1)
	assert.Equal(t, "note2", incoming1[0].SourceID)
}
