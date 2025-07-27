package types

// Config represents the complete kbVault configuration
type Config struct {
	// Vault configuration
	Vault VaultConfig `toml:"vault" json:"vault"`

	// Storage backend configuration
	Storage StorageConfig `toml:"storage" json:"storage"`

	// Server configuration
	Server ServerConfig `toml:"server" json:"server"`

	// Logging configuration
	Logging LoggingConfig `toml:"logging" json:"logging"`

	// TUI configuration
	TUI TUIConfig `toml:"tui" json:"tui"`

	// MCP configuration
	MCP MCPConfig `toml:"mcp" json:"mcp"`
}

// VaultConfig contains vault-specific settings
type VaultConfig struct {
	// Name of the vault
	Name string `toml:"name" json:"name"`

	// NotesDir is the subdirectory for regular notes
	NotesDir string `toml:"notes_dir" json:"notes_dir"`

	// DailyDir is the subdirectory for daily notes
	DailyDir string `toml:"daily_dir" json:"daily_dir"`

	// TemplatesDir is the subdirectory for note templates
	TemplatesDir string `toml:"templates_dir" json:"templates_dir"`

	// DefaultTemplate is the template to use for new notes
	DefaultTemplate string `toml:"default_template" json:"default_template"`

	// MaxFileSize is the maximum allowed file size in bytes
	MaxFileSize int64 `toml:"max_file_size" json:"max_file_size"`

	// DateFormat for daily notes and timestamps
	DateFormat string `toml:"date_format" json:"date_format"`

	// TimeFormat for timestamps
	TimeFormat string `toml:"time_format" json:"time_format"`

	// AutoSave enables automatic saving of modified notes
	AutoSave bool `toml:"auto_save" json:"auto_save"`

	// AutoSync enables automatic synchronization with remote storage
	AutoSync bool `toml:"auto_sync" json:"auto_sync"`
}

// ServerConfig contains HTTP and gRPC server settings
type ServerConfig struct {
	// HTTP server configuration
	HTTP HTTPServerConfig `toml:"http" json:"http"`

	// gRPC server configuration (optional)
	GRPC GRPCServerConfig `toml:"grpc" json:"grpc"`

	// Authentication configuration
	Auth AuthConfig `toml:"auth" json:"auth"`
}

// HTTPServerConfig configures the HTTP REST API server
type HTTPServerConfig struct {
	// Enabled turns the HTTP server on/off
	Enabled bool `toml:"enabled" json:"enabled"`

	// Host to bind to
	Host string `toml:"host" json:"host"`

	// Port to listen on
	Port int `toml:"port" json:"port"`

	// EnableCORS allows cross-origin requests
	EnableCORS bool `toml:"enable_cors" json:"enable_cors"`

	// CORSOrigins specifies allowed CORS origins
	CORSOrigins []string `toml:"cors_origins" json:"cors_origins"`

	// ReadTimeout for requests (seconds)
	ReadTimeout int `toml:"read_timeout" json:"read_timeout"`

	// WriteTimeout for responses (seconds)
	WriteTimeout int `toml:"write_timeout" json:"write_timeout"`

	// IdleTimeout for idle connections (seconds)
	IdleTimeout int `toml:"idle_timeout" json:"idle_timeout"`

	// MaxRequestSize limits request body size (bytes)
	MaxRequestSize int64 `toml:"max_request_size" json:"max_request_size"`

	// TLS configuration
	TLS TLSConfig `toml:"tls" json:"tls"`
}

// GRPCServerConfig configures the gRPC server
type GRPCServerConfig struct {
	// Enabled turns the gRPC server on/off
	Enabled bool `toml:"enabled" json:"enabled"`

	// Host to bind to
	Host string `toml:"host" json:"host"`

	// Port to listen on
	Port int `toml:"port" json:"port"`

	// MaxRecvMsgSize limits incoming message size
	MaxRecvMsgSize int `toml:"max_recv_msg_size" json:"max_recv_msg_size"`

	// MaxSendMsgSize limits outgoing message size
	MaxSendMsgSize int `toml:"max_send_msg_size" json:"max_send_msg_size"`

	// ConnectionTimeout for client connections (seconds)
	ConnectionTimeout int `toml:"connection_timeout" json:"connection_timeout"`

	// TLS configuration
	TLS TLSConfig `toml:"tls" json:"tls"`

	// EnableBulkOperations enables bulk operation services
	EnableBulkOperations bool `toml:"enable_bulk_operations" json:"enable_bulk_operations"`

	// EnableCollaboration enables real-time collaboration features
	EnableCollaboration bool `toml:"enable_collaboration" json:"enable_collaboration"`

	// EnableAgentService enables AI agent-specific services
	EnableAgentService bool `toml:"enable_agent_service" json:"enable_agent_service"`
}

