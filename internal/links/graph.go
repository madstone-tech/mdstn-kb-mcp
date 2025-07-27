package links

import (
	"context"
	"fmt"
	"sort"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

// Graph represents the link structure between notes
type Graph struct {
	// Forward links: source note ID -> target note IDs
	forward map[string]map[string]bool
	
	// Backward links: target note ID -> source note IDs
	backward map[string]map[string]bool
	
	// Link details: (source, target) -> Link
	links map[string]map[string]types.Link
	
	// Note metadata cache
	notes map[string]*types.NoteMetadata
}

// NewGraph creates a new link graph
func NewGraph() *Graph {
	return &Graph{
		forward:  make(map[string]map[string]bool),
		backward: make(map[string]map[string]bool),
		links:    make(map[string]map[string]types.Link),
		notes:    make(map[string]*types.NoteMetadata),
	}
}

// AddNote adds a note to the graph
func (g *Graph) AddNote(note *types.NoteMetadata) {
	g.notes[note.ID] = note
}

// AddLink adds a link to the graph
func (g *Graph) AddLink(link types.Link) {
	// Initialize maps if needed
	if g.forward[link.SourceID] == nil {
		g.forward[link.SourceID] = make(map[string]bool)
	}
	if g.backward[link.TargetID] == nil {
		g.backward[link.TargetID] = make(map[string]bool)
	}
	if g.links[link.SourceID] == nil {
		g.links[link.SourceID] = make(map[string]types.Link)
	}
	
	// Add the link
	g.forward[link.SourceID][link.TargetID] = true
	g.backward[link.TargetID][link.SourceID] = true
	g.links[link.SourceID][link.TargetID] = link
}

// RemoveLink removes a link from the graph
func (g *Graph) RemoveLink(sourceID, targetID string) {
	if g.forward[sourceID] != nil {
		delete(g.forward[sourceID], targetID)
		if len(g.forward[sourceID]) == 0 {
			delete(g.forward, sourceID)
		}
	}
	
	if g.backward[targetID] != nil {
		delete(g.backward[targetID], sourceID)
		if len(g.backward[targetID]) == 0 {
			delete(g.backward, targetID)
		}
	}
	
	if g.links[sourceID] != nil {
		delete(g.links[sourceID], targetID)
		if len(g.links[sourceID]) == 0 {
			delete(g.links, sourceID)
		}
	}
}

// GetOutgoingLinks returns all notes that the given note links to
func (g *Graph) GetOutgoingLinks(noteID string) []types.Link {
	var result []types.Link
	
	if targets, exists := g.forward[noteID]; exists {
		for targetID := range targets {
			if link, exists := g.links[noteID][targetID]; exists {
				result = append(result, link)
			}
		}
	}
	
	// Sort by position in source note
	sort.Slice(result, func(i, j int) bool {
		return result[i].Position < result[j].Position
	})
	
	return result
}

// GetIncomingLinks returns all notes that link to the given note (backlinks)
func (g *Graph) GetIncomingLinks(noteID string) []types.Link {
	var result []types.Link
	
	if sources, exists := g.backward[noteID]; exists {
		for sourceID := range sources {
			if link, exists := g.links[sourceID][noteID]; exists {
				result = append(result, link)
			}
		}
	}
	
	// Sort by source note title
	sort.Slice(result, func(i, j int) bool {
		sourceA := g.notes[result[i].SourceID]
		sourceB := g.notes[result[j].SourceID]
		if sourceA != nil && sourceB != nil {
			return sourceA.Title < sourceB.Title
		}
		return result[i].SourceID < result[j].SourceID
	})
	
	return result
}

// GetConnectedNotes returns all notes connected to the given note
func (g *Graph) GetConnectedNotes(noteID string) []string {
	connected := make(map[string]bool)
	
	// Add outgoing links
	if targets, exists := g.forward[noteID]; exists {
		for targetID := range targets {
			connected[targetID] = true
		}
	}
	
	// Add incoming links
	if sources, exists := g.backward[noteID]; exists {
		for sourceID := range sources {
			connected[sourceID] = true
		}
	}
	
	var result []string
	for id := range connected {
		result = append(result, id)
	}
	
	sort.Strings(result)
	return result
}

// GetOrphanNotes returns notes with no incoming or outgoing links
func (g *Graph) GetOrphanNotes() []string {
	var orphans []string
	
	for noteID := range g.notes {
		hasOutgoing := len(g.forward[noteID]) > 0
		hasIncoming := len(g.backward[noteID]) > 0
		
		if !hasOutgoing && !hasIncoming {
			orphans = append(orphans, noteID)
		}
	}
	
	sort.Strings(orphans)
	return orphans
}

// GetMostLinkedNotes returns notes with the most incoming links
func (g *Graph) GetMostLinkedNotes(limit int) []NoteRank {
	type noteCount struct {
		noteID string
		count  int
	}
	
	var counts []noteCount
	for noteID, sources := range g.backward {
		counts = append(counts, noteCount{
			noteID: noteID,
			count:  len(sources),
		})
	}
	
	// Sort by count descending
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})
	
	// Convert to NoteRank and apply limit
	var result []NoteRank
	for i, nc := range counts {
		if limit > 0 && i >= limit {
			break
		}
		
		note := g.notes[nc.noteID]
		rank := NoteRank{
			NoteID: nc.noteID,
			Score:  nc.count,
		}
		if note != nil {
			rank.Title = note.Title
		}
		
		result = append(result, rank)
	}
	
	return result
}

