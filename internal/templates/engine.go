package templates

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// Engine handles template rendering and management
type Engine struct {
	templateDir string
	templates   map[string]*template.Template
}

// NewEngine creates a new template engine
func NewEngine(templateDir string) *Engine {
	return &Engine{
		templateDir: templateDir,
		templates:   make(map[string]*template.Template),
	}
}

// TemplateData contains data available to templates
type TemplateData struct {
	ID        string
	Title     string
	Tags      []string
	Type      string
	Created   time.Time
	Updated   time.Time
	Now       time.Time
	Date      string
	Time      string
	VaultName string
	Author    string
	Custom    map[string]interface{}
}

// Render renders a template with the given data
func (e *Engine) Render(templateName string, data TemplateData) (string, error) {
	tmpl, err := e.getTemplate(templateName)
	if err != nil {
		return "", fmt.Errorf("failed to get template %s: %w", templateName, err)
	}

	var result strings.Builder
	if err := tmpl.Execute(&result, data); err != nil {
		return "", fmt.Errorf("failed to execute template %s: %w", templateName, err)
	}

	return result.String(), nil
}

// getTemplate loads and parses a template
func (e *Engine) getTemplate(name string) (*template.Template, error) {
	// Check cache first
	if tmpl, exists := e.templates[name]; exists {
		return tmpl, nil
	}

	// Load template from file
	templatePath := filepath.Join(e.templateDir, name+".md")
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read template file: %w", err)
	}

	// Parse template with custom functions
	tmpl, err := template.New(name).Funcs(templateFuncs()).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	// Cache the template
	e.templates[name] = tmpl
	return tmpl, nil
}

// CreateDefaultTemplates creates default templates in the template directory
func (e *Engine) CreateDefaultTemplates() error {
	if err := os.MkdirAll(e.templateDir, 0755); err != nil {
		return fmt.Errorf("failed to create template directory: %w", err)
	}

	templates := map[string]string{
		"default": defaultTemplate,
		"daily":   dailyTemplate,
		"meeting": meetingTemplate,
		"book":    bookTemplate,
	}

	for name, content := range templates {
		templatePath := filepath.Join(e.templateDir, name+".md")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to create template %s: %w", name, err)
			}
		}
	}

	return nil
}

// ListTemplates returns all available template names
func (e *Engine) ListTemplates() ([]string, error) {
	files, err := os.ReadDir(e.templateDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read template directory: %w", err)
	}

	var templates []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
			name := strings.TrimSuffix(file.Name(), ".md")
			templates = append(templates, name)
		}
	}

	return templates, nil
}

// ValidateTemplate validates a template for syntax errors
func (e *Engine) ValidateTemplate(name string) error {
	_, err := e.getTemplate(name)
	return err
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"lower": strings.ToLower,
		"upper": strings.ToUpper,
		"title": func(s string) string {
			if len(s) == 0 {
				return s
			}
			// Simple title case - capitalize first letter of each word
			words := strings.Fields(s)
			for i, word := range words {
				if len(word) > 0 {
					words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
				}
			}
			return strings.Join(words, " ")
		},
		"join": strings.Join,
		"now":  time.Now,
		"date": func(format string) string {
			return time.Now().Format(format)
		},
		"add": func(a, b int) int {
			return a + b
		},
		"seq": func(start, end int) []int {
			var result []int
			for i := start; i <= end; i++ {
				result = append(result, i)
			}
			return result
		},
	}
}

// Default template definitions
const defaultTemplate = `# {{.Title}}

Created: {{.Created.Format "2006-01-02 15:04:05"}}
{{if .Tags}}Tags: {{join .Tags ", "}}{{end}}

## Content

Add your content here...

## Notes

- 
- 
- 

## Links

## References
`

const dailyTemplate = `# Daily Note - {{.Date}}

## 📅 {{.Date}} - {{.Time}}

### Today's Goals
- [ ] 
- [ ] 
- [ ] 

### Notes & Observations

### Meetings & Calls

### Tasks Completed
- [ ] 
- [ ] 

### Tomorrow's Priorities
- [ ] 
- [ ] 

### Reflection

---
*Created: {{.Created.Format "2006-01-02 15:04:05"}}*
`

const meetingTemplate = `# Meeting: {{.Title}}

**Date:** {{.Date}}  
**Time:** {{.Time}}  
{{if .Tags}}**Tags:** {{join .Tags ", "}}{{end}}

## Attendees
- 
- 
- 

## Agenda
1. 
2. 
3. 

## Discussion Points

### Topic 1


### Topic 2


### Topic 3


## Action Items
- [ ] **Person:** Task description (Due: YYYY-MM-DD)
- [ ] **Person:** Task description (Due: YYYY-MM-DD)

## Decisions Made
- 
- 

## Next Steps
- [ ] 
- [ ] 

## Notes & References

---
*Meeting notes created: {{.Created.Format "2006-01-02 15:04:05"}}*
`

const bookTemplate = `# Book: {{.Title}}

{{if .Tags}}**Tags:** {{join .Tags ", "}}{{end}}

## Book Information
- **Author:** 
- **Genre:** 
- **Pages:** 
- **Started:** {{.Date}}
- **Finished:** 
- **Rating:** ⭐⭐⭐⭐⭐ ( /5)

## Summary


## Key Insights
1. 
2. 
3. 

## Favorite Quotes
> "Quote here" - Page X

> "Another quote" - Page Y

## Notes by Chapter

### Chapter 1


### Chapter 2


## Action Items
- [ ] 
- [ ] 

## Related Books/Resources
- 
- 

---
*Book notes created: {{.Created.Format "2006-01-02 15:04:05"}}*
`
