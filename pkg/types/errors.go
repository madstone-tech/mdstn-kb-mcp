package types

import (
	"fmt"
)

// ErrorType represents different categories of errors
type ErrorType string

const (
	ErrorTypeValidation   ErrorType = "validation"
	ErrorTypeNotFound     ErrorType = "not_found"
	ErrorTypeConflict     ErrorType = "conflict"
	ErrorTypePermission   ErrorType = "permission"
	ErrorTypeStorage      ErrorType = "storage"
	ErrorTypeNetwork      ErrorType = "network"
	ErrorTypeTimeout      ErrorType = "timeout"
	ErrorTypeRateLimit    ErrorType = "rate_limit"
	ErrorTypeInternal     ErrorType = "internal"
	ErrorTypeUnavailable  ErrorType = "unavailable"
	ErrorTypeUnauthorized ErrorType = "unauthorized"
)

// KBError represents a structured error from kbVault operations
type KBError struct {
	Type    ErrorType              `json:"type"`
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Details string                 `json:"details,omitempty"`
	Context map[string]interface{} `json:"context,omitempty"`
	Cause   error                  `json:"-"`
}

func (e *KBError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("[%s:%s] %s: %s", e.Type, e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Code, e.Message)
}

func (e *KBError) Unwrap() error {
	return e.Cause
}

// WithContext adds contextual information to the error
func (e *KBError) WithContext(key string, value interface{}) *KBError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithCause sets the underlying cause of the error
func (e *KBError) WithCause(err error) *KBError {
	e.Cause = err
	return e
}

// IsType checks if the error is of a specific type
func (e *KBError) IsType(errType ErrorType) bool {
	return e.Type == errType
}

// IsRetryable returns true if the operation can be retried
func (e *KBError) IsRetryable() bool {
	return e.Type == ErrorTypeNetwork || e.Type == ErrorTypeTimeout || e.Type == ErrorTypeUnavailable
}

// NewKBError creates a new kbVault error
func NewKBError(errType ErrorType, code, message string) *KBError {
	return &KBError{
		Type:    errType,
		Code:    code,
		Message: message,
	}
}

// Validation Errors
func NewValidationError(message string) *KBError {
	return NewKBError(ErrorTypeValidation, "VALIDATION_FAILED", message)
}

func NewInvalidIDError(id string) *KBError {
	return NewKBError(ErrorTypeValidation, "INVALID_ID", "invalid note ID").
		WithContext("id", id)
}

func NewInvalidContentError(reason string) *KBError {
	return NewKBError(ErrorTypeValidation, "INVALID_CONTENT", "invalid note content").
		WithContext("reason", reason)
}

// Not Found Errors
func NewNoteNotFoundError(id string) *KBError {
	return NewKBError(ErrorTypeNotFound, "NOTE_NOT_FOUND", "note not found").
		WithContext("note_id", id)
}

func NewTemplateNotFoundError(name string) *KBError {
	return NewKBError(ErrorTypeNotFound, "TEMPLATE_NOT_FOUND", "template not found").
		WithContext("template", name)
}

func NewVaultNotFoundError(path string) *KBError {
	return NewKBError(ErrorTypeNotFound, "VAULT_NOT_FOUND", "vault not found").
		WithContext("path", path)
}

// Conflict Errors
func NewNoteExistsError(id string) *KBError {
	return NewKBError(ErrorTypeConflict, "NOTE_EXISTS", "note already exists").
		WithContext("note_id", id)
}

func NewConcurrencyConflictError(resource string) *KBError {
	return NewKBError(ErrorTypeConflict, "CONCURRENCY_CONFLICT", "concurrent modification detected").
		WithContext("resource", resource)
}

// Permission Errors
func NewPermissionDeniedError(operation, resource string) *KBError {
	return NewKBError(ErrorTypePermission, "PERMISSION_DENIED", "permission denied").
		WithContext("operation", operation).
		WithContext("resource", resource)
}