// GetMostConnectedNotes returns notes with the most total connections
func (g *Graph) GetMostConnectedNotes(limit int) []NoteRank {
	type noteCount struct {
		noteID string
		count  int
	}
	
	var counts []noteCount
	for noteID := range g.notes {
		outgoing := len(g.forward[noteID])
		incoming := len(g.backward[noteID])
		total := outgoing + incoming
		
		if total > 0 {
			counts = append(counts, noteCount{
				noteID: noteID,
				count:  total,
			})
		}
	}
	
	// Sort by count descending
	sort.Slice(counts, func(i, j int) bool {
		return counts[i].count > counts[j].count
	})
	
	// Convert to NoteRank and apply limit
	var result []NoteRank
	for i, nc := range counts {
		if limit > 0 && i >= limit {
			break
		}
		
		note := g.notes[nc.noteID]
		rank := NoteRank{
			NoteID: nc.noteID,
			Score:  nc.count,
		}
		if note != nil {
			rank.Title = note.Title
		}
		
		result = append(result, rank)
	}
	
	return result
}

// FindPath finds the shortest path between two notes
func (g *Graph) FindPath(sourceID, targetID string) []string {
	if sourceID == targetID {
		return []string{sourceID}
	}
	
	// BFS to find shortest path
	queue := [][]string{{sourceID}}
	visited := make(map[string]bool)
	visited[sourceID] = true
	
	for len(queue) > 0 {
		path := queue[0]
		queue = queue[1:]
		
		currentID := path[len(path)-1]
		
		// Check all outgoing links from current note
		if targets, exists := g.forward[currentID]; exists {
			for nextID := range targets {
				if nextID == targetID {
					// Found target
					return append(path, nextID)
				}
				
				if !visited[nextID] {
					visited[nextID] = true
					newPath := make([]string, len(path)+1)
					copy(newPath, path)
					newPath[len(path)] = nextID
					queue = append(queue, newPath)
				}
			}
		}
	}
	
	// No path found
	return nil
}

// GetStatistics returns graph statistics
func (g *Graph) GetStatistics() *GraphStatistics {
	stats := &GraphStatistics{
		TotalNotes:  len(g.notes),
		TotalLinks:  0,
		OrphanNotes: len(g.GetOrphanNotes()),
	}
	
	// Count total links
	for _, targets := range g.forward {
		stats.TotalLinks += len(targets)
	}
	
	// Calculate average connections per note
	if stats.TotalNotes > 0 {
		totalConnections := 0
		for noteID := range g.notes {
			outgoing := len(g.forward[noteID])
			incoming := len(g.backward[noteID])
			totalConnections += outgoing + incoming
		}
		stats.AvgConnectionsPerNote = float64(totalConnections) / float64(stats.TotalNotes)
	}
	
	return stats
}

// NoteRank represents a note with a ranking score
type NoteRank struct {
	NoteID string `json:"note_id"`
	Title  string `json:"title"`
	Score  int    `json:"score"`
}

// GraphStatistics contains graph analysis results
type GraphStatistics struct {
	TotalNotes             int     `json:"total_notes"`
	TotalLinks             int     `json:"total_links"`
	OrphanNotes            int     `json:"orphan_notes"`
	AvgConnectionsPerNote  float64 `json:"avg_connections_per_note"`
}

// Builder helps construct a graph from a collection of notes
type Builder struct {
	parser   *Parser
	graph    *Graph
}

// NewBuilder creates a new graph builder
func NewBuilder(parser *Parser) *Builder {
	return &Builder{
		parser: parser,
		graph:  NewGraph(),
	}
}

// BuildFromNotes constructs a graph from a slice of notes
func (b *Builder) BuildFromNotes(ctx context.Context, notes []*types.Note) (*Graph, error) {
	// First pass: add all notes to graph
	for _, note := range notes {
		metadata := note.ToMetadata()
		b.graph.AddNote(&metadata)
	}
	
	// Second pass: parse and add links
	for _, note := range notes {
		links, err := b.parser.ParseLinks(note)
		if err != nil {
			return nil, fmt.Errorf("failed to parse links for note %s: %w", note.ID, err)
		}
		
		for _, link := range links {
			b.graph.AddLink(link)
		}
	}
	
	return b.graph, nil
}

// GetGraph returns the constructed graph
func (b *Builder) GetGraph() *Graph {
	return b.graph
}