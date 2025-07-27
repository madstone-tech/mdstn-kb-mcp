package ulid

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/oklog/ulid/v2"
)

var (
	// ULID regex pattern: 26 characters, Crockford's Base32 alphabet
	// 01ARZ3NDEKTSV4RRFFQ69G5FAV
	ulidPattern = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}$`)

	// Filename pattern: ULID + .md extension
	filenamePattern = regexp.MustCompile(`^[0-7][0-9A-HJKMNP-TV-Z]{25}\.md$`)
)

// ValidationResult contains the result of ULID validation
type ValidationResult struct {
	IsValid  bool                   `json:"is_valid"`
	Errors   []string               `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Validator provides comprehensive ULID validation
type Validator struct {
	strictMode bool
}

// NewValidator creates a new ULID validator
func NewValidator(strictMode bool) *Validator {
	return &Validator{
		strictMode: strictMode,
	}
}

// ValidateString performs comprehensive validation on a ULID string
func (v *Validator) ValidateString(id string) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Metadata: make(map[string]interface{}),
	}

	// Basic checks
	if id == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "ULID cannot be empty")
		return result
	}

	if len(id) != 26 {
		result.IsValid = false
		result.Errors = append(result.Errors, "ULID must be exactly 26 characters long")
		return result
	}

	// Pattern check
	if !ulidPattern.MatchString(id) {
		result.IsValid = false
		result.Errors = append(result.Errors, "ULID contains invalid characters")
		return result
	}

	// Parse check
	parsed, err := ulid.Parse(id)
	if err != nil {
		result.IsValid = false
		result.Errors = append(result.Errors, "invalid ULID format: "+err.Error())
		return result
	}

	// Extract metadata
	timestamp, _ := ExtractTimestamp(id)
	result.Metadata["timestamp"] = timestamp
	result.Metadata["timestamp_unix"] = timestamp.Unix()
	result.Metadata["age_hours"] = int(timestamp.Sub(timestamp).Hours())

	// Strict mode checks
	if v.strictMode {
		v.performStrictChecks(id, parsed, result)
	}

	return result
}

// ValidateFilename validates a filename containing a ULID
func (v *Validator) ValidateFilename(filename string) *ValidationResult {
	result := &ValidationResult{
		IsValid:  true,
		Metadata: make(map[string]interface{}),
	}

	if filename == "" {
		result.IsValid = false
		result.Errors = append(result.Errors, "filename cannot be empty")
		return result
	}

	// Check extension
	if !strings.HasSuffix(filename, ".md") {
		if v.strictMode {
			result.IsValid = false
			result.Errors = append(result.Errors, "filename must have .md extension")
			return result
		} else {
			result.Warnings = append(result.Warnings, "filename should have .md extension")
		}
	}

	// Pattern check
	if strings.HasSuffix(filename, ".md") && !filenamePattern.MatchString(filename) {
		result.IsValid = false
		result.Errors = append(result.Errors, "filename contains invalid ULID")
		return result
	}

	// Extract and validate ULID
	id := strings.TrimSuffix(filename, ".md")
	ulidResult := v.ValidateString(id)

	if !ulidResult.IsValid {
		result.IsValid = false
		result.Errors = append(result.Errors, ulidResult.Errors...)
	}

	result.Warnings = append(result.Warnings, ulidResult.Warnings...)

	// Copy metadata
	for k, v := range ulidResult.Metadata {
		result.Metadata[k] = v
	}

	result.Metadata["filename"] = filename
	result.Metadata["extracted_ulid"] = id

	return result
}

// ValidateBatch validates multiple ULIDs and returns summary statistics
func (v *Validator) ValidateBatch(ids []string) *BatchValidationResult {
	result := &BatchValidationResult{
		Total:   len(ids),
		Valid:   0,
		Invalid: 0,
		Results: make(map[string]*ValidationResult),
	}

	for _, id := range ids {
		validation := v.ValidateString(id)
		result.Results[id] = validation

		if validation.IsValid {
			result.Valid++
		} else {
			result.Invalid++
		}
	}

	return result
}

// BatchValidationResult contains results for batch validation
type BatchValidationResult struct {
	Total   int                          `json:"total"`
	Valid   int                          `json:"valid"`
	Invalid int                          `json:"invalid"`
	Results map[string]*ValidationResult `json:"results"`
}

// GetInvalidULIDs returns a list of invalid ULIDs from batch validation
func (r *BatchValidationResult) GetInvalidULIDs() []string {
	var invalid []string
	for id, result := range r.Results {
		if !result.IsValid {
			invalid = append(invalid, id)
		}
	}
	return invalid
}

// GetValidULIDs returns a list of valid ULIDs from batch validation
func (r *BatchValidationResult) GetValidULIDs() []string {
	var valid []string
	for id, result := range r.Results {
		if result.IsValid {
			valid = append(valid, id)
		}
	}
	return valid
}