// TLSConfig configures TLS/SSL settings
type TLSConfig struct {
	// Enabled turns on TLS
	Enabled bool `toml:"enabled" json:"enabled"`

	// CertFile path to certificate file
	CertFile string `toml:"cert_file" json:"cert_file"`

	// KeyFile path to private key file
	KeyFile string `toml:"key_file" json:"key_file"`

	// CAFile path to CA certificate (for client cert verification)
	CAFile string `toml:"ca_file" json:"ca_file"`

	// RequireClientCert forces client certificate authentication
	RequireClientCert bool `toml:"require_client_cert" json:"require_client_cert"`
}

// AuthConfig configures authentication and authorization
type AuthConfig struct {
	// Type of authentication (none, apikey, jwt)
	Type string `toml:"type" json:"type"`

	// APIKeys for API key authentication
	APIKeys []string `toml:"api_keys" json:"api_keys"`

	// JWT configuration
	JWT JWTConfig `toml:"jwt" json:"jwt"`

	// RateLimit configuration
	RateLimit RateLimitConfig `toml:"rate_limit" json:"rate_limit"`
}

// JWTConfig configures JWT token authentication
type JWTConfig struct {
	// Secret for signing tokens
	Secret string `toml:"secret" json:"secret"`

	// Issuer for token validation
	Issuer string `toml:"issuer" json:"issuer"`

	// Audience for token validation
	Audience string `toml:"audience" json:"audience"`

	// ExpiryHours for token expiration
	ExpiryHours int `toml:"expiry_hours" json:"expiry_hours"`
}

// RateLimitConfig configures API rate limiting
type RateLimitConfig struct {
	// Enabled turns on rate limiting
	Enabled bool `toml:"enabled" json:"enabled"`

	// RequestsPerMinute limit per IP
	RequestsPerMinute int `toml:"requests_per_minute" json:"requests_per_minute"`

	// BurstSize allows temporary bursts above the rate limit
	BurstSize int `toml:"burst_size" json:"burst_size"`
}

// LoggingConfig configures application logging
type LoggingConfig struct {
	// Level sets the log level (DEBUG, INFO, WARN, ERROR)
	Level string `toml:"level" json:"level"`

	// Output destination (stdout, file, remote)
	Output string `toml:"output" json:"output"`

	// FilePath for file-based logging
	FilePath string `toml:"file_path" json:"file_path"`

	// Format for log messages (text, json)
	Format string `toml:"format" json:"format"`

	// EnableColors adds color to console output
	EnableColors bool `toml:"enable_colors" json:"enable_colors"`

	// EnableTimestamp includes timestamps in logs
	EnableTimestamp bool `toml:"enable_timestamp" json:"enable_timestamp"`

	// EnableCaller includes caller information
	EnableCaller bool `toml:"enable_caller" json:"enable_caller"`

	// RotateSize triggers log rotation at this size (MB)
	RotateSize int `toml:"rotate_size" json:"rotate_size"`

	// RotateCount is the number of old log files to keep
	RotateCount int `toml:"rotate_count" json:"rotate_count"`
}

// TUIConfig configures the terminal user interface
type TUIConfig struct {
	// Theme for the TUI (default, dark, light)
	Theme string `toml:"theme" json:"theme"`

	// VimMode enables vim-style key bindings
	VimMode bool `toml:"vim_mode" json:"vim_mode"`

	// ShowHelp displays help information
	ShowHelp bool `toml:"show_help" json:"show_help"`

	// RefreshInterval for auto-refresh (seconds)
	RefreshInterval int `toml:"refresh_interval" json:"refresh_interval"`

	// PageSize for paginated lists
	PageSize int `toml:"page_size" json:"page_size"`

	// EnableMouse allows mouse interaction
	EnableMouse bool `toml:"enable_mouse" json:"enable_mouse"`
}