func NewReadOnlyError(resource string) *KBError {
	return NewKBError(ErrorTypePermission, "READ_ONLY", "resource is read-only").
		WithContext("resource", resource)
}

// Storage Errors
func NewStorageUnavailableError(backend string) *KBError {
	return NewKBError(ErrorTypeUnavailable, "STORAGE_UNAVAILABLE", "storage backend unavailable").
		WithContext("backend", backend)
}

func NewStorageConnectionError(backend string, err error) *KBError {
	return NewKBError(ErrorTypeNetwork, "STORAGE_CONNECTION_FAILED", "failed to connect to storage").
		WithContext("backend", backend).
		WithCause(err)
}

func NewStorageTimeoutError(backend, operation string) *KBError {
	return NewKBError(ErrorTypeTimeout, "STORAGE_TIMEOUT", "storage operation timed out").
		WithContext("backend", backend).
		WithContext("operation", operation)
}

// Authentication/Authorization Errors
func NewUnauthorizedError(message string) *KBError {
	return NewKBError(ErrorTypeUnauthorized, "UNAUTHORIZED", message)
}

func NewInvalidTokenError() *KBError {
	return NewKBError(ErrorTypeUnauthorized, "INVALID_TOKEN", "authentication token is invalid")
}

func NewTokenExpiredError() *KBError {
	return NewKBError(ErrorTypeUnauthorized, "TOKEN_EXPIRED", "authentication token has expired")
}

// Rate Limiting Errors
func NewRateLimitError(limit int, window string) *KBError {
	return NewKBError(ErrorTypeRateLimit, "RATE_LIMIT_EXCEEDED", "rate limit exceeded").
		WithContext("limit", limit).
		WithContext("window", window)
}

// Internal Errors
func NewInternalError(message string, err error) *KBError {
	return NewKBError(ErrorTypeInternal, "INTERNAL_ERROR", message).
		WithCause(err)
}

func NewConfigurationError(message string) *KBError {
	return NewKBError(ErrorTypeInternal, "CONFIGURATION_ERROR", message)
}

func NewCircuitBreakerError(service string) *KBError {
	return NewKBError(ErrorTypeUnavailable, "CIRCUIT_BREAKER_OPEN", "service temporarily unavailable").
		WithContext("service", service)
}

// Service Unavailable Errors
func NewServiceUnavailableError(service string) *KBError {
	return NewKBError(ErrorTypeUnavailable, "SERVICE_UNAVAILABLE", "service is temporarily unavailable").
		WithContext("service", service)
}

func NewMaintenanceModeError() *KBError {
	return NewKBError(ErrorTypeUnavailable, "MAINTENANCE_MODE", "system is in maintenance mode")
}

// Helper functions for error checking
func IsValidationError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsType(ErrorTypeValidation)
	}
	return false
}

func IsNotFoundError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsType(ErrorTypeNotFound)
	}
	return false
}

func IsConflictError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsType(ErrorTypeConflict)
	}
	return false
}

func IsRetryableError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsRetryable()
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsType(ErrorTypeUnauthorized)
	}
	return false
}

func IsRateLimitError(err error) bool {
	if kbErr, ok := err.(*KBError); ok {
		return kbErr.IsType(ErrorTypeRateLimit)
	}
	return false
}

// HTTPStatusCode returns the appropriate HTTP status code for the error
func (e *KBError) HTTPStatusCode() int {
	switch e.Type {
	case ErrorTypeValidation:
		return 400 // Bad Request
	case ErrorTypeNotFound:
		return 404 // Not Found
	case ErrorTypeConflict:
		return 409 // Conflict
	case ErrorTypePermission:
		return 403 // Forbidden
	case ErrorTypeUnauthorized:
		return 401 // Unauthorized
	case ErrorTypeRateLimit:
		return 429 // Too Many Requests
	case ErrorTypeTimeout:
		return 408 // Request Timeout
	case ErrorTypeUnavailable:
		return 503 // Service Unavailable
	case ErrorTypeInternal:
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}
