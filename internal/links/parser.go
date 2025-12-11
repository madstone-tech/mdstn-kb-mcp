package links

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

var (
	// WikiLinkRegex matches [[Note Title]] or [[note-id]]
	WikiLinkRegex = regexp.MustCompile(`\[\[([^\]]+)\]\]`)

	// MarkdownLinkRegex matches [Link Text](note-id) or [Link Text](path/to/note.md)
	MarkdownLinkRegex = regexp.MustCompile(`\[([^\]]+)\]\(([^\)]+)\)`)

	// HashtagRegex matches #tag
	HashtagRegex = regexp.MustCompile(`#([a-zA-Z0-9_-]+)`)
)

// Parser extracts links and relationships from note content
type Parser struct {
	// NoteResolver resolves note titles/IDs to actual notes
	NoteResolver NoteResolver
}

// NoteResolver interface for resolving note references
type NoteResolver interface {
	// ResolveByTitle finds a note by its title
	ResolveByTitle(title string) (*types.Note, error)

	// ResolveByID finds a note by its ID
	ResolveByID(id string) (*types.Note, error)

	// ResolveByPath finds a note by its file path
	ResolveByPath(path string) (*types.Note, error)
}

// New creates a new link parser
func New(resolver NoteResolver) *Parser {
	return &Parser{
		NoteResolver: resolver,
	}
}

// ParseLinks extracts all links from note content
func (p *Parser) ParseLinks(note *types.Note) ([]types.Link, error) {
	var links []types.Link

	// Parse wiki links
	wikiLinks, err := p.parseWikiLinks(note)
	if err != nil {
		return nil, fmt.Errorf("failed to parse wiki links: %w", err)
	}
	links = append(links, wikiLinks...)

	// Parse markdown links
	mdLinks, err := p.parseMarkdownLinks(note)
	if err != nil {
		return nil, fmt.Errorf("failed to parse markdown links: %w", err)
	}
	links = append(links, mdLinks...)

	return links, nil
}

// parseWikiLinks extracts [[Note Title]] style links
func (p *Parser) parseWikiLinks(note *types.Note) ([]types.Link, error) {
	var links []types.Link

	matches := WikiLinkRegex.FindAllStringSubmatch(note.Content, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		linkText := match[1]
		position := strings.Index(note.Content, match[0])

		// Try to resolve the link
		targetNote, err := p.resolveWikiLink(linkText)

		link := types.Link{
			SourceID: note.ID,
			LinkText: linkText,
			Position: position,
			Type:     types.LinkTypeWiki,
			IsValid:  err == nil,
		}

		if targetNote != nil {
			link.TargetID = targetNote.ID
			link.TargetTitle = targetNote.Title
		}

		links = append(links, link)
	}

	return links, nil
}

// parseMarkdownLinks extracts [Link Text](target) style links
func (p *Parser) parseMarkdownLinks(note *types.Note) ([]types.Link, error) {
	var links []types.Link

	matches := MarkdownLinkRegex.FindAllStringSubmatch(note.Content, -1)

	for _, match := range matches {
		if len(match) < 3 {
			continue
		}

		linkText := match[1]
		target := match[2]
		position := strings.Index(note.Content, match[0])

		// Skip external URLs
		if strings.HasPrefix(target, "http://") || strings.HasPrefix(target, "https://") {
			continue
		}

		// Try to resolve the link
		targetNote, err := p.resolveMarkdownLink(target)

		link := types.Link{
			SourceID: note.ID,
			LinkText: linkText,
			Position: position,
			Type:     types.LinkTypeMarkdown,
			IsValid:  err == nil,
		}

		if targetNote != nil {
			link.TargetID = targetNote.ID
			link.TargetTitle = targetNote.Title
		}

		links = append(links, link)
	}

	return links, nil
}

// resolveWikiLink resolves [[Note Title]] or [[note-id]]
func (p *Parser) resolveWikiLink(linkText string) (*types.Note, error) {
	// First try as title
	if note, err := p.NoteResolver.ResolveByTitle(linkText); err == nil {
		return note, nil
	}

	// Then try as ID
	if note, err := p.NoteResolver.ResolveByID(linkText); err == nil {
		return note, nil
	}

	return nil, fmt.Errorf("could not resolve wiki link: %s", linkText)
}

// resolveMarkdownLink resolves [Link](target) where target can be ID or path
func (p *Parser) resolveMarkdownLink(target string) (*types.Note, error) {
	// Try as ID first
	if note, err := p.NoteResolver.ResolveByID(target); err == nil {
		return note, nil
	}

	// Try as path
	if note, err := p.NoteResolver.ResolveByPath(target); err == nil {
		return note, nil
	}

	return nil, fmt.Errorf("could not resolve markdown link: %s", target)
}

// ParseHashtags extracts hashtags from note content
func (p *Parser) ParseHashtags(content string) []string {
	matches := HashtagRegex.FindAllStringSubmatch(content, -1)

	var tags []string
	seen := make(map[string]bool)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		tag := match[1]
		if !seen[tag] {
			tags = append(tags, tag)
			seen[tag] = true
		}
	}

	return tags
}

// FindBrokenLinks identifies links that don't resolve to valid notes
func (p *Parser) FindBrokenLinks(note *types.Note) ([]types.Link, error) {
	allLinks, err := p.ParseLinks(note)
	if err != nil {
		return nil, err
	}

	var brokenLinks []types.Link
	for _, link := range allLinks {
		if !link.IsValid {
			brokenLinks = append(brokenLinks, link)
		}
	}

	return brokenLinks, nil
}

// ValidateLinks checks all links in a note and returns validation results
func (p *Parser) ValidateLinks(note *types.Note) (*LinkValidation, error) {
	allLinks, err := p.ParseLinks(note)
	if err != nil {
		return nil, err
	}

	validation := &LinkValidation{
		NoteID:      note.ID,
		TotalLinks:  len(allLinks),
		ValidLinks:  0,
		BrokenLinks: make([]types.Link, 0),
	}

	for _, link := range allLinks {
		if link.IsValid {
			validation.ValidLinks++
		} else {
			validation.BrokenLinks = append(validation.BrokenLinks, link)
		}
	}

	return validation, nil
}

// LinkValidation contains results of link validation
type LinkValidation struct {
	NoteID      string       `json:"note_id"`
	TotalLinks  int          `json:"total_links"`
	ValidLinks  int          `json:"valid_links"`
	BrokenLinks []types.Link `json:"broken_links"`
}

// IsValid returns true if all links are valid
func (lv *LinkValidation) IsValid() bool {
	return len(lv.BrokenLinks) == 0
}

// BrokenCount returns the number of broken links
func (lv *LinkValidation) BrokenCount() int {
	return len(lv.BrokenLinks)
}

// ValidPercentage returns the percentage of valid links (0-100)
func (lv *LinkValidation) ValidPercentage() float64 {
	if lv.TotalLinks == 0 {
		return 100.0
	}
	return float64(lv.ValidLinks) / float64(lv.TotalLinks) * 100.0
}