// MCPConfig configures Model Context Protocol integration
type MCPConfig struct {
	// Enabled turns on MCP server
	Enabled bool `toml:"enabled" json:"enabled"`

	// SocketPath for Unix socket communication
	SocketPath string `toml:"socket_path" json:"socket_path"`

	// UseStdio uses stdin/stdout instead of socket
	UseStdio bool `toml:"use_stdio" json:"use_stdio"`

	// MaxRequestSize limits MCP request size
	MaxRequestSize int64 `toml:"max_request_size" json:"max_request_size"`

	// ResponseTimeout for MCP operations (seconds)
	ResponseTimeout int `toml:"response_timeout" json:"response_timeout"`

	// EnableBulkOperations allows bulk note operations
	EnableBulkOperations bool `toml:"enable_bulk_operations" json:"enable_bulk_operations"`

	// MaxBulkSize limits bulk operation size
	MaxBulkSize int `toml:"max_bulk_size" json:"max_bulk_size"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Vault: VaultConfig{
			Name:            "my-kb",
			NotesDir:        "notes",
			DailyDir:        "notes/dailies",
			TemplatesDir:    "templates",
			DefaultTemplate: "default",
			MaxFileSize:     10 * 1024 * 1024, // 10MB
			DateFormat:      "2006-01-02",
			TimeFormat:      "15:04:05",
			AutoSave:        true,
			AutoSync:        false,
		},
		Storage: StorageConfig{
			Type: StorageTypeLocal,
			Local: LocalStorageConfig{
				Path:          "./vault",
				CreateDirs:    true,
				DirPerms:      "0755",
				FilePerms:     "0644",
				EnableLocking: true,
				LockTimeout:   10,
			},
			Cache: CacheConfig{
				Enabled:    false,
				AutoEnable: true,
				Memory: MemoryCacheConfig{
					Enabled:    true,
					MaxSizeMB:  100,
					MaxItems:   1000,
					TTLMinutes: 5,
				},
				Disk: DiskCacheConfig{
					Enabled:              true,
					Path:                 "/tmp/kbvault-cache",
					MaxSizeMB:            1000,
					TTLHours:             24,
					CleanupIntervalHours: 6,
				},
			},
		},
		Server: ServerConfig{
			HTTP: HTTPServerConfig{
				Enabled:        true,
				Host:           "localhost",
				Port:           8080,
				EnableCORS:     true,
				CORSOrigins:    []string{"*"},
				ReadTimeout:    30,
				WriteTimeout:   30,
				IdleTimeout:    60,
				MaxRequestSize: 10 * 1024 * 1024, // 10MB
			},
			GRPC: GRPCServerConfig{
				Enabled:              false,
				Host:                 "localhost",
				Port:                 9090,
				MaxRecvMsgSize:       4 * 1024 * 1024, // 4MB
				MaxSendMsgSize:       4 * 1024 * 1024, // 4MB
				ConnectionTimeout:    30,
				EnableBulkOperations: false,
				EnableCollaboration:  false,
				EnableAgentService:   true,
			},
			Auth: AuthConfig{
				Type: "none",
				RateLimit: RateLimitConfig{
					Enabled:           false,
					RequestsPerMinute: 100,
					BurstSize:         10,
				},
			},
		},
		Logging: LoggingConfig{
			Level:           "WARN",
			Output:          "stdout",
			Format:          "text",
			EnableColors:    true,
			EnableTimestamp: true,
			EnableCaller:    false,
			RotateSize:      100,
			RotateCount:     5,
		},
		TUI: TUIConfig{
			Theme:           "default",
			VimMode:         false,
			ShowHelp:        true,
			RefreshInterval: 30,
			PageSize:        20,
			EnableMouse:     true,
		},
		MCP: MCPConfig{
			Enabled:              true,
			SocketPath:           "/tmp/kbvault.sock",
			UseStdio:             false,
			MaxRequestSize:       10 * 1024 * 1024, // 10MB
			ResponseTimeout:      30,
			EnableBulkOperations: true,
			MaxBulkSize:          100,
		},
	}
}

// Validate performs validation on the configuration
func (c *Config) Validate() error {
	// Validate vault config
	if c.Vault.Name == "" {
		return NewValidationError("vault name cannot be empty")
	}
	if c.Vault.MaxFileSize <= 0 {
		return NewValidationError("vault max_file_size must be positive")
	}

	// Validate storage config
	if c.Storage.Type != StorageTypeLocal && c.Storage.Type != StorageTypeS3 {
		return NewValidationError("storage type must be 'local' or 's3'")
	}

	// Validate server config
	if c.Server.HTTP.Enabled {
		if c.Server.HTTP.Port <= 0 || c.Server.HTTP.Port > 65535 {
			return NewValidationError("HTTP port must be between 1 and 65535")
		}
	}
	if c.Server.GRPC.Enabled {
		if c.Server.GRPC.Port <= 0 || c.Server.GRPC.Port > 65535 {
			return NewValidationError("gRPC port must be between 1 and 65535")
		}
		if c.Server.HTTP.Enabled && c.Server.HTTP.Port == c.Server.GRPC.Port {
			return NewValidationError("HTTP and gRPC ports cannot be the same")
		}
	}

	// Validate auth config
	if c.Server.Auth.Type != "none" && c.Server.Auth.Type != "apikey" && c.Server.Auth.Type != "jwt" {
		return NewValidationError("auth type must be 'none', 'apikey', or 'jwt'")
	}

	return nil
}
