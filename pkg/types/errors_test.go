package types

import (
	"errors"
	"testing"
)

func TestNewValidationError(t *testing.T) {
	err := NewInvalidIDError("test-id")
	
	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected error type %s, got %s", ErrorTypeValidation, err.Type)
	}
	
	if err.Code != "INVALID_ID" {
		t.Errorf("Expected error code INVALID_ID, got %s", err.Code)
	}
	
	if err.Message != "invalid note ID" {
		t.Errorf("Expected message 'invalid note ID', got %q", err.Message)
	}
}

func TestNewContentError(t *testing.T) {
	err := NewInvalidContentError("content too large")
	
	if err.Type != ErrorTypeValidation {
		t.Errorf("Expected error type %s, got %s", ErrorTypeValidation, err.Type)
	}
	
	if err.Code != "INVALID_CONTENT" {
		t.Errorf("Expected error code INVALID_CONTENT, got %s", err.Code)
	}
}

func TestNewNotFoundErrors(t *testing.T) {
	tests := []struct {
		name     string
		createFn func() *KBError
		wantCode string
	}{
		{
			name: "note_not_found",
			createFn: func() *KBError {
				return NewNoteNotFoundError("test-id")
			},
			wantCode: "NOTE_NOT_FOUND",
		},
		{
			name: "template_not_found",
			createFn: func() *KBError {
				return NewTemplateNotFoundError("test-template")
			},
			wantCode: "TEMPLATE_NOT_FOUND",
		},
		{
			name: "vault_not_found",
			createFn: func() *KBError {
				return NewVaultNotFoundError("/test/path")
			},
			wantCode: "VAULT_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFn()
			
			if err.Type != ErrorTypeNotFound {
				t.Errorf("Expected error type %s, got %s", ErrorTypeNotFound, err.Type)
			}
			
			if err.Code != tt.wantCode {
				t.Errorf("Expected error code %s, got %s", tt.wantCode, err.Code)
			}
		})
	}
}

func TestNewConflictErrors(t *testing.T) {
	tests := []struct {
		name     string
		createFn func() *KBError
		wantCode string
	}{
		{
			name: "note_exists",
			createFn: func() *KBError {
				return NewNoteExistsError("test-id")
			},
			wantCode: "NOTE_EXISTS",
		},
		{
			name: "concurrency_conflict",
			createFn: func() *KBError {
				return NewConcurrencyConflictError("resource")
			},
			wantCode: "CONCURRENCY_CONFLICT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.createFn()
			
			if err.Type != ErrorTypeConflict {
				t.Errorf("Expected error type %s, got %s", ErrorTypeConflict, err.Type)
			}
			
			if err.Code != tt.wantCode {
				t.Errorf("Expected error code %s, got %s", tt.wantCode, err.Code)
			}
		})
	}
}

func TestKBError_Error(t *testing.T) {
	err := &KBError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "test error message",
	}
	
	expected := "[validation:TEST_ERROR] test error message"
	if err.Error() != expected {
		t.Errorf("Expected error string %q, got %q", expected, err.Error())
	}
}

func TestKBError_ErrorWithDetails(t *testing.T) {
	err := &KBError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "test error message",
		Details: "additional details",
	}
	
	expected := "[validation:TEST_ERROR] test error message: additional details"
	if err.Error() != expected {
		t.Errorf("Expected error string %q, got %q", expected, err.Error())
	}
}

func TestKBError_Unwrap(t *testing.T) {
	cause := errors.New("underlying error")
	err := &KBError{
		Type:    ErrorTypeInternal,
		Code:    "TEST_ERROR",
		Message: "test error",
		Cause:   cause,
	}
	
	if err.Unwrap() != cause {
		t.Error("Unwrap should return the underlying cause")
	}
}

func TestKBError_WithContext(t *testing.T) {
	err := &KBError{
		Type:    ErrorTypeValidation,
		Code:    "TEST_ERROR",
		Message: "test error",
	}
	
	newErr := err.WithContext("field", "test_field")
	
	if newErr.Context == nil {
		t.Error("Context should be set")
	}
	
	if newErr.Context["field"] != "test_field" {
		t.Error("Context field not set correctly")
	}
}

func TestKBError_WithCause(t *testing.T) {
	err := &KBError{
		Type:    ErrorTypeInternal,
		Code:    "TEST_ERROR",
		Message: "test error",
	}
	
	cause := errors.New("underlying cause")
	newErr := err.WithCause(cause)
	
	if newErr.Cause != cause {
		t.Error("Cause should be set")
	}
}

func TestKBError_IsType(t *testing.T) {
	err := &KBError{
		Type: ErrorTypeValidation,
		Code: "TEST_ERROR",
	}
	
	if !err.IsType(ErrorTypeValidation) {
		t.Error("IsType should return true for matching type")
	}
	
	if err.IsType(ErrorTypeNotFound) {
		t.Error("IsType should return false for non-matching type")
	}
}

func TestErrorTypeCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		checker  func(error) bool
		expected bool
	}{
		{
			name:     "validation_error",
			err:      NewInvalidIDError("test"),
			checker:  IsValidationError,
			expected: true,
		},
		{
			name:     "not_found_error",
			err:      NewNoteNotFoundError("test"),
			checker:  IsNotFoundError,
			expected: true,
		},
		{
			name:     "conflict_error",
			err:      NewNoteExistsError("test"),
			checker:  IsConflictError,
			expected: true,
		},
		{
			name:     "regular_error",
			err:      errors.New("regular error"),
			checker:  IsValidationError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.checker(tt.err)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

