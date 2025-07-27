package ulid

import (
	"strings"
	"testing"
	"time"
)

func TestGenerator_Generate(t *testing.T) {
	gen := NewGenerator()

	// Test basic generation
	id1 := gen.Generate()
	if len(id1) != 26 {
		t.Errorf("Expected ULID length 26, got %d", len(id1))
	}

	// Test uniqueness
	id2 := gen.Generate()
	if id1 == id2 {
		t.Error("Generated ULIDs should be unique")
	}

	// Test validity
	if !IsValid(id1) {
		t.Errorf("Generated ULID should be valid: %s", id1)
	}
}

func TestGenerator_GenerateAt(t *testing.T) {
	gen := NewGenerator()
	testTime := time.Date(2024, 1, 15, 14, 23, 0, 0, time.UTC)

	id := gen.GenerateAt(testTime)

	// Extract timestamp
	extractedTime, err := ExtractTimestamp(id)
	if err != nil {
		t.Fatalf("Failed to extract timestamp: %v", err)
	}

	// Should be within the same millisecond
	if extractedTime.Unix() != testTime.Unix() {
		t.Errorf("Expected timestamp %v, got %v", testTime.Unix(), extractedTime.Unix())
	}
}

func TestGenerator_GenerateWithPrefix(t *testing.T) {
	gen := NewGenerator()

	filename := gen.GenerateWithPrefix()

	if !strings.HasSuffix(filename, ".md") {
		t.Error("Generated filename should have .md suffix")
	}

	// Extract ULID part
	id := strings.TrimSuffix(filename, ".md")
	if !IsValid(id) {
		t.Errorf("ULID part should be valid: %s", id)
	}
}

func TestParseFromFilename(t *testing.T) {
	testCases := []struct {
		filename string
		valid    bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV.md", true},
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"invalid-ulid.md", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			_, err := ParseFromFilename(tc.filename)

			if tc.valid && err != nil {
				t.Errorf("Expected valid filename %s, got error: %v", tc.filename, err)
			}
			if !tc.valid && err == nil {
				t.Errorf("Expected invalid filename %s to return error", tc.filename)
			}
		})
	}
}

func TestExtractTimestamp(t *testing.T) {
	// Generate ULID with known timestamp
	gen := NewGenerator()
	testTime := time.Date(2024, 1, 15, 14, 23, 0, 0, time.UTC)
	id := gen.GenerateAt(testTime)

	extractedTime, err := ExtractTimestamp(id)
	if err != nil {
		t.Fatalf("Failed to extract timestamp: %v", err)
	}

	// Times should be equal within millisecond precision
	timeDiff := extractedTime.Sub(testTime)
	if timeDiff > time.Millisecond || timeDiff < -time.Millisecond {
		t.Errorf("Time difference too large: %v", timeDiff)
	}
}

func TestIsValid(t *testing.T) {
	testCases := []struct {
		id    string
		valid bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"", false},
		{"invalid", false},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA", false},   // too short
		{"01ARZ3NDEKTSV4RRFFQ69G5FAVX", false}, // too long
		{"01ARZ3NDEKTSV4RRFFQ69G5FI", false},   // invalid character 'I'
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			result := IsValid(tc.id)
			if result != tc.valid {
				t.Errorf("IsValid(%s) = %v, expected %v", tc.id, result, tc.valid)
			}
		})
	}
}

func TestIsValidFilename(t *testing.T) {
	testCases := []struct {
		filename string
		valid    bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV.md", true},
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true},
		{"invalid.md", false},
		{"", false},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA.md", false}, // too short
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			result := IsValidFilename(tc.filename)
			if result != tc.valid {
				t.Errorf("IsValidFilename(%s) = %v, expected %v", tc.filename, result, tc.valid)
			}
		})
	}
}

func TestCompare(t *testing.T) {
	gen := NewGenerator()

	// Generate two ULIDs at different times
	id1 := gen.Generate()
	time.Sleep(time.Millisecond) // Ensure different timestamps
	id2 := gen.Generate()

	// Compare should return -1 (id1 < id2)
	result := Compare(id1, id2)
	if result != -1 {
		t.Errorf("Expected Compare(%s, %s) = -1, got %d", id1, id2, result)
	}

	// Same ID should return 0
	result = Compare(id1, id1)
	if result != 0 {
		t.Errorf("Expected Compare(%s, %s) = 0, got %d", id1, id1, result)
	}
}

