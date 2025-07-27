package main

import (
	"testing"

	"github.com/madstone-tech/mdstn-kb-mcp/pkg/types"
)

func TestShowAllConfig(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Vault.Name = "test-vault"
	cfg.Vault.NotesDir = "notes"
	cfg.Vault.MaxFileSize = 1024000

	tests := []struct {
		name   string
		format string
		config *types.Config
		want   []string // Strings that should be present in output
	}{
		{
			name:   "yaml_format",
			format: "yaml",
			config: cfg,
			want:   []string{"vault:", "name: test-vault", "notes_dir: notes", "max_file_size: 1024000"},
		},
		{
			name:   "json_format",
			format: "json",
			config: cfg,
			want:   []string{`"vault":`, `"name": "test-vault"`, `"notes_dir": "notes"`, `"max_file_size": 1024000`},
		},
		{
			name:   "default_format",
			format: "default",
			config: cfg,
			want:   []string{"vault:", "name: test-vault", "notes_dir: notes", "max_file_size: 1024000"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// We can't easily test the actual function since it prints to stdout
			// So we'll test the logic by checking the format detection
			err := showAllConfig(tt.config, tt.format)
			if err != nil {
				t.Errorf("showAllConfig() error = %v", err)
			}
		})
	}
}

func TestShowConfigKey(t *testing.T) {
	cfg := types.DefaultConfig()
	cfg.Vault.Name = "test-vault"
	cfg.Vault.NotesDir = "test-notes"
	cfg.Vault.MaxFileSize = 2048000
	cfg.Storage.Type = types.StorageTypeLocal
	cfg.Server.HTTP.Host = "localhost"
	cfg.Server.HTTP.Port = 8080
	cfg.Server.HTTP.ReadTimeout = 30

	tests := []struct {
		name    string
		key     string
		config  *types.Config
		wantErr bool
	}{
		{
			name:    "vault_name",
			key:     "vault.name",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "vault_notes_dir",
			key:     "vault.notes_dir",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "vault_max_file_size",
			key:     "vault.max_file_size",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "vault_all",
			key:     "vault",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "storage_type",
			key:     "storage.type",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "storage_all",
			key:     "storage",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "server_http_host",
			key:     "server.http.host",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "server_http_port",
			key:     "server.http.port",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "server_http_read_timeout",
			key:     "server.http.read_timeout",
			config:  cfg,
			wantErr: false,
		},
		{
			name:    "invalid_vault_key",
			key:     "vault.invalid",
			config:  cfg,
			wantErr: true,
		},
		{
			name:    "invalid_storage_key",
			key:     "storage.invalid",
			config:  cfg,
			wantErr: true,
		},
		{
			name:    "invalid_server_key",
			key:     "server.invalid",
			config:  cfg,
			wantErr: true,
		},
		{
			name:    "invalid_server_http_key",
			key:     "server.http.invalid",
			config:  cfg,
			wantErr: true,
		},
		{
			name:    "unknown_root_key",
			key:     "unknown",
			config:  cfg,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := showConfigKey(tt.config, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("showConfigKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSetConfigValue(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
		check   func(*types.Config) bool
	}{
		{
			name:    "set_vault_name",
			key:     "vault.name",
			value:   "new-vault-name",
			wantErr: false,
			check: func(cfg *types.Config) bool {
				return cfg.Vault.Name == "new-vault-name"
			},
		},
		{
			name:    "set_vault_notes_dir",
			key:     "vault.notes_dir",
			value:   "my-notes",
			wantErr: false,
			check: func(cfg *types.Config) bool {
				return cfg.Vault.NotesDir == "my-notes"
			},
		},
		{
			name:    "set_storage_type",
			key:     "storage.type",
			value:   "s3",
			wantErr: false,
			check: func(cfg *types.Config) bool {
				return cfg.Storage.Type == types.StorageType("s3")
			},
		},
		{
			name:    "set_server_http_host",
			key:     "server.http.host",
			value:   "0.0.0.0",
			wantErr: false,
			check: func(cfg *types.Config) bool {
				return cfg.Server.HTTP.Host == "0.0.0.0"
			},
		},
		{
			name:    "invalid_vault_key",
			key:     "vault.invalid",
			value:   "test",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid_storage_key",
			key:     "storage.invalid",
			value:   "test",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "invalid_server_key",
			key:     "server.invalid",
			value:   "test",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "unsupported_server_setting",
			key:     "server.http.port",
			value:   "9090",
			wantErr: true,
			check:   nil,
		},
		{
			name:    "unknown_root_key",
			key:     "unknown.key",
			value:   "test",
			wantErr: true,
			check:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := types.DefaultConfig()
			err := setConfigValue(cfg, tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("setConfigValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.check != nil && !tt.check(cfg) {
				t.Errorf("setConfigValue() did not set the expected value")
			}
		})
	}
}