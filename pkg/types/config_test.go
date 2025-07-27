package types

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Test vault defaults
	if config.Vault.Name == "" {
		t.Error("Default config should have vault name")
	}

	if config.Vault.MaxFileSize != 10*1024*1024 {
		t.Errorf("Expected max file size 10MB, got %d", config.Vault.MaxFileSize)
	}

	if config.Vault.NotesDir != "notes" {
		t.Errorf("Expected notes dir 'notes', got %s", config.Vault.NotesDir)
	}

	// Test storage defaults
	if config.Storage.Type != StorageTypeLocal {
		t.Errorf("Expected default storage type %s, got %s", StorageTypeLocal, config.Storage.Type)
	}

	if config.Storage.Local.Path != "./vault" {
		t.Errorf("Expected default local path './vault', got %s", config.Storage.Local.Path)
	}

	// Test server defaults
	if !config.Server.HTTP.Enabled {
		t.Error("HTTP server should be enabled by default")
	}

	if config.Server.HTTP.Port != 8080 {
		t.Errorf("Expected default HTTP port 8080, got %d", config.Server.HTTP.Port)
	}

	if config.Server.GRPC.Enabled {
		t.Error("gRPC server should be disabled by default")
	}

	// Test auth defaults
	if config.Server.Auth.Type != "none" {
		t.Errorf("Expected default auth type 'none', got %s", config.Server.Auth.Type)
	}

	// Test logging defaults
	if config.Logging.Level != "WARN" {
		t.Errorf("Expected default log level 'WARN', got %s", config.Logging.Level)
	}

	// Test MCP defaults
	if !config.MCP.Enabled {
		t.Error("MCP should be enabled by default")
	}
}

