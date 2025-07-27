package main

import (
	"fmt"
	"os"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
	"github.com/madstone-tech/mdstn-kb-mcp/pkg/ulid"
)

var (
	// Build information set via ldflags
	version    = "dev"
	commitHash = "unknown"
	buildTime  = "unknown"
)

func main() {
	// Check for flags
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("kbVault v%s\n", version)
			fmt.Printf("Built: %s (commit: %s)\n", buildTime, commitHash)
			return
		case "--help", "-h":
			fmt.Printf("kbVault v%s - High-performance Go knowledge management tool\n", version)
			fmt.Println("\nUsage:")
			fmt.Println("  kbvault [flags]")
			fmt.Println("\nFlags:")
			fmt.Println("  -h, --help     Show this help message")
			fmt.Println("  -v, --version  Show version information")
			fmt.Println("\nSession 2 Demo:")
			fmt.Println("  Currently demonstrates foundation features from Session 1")
			return
		}
	}

	// Placeholder main function with basic functionality demonstration
	fmt.Printf("kbVault v%s\n", version)
	fmt.Printf("Built: %s (commit: %s)\n", buildTime, commitHash)
	fmt.Println("\nkbVault - High-performance Go knowledge management tool")

	// Demonstrate core functionality from Session 1
	fmt.Println("\n=== Session 1 Demo: Foundation & Core Types ===")

	// Test ULID generation
	fmt.Println("\n1. ULID Generation:")
	id := ulid.New()
	fmt.Printf("   Generated ULID: %s\n", id)
	fmt.Printf("   Is valid: %v\n", ulid.IsValid(id))

	timestamp, _ := ulid.ExtractTimestamp(id)
	fmt.Printf("   Timestamp: %s\n", timestamp.Format("2006-01-02 15:04:05"))

	// Test configuration
	fmt.Println("\n2. Configuration:")
	config := types.DefaultConfig()
	fmt.Printf("   Vault name: %s\n", config.Vault.Name)
	fmt.Printf("   Storage type: %s\n", config.Storage.Type)
	fmt.Printf("   HTTP port: %d\n", config.Server.HTTP.Port)

	// Test validation
	fmt.Println("\n3. Configuration Validation:")
	if err := config.Validate(); err != nil {
		fmt.Printf("   Validation failed: %v\n", err)
	} else {
		fmt.Printf("   Configuration is valid ✓\n")
	}

	// Test note creation
	fmt.Println("\n4. Note Structure:")
	note := &types.Note{
		ID:             id,
		Title:          "Demo Note",
		Content:        "This is a demonstration note created in Session 1",
		FilePath:       "notes/" + ulid.ToFilename(id),
		StorageBackend: types.StorageTypeLocal,
		Frontmatter: types.Frontmatter{
			ID:      id,
			Title:   "Demo Note",
			Tags:    []string{"demo", "session-1"},
			Type:    "note",
			Storage: "local",
		},
	}

	fmt.Printf("   Note ID: %s\n", note.ID)
	fmt.Printf("   Title: %s\n", note.Title)
	fmt.Printf("   File path: %s\n", note.FilePath)
	fmt.Printf("   Tags: %v\n", note.Frontmatter.Tags)

	if err := note.Validate(); err != nil {
		fmt.Printf("   Note validation failed: %v\n", err)
	} else {
		fmt.Printf("   Note is valid ✓\n")
	}

	fmt.Println("\n=== Session 1 Complete ===")
	fmt.Println("✓ MIT License")
	fmt.Println("✓ Go modules and dependencies")
	fmt.Println("✓ Makefile targets")
	fmt.Println("✓ Core data types (Note, Config, Storage)")
	fmt.Println("✓ ULID integration with validation")
	fmt.Println("✓ Basic testing infrastructure (>80% coverage)")

	fmt.Println("\nNext: Session 2 - Configuration & Local Storage")
}
