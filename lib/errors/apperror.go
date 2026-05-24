package errors

import "net/http"

type AppError struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	StatusCode int    `json:"-"`
}

func (e *AppError) Error() string {
	return e.Message
}

func New(code string, statusCode int, message string) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode}
}

var (
	ErrUnauthenticated = New("UNAUTHENTICATED", http.StatusUnauthorized, "Authentication required")
	ErrForbidden       = New("FORBIDDEN", http.StatusForbidden, "Insufficient permissions")
	ErrValidation      = New("VALIDATION_ERROR", http.StatusBadRequest, "Validation failed")
	ErrInternal        = New("INTERNAL_ERROR", http.StatusInternalServerError, "Something went wrong")
)