func TestConfig_Validate(t *testing.T) {
	testCases := []struct {
		name        string
		modifyFunc  func(*Config)
		expectError bool
	}{
		{
			name:        "default config valid",
			modifyFunc:  func(c *Config) {},
			expectError: false,
		},
		{
			name: "empty vault name",
			modifyFunc: func(c *Config) {
				c.Vault.Name = ""
			},
			expectError: true,
		},
		{
			name: "zero max file size",
			modifyFunc: func(c *Config) {
				c.Vault.MaxFileSize = 0
			},
			expectError: true,
		},
		{
			name: "invalid storage type",
			modifyFunc: func(c *Config) {
				c.Storage.Type = StorageType("invalid")
			},
			expectError: true,
		},
		{
			name: "invalid HTTP port",
			modifyFunc: func(c *Config) {
				c.Server.HTTP.Port = 0
			},
			expectError: true,
		},
		{
			name: "HTTP port too high",
			modifyFunc: func(c *Config) {
				c.Server.HTTP.Port = 70000
			},
			expectError: true,
		},
		{
			name: "invalid gRPC port",
			modifyFunc: func(c *Config) {
				c.Server.GRPC.Enabled = true
				c.Server.GRPC.Port = 0
			},
			expectError: true,
		},
		{
			name: "same HTTP and gRPC ports",
			modifyFunc: func(c *Config) {
				c.Server.HTTP.Enabled = true
				c.Server.GRPC.Enabled = true
				c.Server.HTTP.Port = 8080
				c.Server.GRPC.Port = 8080
			},
			expectError: true,
		},
		{
			name: "invalid auth type",
			modifyFunc: func(c *Config) {
				c.Server.Auth.Type = "invalid"
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := DefaultConfig()
			tc.modifyFunc(config)

			err := config.Validate()

			if tc.expectError && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.expectError && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestStorageType_Constants(t *testing.T) {
	if StorageTypeLocal != "local" {
		t.Errorf("Expected StorageTypeLocal to be 'local', got %s", StorageTypeLocal)
	}

	if StorageTypeS3 != "s3" {
		t.Errorf("Expected StorageTypeS3 to be 's3', got %s", StorageTypeS3)
	}
}

func TestVaultConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	vault := config.Vault

	if vault.DateFormat != "2006-01-02" {
		t.Errorf("Expected date format '2006-01-02', got %s", vault.DateFormat)
	}

	if vault.TimeFormat != "15:04:05" {
		t.Errorf("Expected time format '15:04:05', got %s", vault.TimeFormat)
	}

	if !vault.AutoSave {
		t.Error("AutoSave should be enabled by default")
	}

	if vault.AutoSync {
		t.Error("AutoSync should be disabled by default")
	}
}

func TestHTTPServerConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	http := config.Server.HTTP

	if http.Host != "localhost" {
		t.Errorf("Expected HTTP host 'localhost', got %s", http.Host)
	}

	if !http.EnableCORS {
		t.Error("CORS should be enabled by default")
	}

	if len(http.CORSOrigins) == 0 {
		t.Error("CORS origins should be configured by default")
	}

	if http.ReadTimeout != 30 {
		t.Errorf("Expected read timeout 30, got %d", http.ReadTimeout)
	}

	if http.MaxRequestSize != 10*1024*1024 {
		t.Errorf("Expected max request size 10MB, got %d", http.MaxRequestSize)
	}
}

func TestGRPCServerConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	grpc := config.Server.GRPC

	if grpc.Enabled {
		t.Error("gRPC should be disabled by default")
	}

	if grpc.Host != "localhost" {
		t.Errorf("Expected gRPC host 'localhost', got %s", grpc.Host)
	}

	if grpc.Port != 9090 {
		t.Errorf("Expected gRPC port 9090, got %d", grpc.Port)
	}

	if grpc.MaxRecvMsgSize != 4*1024*1024 {
		t.Errorf("Expected max recv msg size 4MB, got %d", grpc.MaxRecvMsgSize)
	}

	if !grpc.EnableAgentService {
		t.Error("Agent service should be enabled by default")
	}
}

func TestCacheConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	cache := config.Storage.Cache

	if cache.Enabled {
		t.Error("Cache should be disabled by default")
	}

	if !cache.AutoEnable {
		t.Error("Cache auto-enable should be true by default")
	}

	if cache.Memory.MaxSizeMB != 100 {
		t.Errorf("Expected memory cache size 100MB, got %d", cache.Memory.MaxSizeMB)
	}

	if cache.Memory.TTLMinutes != 5 {
		t.Errorf("Expected memory TTL 5 minutes, got %d", cache.Memory.TTLMinutes)
	}

	if cache.Disk.TTLHours != 24 {
		t.Errorf("Expected disk TTL 24 hours, got %d", cache.Disk.TTLHours)
	}
}

func TestTUIConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	tui := config.TUI

	if tui.Theme != "default" {
		t.Errorf("Expected TUI theme 'default', got %s", tui.Theme)
	}

	if tui.VimMode {
		t.Error("Vim mode should be disabled by default")
	}

	if !tui.ShowHelp {
		t.Error("Show help should be enabled by default")
	}

	if tui.PageSize != 20 {
		t.Errorf("Expected page size 20, got %d", tui.PageSize)
	}

	if !tui.EnableMouse {
		t.Error("Mouse should be enabled by default")
	}
}

func TestMCPConfig_Defaults(t *testing.T) {
	config := DefaultConfig()
	mcp := config.MCP

	if !mcp.Enabled {
		t.Error("MCP should be enabled by default")
	}

	if mcp.SocketPath != "/tmp/kbvault.sock" {
		t.Errorf("Expected socket path '/tmp/kbvault.sock', got %s", mcp.SocketPath)
	}

	if mcp.UseStdio {
		t.Error("UseStdio should be false by default")
	}

	if mcp.MaxRequestSize != 10*1024*1024 {
		t.Errorf("Expected max request size 10MB, got %d", mcp.MaxRequestSize)
	}

	if !mcp.EnableBulkOperations {
		t.Error("Bulk operations should be enabled by default")
	}
}
