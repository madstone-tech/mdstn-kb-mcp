package ulid

import (
	"crypto/rand"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/oklog/ulid/v2"
)

// Generator provides thread-safe ULID generation
type Generator struct {
	entropy *ulid.MonotonicEntropy
	mu      sync.Mutex
}

// NewGenerator creates a new ULID generator with cryptographic entropy
func NewGenerator() *Generator {
	entropy := ulid.Monotonic(rand.Reader, 0)
	return &Generator{
		entropy: entropy,
	}
}

// Generate creates a new ULID
func (g *Generator) Generate() string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(time.Now()), g.entropy)
	return id.String()
}

// GenerateAt creates a ULID with a specific timestamp
func (g *Generator) GenerateAt(t time.Time) string {
	g.mu.Lock()
	defer g.mu.Unlock()

	id := ulid.MustNew(ulid.Timestamp(t), g.entropy)
	return id.String()
}

// GenerateWithPrefix creates a ULID and formats it for file naming
// Returns: "{ulid}.md"
func (g *Generator) GenerateWithPrefix() string {
	id := g.Generate()
	return id + ".md"
}

// Parse parses a ULID string and returns the underlying ULID
func Parse(s string) (ulid.ULID, error) {
	return ulid.Parse(s)
}

// ParseFromFilename extracts ULID from a filename
// Handles both "01ARZ3NDEKTSV4RRFFQ69G5FAV.md" and "01ARZ3NDEKTSV4RRFFQ69G5FAV"
func ParseFromFilename(filename string) (ulid.ULID, error) {
	// Remove .md extension if present
	id := strings.TrimSuffix(filename, ".md")
	return ulid.Parse(id)
}

// ExtractTimestamp extracts the timestamp from a ULID
func ExtractTimestamp(id string) (time.Time, error) {
	parsed, err := ulid.Parse(id)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(parsed.Time())*int64(time.Millisecond)), nil
}

// ExtractTimestampFromFilename extracts timestamp from a filename containing ULID
func ExtractTimestampFromFilename(filename string) (time.Time, error) {
	parsed, err := ParseFromFilename(filename)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(parsed.Time())*int64(time.Millisecond)), nil
}

// IsValid checks if a string is a valid ULID
func IsValid(s string) bool {
	_, err := ulid.Parse(s)
	return err == nil
}

// IsValidFilename checks if a filename contains a valid ULID
func IsValidFilename(filename string) bool {
	id := strings.TrimSuffix(filename, ".md")
	return IsValid(id)
}

// Compare compares two ULIDs lexicographically
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func Compare(a, b string) int {
	return strings.Compare(a, b)
}

// CompareTimestamps compares the timestamps of two ULIDs
// Returns -1 if a is older, 0 if same time, 1 if a is newer
func CompareTimestamps(a, b string) (int, error) {
	timeA, err := ExtractTimestamp(a)
	if err != nil {
		return 0, err
	}

	timeB, err := ExtractTimestamp(b)
	if err != nil {
		return 0, err
	}

	if timeA.Before(timeB) {
		return -1, nil
	} else if timeA.After(timeB) {
		return 1, nil
	}
	return 0, nil
}

// SortByTimestamp sorts ULID strings by their embedded timestamps
func SortByTimestamp(ids []string, ascending bool) []string {
	sorted := make([]string, len(ids))
	copy(sorted, ids)

	if ascending {
		// ULIDs are naturally sortable in ascending order
		return sorted
	}

	// For descending, reverse the slice
	for i := len(sorted)/2 - 1; i >= 0; i-- {
		opp := len(sorted) - 1 - i
		sorted[i], sorted[opp] = sorted[opp], sorted[i]
	}

	return sorted
}

// TimestampRange represents a time range for filtering ULIDs
type TimestampRange struct {
	Start time.Time
	End   time.Time
}

// FilterByTimeRange filters ULIDs to only include those within the time range
func FilterByTimeRange(ids []string, timeRange TimestampRange) ([]string, error) {
	var filtered []string

	for _, id := range ids {
		timestamp, err := ExtractTimestamp(id)
		if err != nil {
			continue // Skip invalid ULIDs
		}

		if (timeRange.Start.IsZero() || timestamp.After(timeRange.Start) || timestamp.Equal(timeRange.Start)) &&
			(timeRange.End.IsZero() || timestamp.Before(timeRange.End) || timestamp.Equal(timeRange.End)) {
			filtered = append(filtered, id)
		}
	}

	return filtered, nil
}

// GetTimeFromULID extracts time.Time from ULID string with error handling
func GetTimeFromULID(id string) (time.Time, error) {
	return ExtractTimestamp(id)
}

// GetDayFromULID extracts the date (YYYY-MM-DD) from a ULID
func GetDayFromULID(id string) (string, error) {
	timestamp, err := ExtractTimestamp(id)
	if err != nil {
		return "", err
	}
	return timestamp.Format("2006-01-02"), nil
}

// IsFromToday checks if a ULID was created today
func IsFromToday(id string) (bool, error) {
	timestamp, err := ExtractTimestamp(id)
	if err != nil {
		return false, err
	}

	now := time.Now()
	return timestamp.Year() == now.Year() &&
		timestamp.YearDay() == now.YearDay(), nil
}

// IsFromDate checks if a ULID was created on a specific date
func IsFromDate(id string, date time.Time) (bool, error) {
	timestamp, err := ExtractTimestamp(id)
	if err != nil {
		return false, err
	}

	return timestamp.Year() == date.Year() &&
		timestamp.YearDay() == date.YearDay(), nil
}

// AgeInDays returns how many days old a ULID is
func AgeInDays(id string) (int, error) {
	timestamp, err := ExtractTimestamp(id)
	if err != nil {
		return 0, err
	}

	duration := time.Since(timestamp)
	return int(math.Floor(duration.Hours() / 24)), nil
}

// ToFilename converts a ULID to a filename format
func ToFilename(id string) string {
	if !strings.HasSuffix(id, ".md") {
		return id + ".md"
	}
	return id
}

// FromFilename extracts ULID from filename format
func FromFilename(filename string) string {
	return strings.TrimSuffix(filename, ".md")
}

// Validate performs comprehensive validation on a ULID string
func Validate(id string) error {
	if id == "" {
		return NewULIDError("ULID cannot be empty")
	}

	if len(id) != 26 {
		return NewULIDError("ULID must be exactly 26 characters long")
	}

	if _, err := ulid.Parse(id); err != nil {
		return NewULIDError("invalid ULID format: " + err.Error())
	}

	return nil
}

// ValidateFilename validates a filename containing a ULID
func ValidateFilename(filename string) error {
	if filename == "" {
		return NewULIDError("filename cannot be empty")
	}

	if !strings.HasSuffix(filename, ".md") {
		return NewULIDError("filename must have .md extension")
	}

	id := strings.TrimSuffix(filename, ".md")
	return Validate(id)
}

// ULIDError represents an error related to ULID operations
type ULIDError struct {
	Message string
}

func (e *ULIDError) Error() string {
	return "ULID error: " + e.Message
}

// NewULIDError creates a new ULID error
func NewULIDError(message string) *ULIDError {
	return &ULIDError{Message: message}
}

// Global generator instance for convenience
var defaultGenerator = NewGenerator()

// New generates a new ULID using the default generator
func New() string {
	return defaultGenerator.Generate()
}

// NewWithTimestamp generates a new ULID with a specific timestamp
func NewWithTimestamp(t time.Time) string {
	return defaultGenerator.GenerateAt(t)
}

// NewFilename generates a new ULID filename
func NewFilename() string {
	return defaultGenerator.GenerateWithPrefix()
}
