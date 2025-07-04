package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name: "error without wrapped error",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
			},
			want: "TEST_ERROR: Test error message",
		},
		{
			name: "error with wrapped error",
			appError: &AppError{
				Code:    "TEST_ERROR",
				Message: "Test error message",
				Err:     errors.New("wrapped error"),
			},
			want: "TEST_ERROR: Test error message (wrapped error)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.appError.Error(); got != tt.want {
				t.Errorf("AppError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("wrapped error")
	appErr := &AppError{
		Code:    "TEST_ERROR",
		Message: "Test error message",
		Err:     wrappedErr,
	}

	if got := appErr.Unwrap(); got != wrappedErr {
		t.Errorf("AppError.Unwrap() = %v, want %v", got, wrappedErr)
	}
}

func TestNewAppError(t *testing.T) {
	code := "TEST_ERROR"
	message := "Test error message"
	statusCode := http.StatusBadRequest

	appErr := NewAppError(code, message, statusCode)

	if appErr.Code != code {
		t.Errorf("NewAppError().Code = %v, want %v", appErr.Code, code)
	}
	if appErr.Message != message {
		t.Errorf("NewAppError().Message = %v, want %v", appErr.Message, message)
	}
	if appErr.StatusCode != statusCode {
		t.Errorf("NewAppError().StatusCode = %v, want %v", appErr.StatusCode, statusCode)
	}
	if appErr.Err != nil {
		t.Errorf("NewAppError().Err = %v, want nil", appErr.Err)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	code := "WRAPPED_ERROR"
	message := "Wrapped error message"
	statusCode := http.StatusInternalServerError

	appErr := WrapError(originalErr, code, message, statusCode)

	if appErr.Code != code {
		t.Errorf("WrapError().Code = %v, want %v", appErr.Code, code)
	}
	if appErr.Message != message {
		t.Errorf("WrapError().Message = %v, want %v", appErr.Message, message)
	}
	if appErr.StatusCode != statusCode {
		t.Errorf("WrapError().StatusCode = %v, want %v", appErr.StatusCode, statusCode)
	}
	if appErr.Err != originalErr {
		t.Errorf("WrapError().Err = %v, want %v", appErr.Err, originalErr)
	}
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := NewAppError("TEST_ERROR", "Test error message", http.StatusBadRequest)
	details := "Additional error details"

	result := appErr.WithDetails(details)

	if result.Details != details {
		t.Errorf("AppError.WithDetails().Details = %v, want %v", result.Details, details)
	}
	// Should return the same instance
	if result != appErr {
		t.Errorf("AppError.WithDetails() should return the same instance")
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		name       string
		err        *AppError
		wantCode   string
		wantStatus int
	}{
		{
			name:       "ErrValidationFailed",
			err:        ErrValidationFailed,
			wantCode:   "VALIDATION_FAILED",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrBrokerNotFound",
			err:        ErrBrokerNotFound,
			wantCode:   "BROKER_NOT_FOUND",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrBrokerOffline",
			err:        ErrBrokerOffline,
			wantCode:   "BROKER_OFFLINE",
			wantStatus: http.StatusServiceUnavailable,
		},
		{
			name:       "ErrInvalidSymbol",
			err:        ErrInvalidSymbol,
			wantCode:   "INVALID_SYMBOL",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ErrDataNotFound",
			err:        ErrDataNotFound,
			wantCode:   "DATA_NOT_FOUND",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "ErrStorageError",
			err:        ErrStorageError,
			wantCode:   "STORAGE_ERROR",
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.wantCode {
				t.Errorf("%s.Code = %v, want %v", tt.name, tt.err.Code, tt.wantCode)
			}
			if tt.err.StatusCode != tt.wantStatus {
				t.Errorf("%s.StatusCode = %v, want %v", tt.name, tt.err.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestErrorsAs(t *testing.T) {
	originalErr := errors.New("original error")
	appErr := WrapError(originalErr, "TEST_ERROR", "Test message", http.StatusBadRequest)

	var targetAppErr *AppError
	if !errors.As(appErr, &targetAppErr) {
		t.Error("errors.As should work with AppError")
	}

	if targetAppErr.Code != "TEST_ERROR" {
		t.Errorf("errors.As result Code = %v, want TEST_ERROR", targetAppErr.Code)
	}
}
