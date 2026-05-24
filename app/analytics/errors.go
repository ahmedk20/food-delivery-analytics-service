package analytics

import (
	"net/http"

	apperr "github.com/quickbite/analytics-service/lib/errors"
)

var (
	ErrInvalidDateRange = apperr.New("ANALYTICS_INVALID_DATE_RANGE", http.StatusBadRequest, "'from' must be before or equal to 'to'")
	ErrMissingFrom      = apperr.New("VALIDATION_ERROR", http.StatusBadRequest, "Query parameter 'from' is required (YYYY-MM-DD)")
	ErrMissingTo        = apperr.New("VALIDATION_ERROR", http.StatusBadRequest, "Query parameter 'to' is required (YYYY-MM-DD)")
	ErrInvalidFrom      = apperr.New("VALIDATION_ERROR", http.StatusBadRequest, "Query parameter 'from' must be a valid date (YYYY-MM-DD)")
	ErrInvalidTo        = apperr.New("VALIDATION_ERROR", http.StatusBadRequest, "Query parameter 'to' must be a valid date (YYYY-MM-DD)")
	ErrInvalidID        = apperr.New("VALIDATION_ERROR", http.StatusBadRequest, "Path parameter 'restaurantId' must be a positive integer")
)