// performStrictChecks performs additional validation in strict mode
func (v *Validator) performStrictChecks(id string, parsed ulid.ULID, result *ValidationResult) {
	timestamp := time.Unix(0, int64(parsed.Time())*int64(time.Millisecond))
	now := time.Now()

	// Check if timestamp is in the future
	if timestamp.After(now.Add(time.Minute)) { // Allow 1 minute tolerance
		result.Warnings = append(result.Warnings, "ULID timestamp is in the future")
	}

	// Check if timestamp is too old (more than 10 years)
	if timestamp.Before(now.Add(-10 * 365 * 24 * time.Hour)) {
		result.Warnings = append(result.Warnings, "ULID timestamp is very old (>10 years)")
	}

	// Check for potential collision (same millisecond)
	// This is more of an informational warning
	if timestamp.UnixMilli() == now.UnixMilli() {
		result.Warnings = append(result.Warnings, "ULID generated in current millisecond (potential collision)")
	}
}

// FixCommonIssues attempts to fix common ULID formatting issues
func FixCommonIssues(input string) (string, []string) {
	var fixes []string
	fixed := input

	// Remove common prefixes/suffixes
	if strings.HasPrefix(fixed, "ulid:") {
		fixed = strings.TrimPrefix(fixed, "ulid:")
		fixes = append(fixes, "removed 'ulid:' prefix")
	}

	// Trim whitespace
	trimmed := strings.TrimSpace(fixed)
	if trimmed != fixed {
		fixed = trimmed
		fixes = append(fixes, "trimmed whitespace")
	}

	// Convert to uppercase (ULIDs should be uppercase)
	upper := strings.ToUpper(fixed)
	if upper != fixed {
		fixed = upper
		fixes = append(fixes, "converted to uppercase")
	}

	// Remove .md extension if present during validation
	if strings.HasSuffix(fixed, ".md") {
		withoutExt := strings.TrimSuffix(fixed, ".md")
		if IsValid(withoutExt) {
			// Don't modify the fixed value, just note it
			fixes = append(fixes, "note: .md extension detected (valid for filename)")
		}
	}

	return fixed, fixes
}

// SuggestCorrections suggests possible corrections for invalid ULIDs
func SuggestCorrections(input string) []string {
	var suggestions []string

	// Try fixing common issues first
	fixed, fixes := FixCommonIssues(input)
	if len(fixes) > 0 && IsValid(fixed) {
		suggestions = append(suggestions, "Try: "+fixed+" ("+strings.Join(fixes, ", ")+")")
	}

	// Check length and suggest padding/truncation
	if len(fixed) < 26 {
		suggestions = append(suggestions, "ULID is too short (need "+fmt.Sprintf("%d", 26-len(fixed))+" more characters)")
	} else if len(fixed) > 26 {
		suggestions = append(suggestions, "ULID is too long (remove "+fmt.Sprintf("%d", len(fixed)-26)+" characters)")
	}

	// Check for invalid characters
	invalidChars := findInvalidCharacters(fixed)
	if len(invalidChars) > 0 {
		suggestions = append(suggestions, "Contains invalid characters: "+strings.Join(invalidChars, ", "))
		suggestions = append(suggestions, "Valid characters: 0-9, A-Z (excluding I, L, O, U)")
	}

	// Suggest generating a new ULID if all else fails
	if !IsValid(fixed) {
		suggestions = append(suggestions, "Consider generating a new ULID: "+New())
	}

	return suggestions
}

// findInvalidCharacters returns characters that are not valid in a ULID
func findInvalidCharacters(input string) []string {
	validChars := "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	var invalid []string
	seen := make(map[rune]bool)

	for _, char := range input {
		if !strings.ContainsRune(validChars, char) && !seen[char] {
			invalid = append(invalid, string(char))
			seen[char] = true
		}
	}

	return invalid
}

// Global validator instances
var (
	standardValidator = NewValidator(false)
	strictValidator   = NewValidator(true)
)

// Quick validation functions using global validators

// QuickValidate performs standard validation on a ULID string
func QuickValidate(id string) bool {
	return standardValidator.ValidateString(id).IsValid
}

// QuickValidateStrict performs strict validation on a ULID string
func QuickValidateStrict(id string) bool {
	return strictValidator.ValidateString(id).IsValid
}

// QuickValidateFilename performs standard validation on a filename
func QuickValidateFilename(filename string) bool {
	return standardValidator.ValidateFilename(filename).IsValid
}

// QuickValidateFilenameStrict performs strict validation on a filename
func QuickValidateFilenameStrict(filename string) bool {
	return strictValidator.ValidateFilename(filename).IsValid
}

// DetailedValidate returns full validation details for a ULID
func DetailedValidate(id string) *ValidationResult {
	return standardValidator.ValidateString(id)
}

// DetailedValidateStrict returns full strict validation details for a ULID
func DetailedValidateStrict(id string) *ValidationResult {
	return strictValidator.ValidateString(id)
}