func TestFilterByTimeRange(t *testing.T) {
	gen := NewGenerator()

	// Generate ULIDs with specific timestamps
	baseTime := time.Date(2024, 1, 15, 14, 0, 0, 0, time.UTC)

	var ids []string
	for i := 0; i < 5; i++ {
		testTime := baseTime.Add(time.Duration(i) * time.Hour)
		id := gen.GenerateAt(testTime)
		ids = append(ids, id)
	}

	// Filter for middle range
	timeRange := TimestampRange{
		Start: baseTime.Add(1 * time.Hour),
		End:   baseTime.Add(3 * time.Hour),
	}

	filtered, err := FilterByTimeRange(ids, timeRange)
	if err != nil {
		t.Fatalf("FilterByTimeRange failed: %v", err)
	}

	// Should include IDs at hours 1, 2, 3 (3 total)
	if len(filtered) != 3 {
		t.Errorf("Expected 3 filtered ULIDs, got %d", len(filtered))
	}
}

func TestAgeInDays(t *testing.T) {
	gen := NewGenerator()

	// Generate ULID from 3 days ago
	pastTime := time.Now().Add(-3 * 24 * time.Hour)
	id := gen.GenerateAt(pastTime)

	age, err := AgeInDays(id)
	if err != nil {
		t.Fatalf("AgeInDays failed: %v", err)
	}

	// Should be approximately 3 days (allow for slight timing differences)
	if age < 2 || age > 4 {
		t.Errorf("Expected age around 3 days, got %d", age)
	}
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		id          string
		expectError bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", false},
		{"", true},
		{"invalid", true},
		{"01ARZ3NDEKTSV4RRFFQ69G5FA", true}, // too short
	}

	for _, tc := range testCases {
		t.Run(tc.id, func(t *testing.T) {
			err := Validate(tc.id)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for ID %s, got nil", tc.id)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for ID %s, got: %v", tc.id, err)
			}
		})
	}
}

func TestValidateFilename(t *testing.T) {
	testCases := []struct {
		filename    string
		expectError bool
	}{
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV.md", false},
		{"", true},
		{"invalid.md", true},
		{"01ARZ3NDEKTSV4RRFFQ69G5FAV", true}, // missing .md extension
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			err := ValidateFilename(tc.filename)

			if tc.expectError && err == nil {
				t.Errorf("Expected error for filename %s, got nil", tc.filename)
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no error for filename %s, got: %v", tc.filename, err)
			}
		})
	}
}

func TestGlobalFunctions(t *testing.T) {
	// Test global convenience functions
	id := New()
	if !IsValid(id) {
		t.Errorf("New() should generate valid ULID, got: %s", id)
	}

	filename := NewFilename()
	if !strings.HasSuffix(filename, ".md") {
		t.Error("NewFilename() should generate filename with .md extension")
	}

	testTime := time.Date(2024, 1, 15, 14, 23, 0, 0, time.UTC)
	timestampedID := NewWithTimestamp(testTime)
	if !IsValid(timestampedID) {
		t.Errorf("NewWithTimestamp() should generate valid ULID, got: %s", timestampedID)
	}
}

// Benchmark tests
func BenchmarkGenerator_Generate(b *testing.B) {
	gen := NewGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		gen.Generate()
	}
}

func BenchmarkIsValid(b *testing.B) {
	gen := NewGenerator()
	id := gen.Generate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsValid(id)
	}
}

func BenchmarkExtractTimestamp(b *testing.B) {
	gen := NewGenerator()
	id := gen.Generate()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ExtractTimestamp(id)
	}
}

// Test concurrent generation
func TestConcurrentGeneration(t *testing.T) {
	gen := NewGenerator()
	const numGoroutines = 100
	const numULIDs = 10

	results := make(chan string, numGoroutines*numULIDs)

	// Launch goroutines
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < numULIDs; j++ {
				results <- gen.Generate()
			}
		}()
	}

	// Collect results
	seen := make(map[string]bool)
	for i := 0; i < numGoroutines*numULIDs; i++ {
		id := <-results

		// Check validity
		if !IsValid(id) {
			t.Errorf("Generated invalid ULID: %s", id)
		}

		// Check uniqueness
		if seen[id] {
			t.Errorf("Duplicate ULID generated: %s", id)
		}
		seen[id] = true
	}
}
