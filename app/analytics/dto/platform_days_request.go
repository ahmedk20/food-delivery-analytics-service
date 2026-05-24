package dto

import (
	"net/http"
	"time"

	"github.com/quickbite/analytics-service/app/analytics"
)

type GetPlatformDaysRequest struct {
	From string
	To   string
}

func ParseGetPlatformDaysRequest(r *http.Request) (*GetPlatformDaysRequest, error) {
	from := r.URL.Query().Get("from")
	if from == "" {
		return nil, analytics.ErrMissingFrom
	}
	if _, err := time.Parse("2006-01-02", from); err != nil {
		return nil, analytics.ErrInvalidFrom
	}

	to := r.URL.Query().Get("to")
	if to == "" {
		return nil, analytics.ErrMissingTo
	}
	if _, err := time.Parse("2006-01-02", to); err != nil {
		return nil, analytics.ErrInvalidTo
	}

	return &GetPlatformDaysRequest{From: from, To: to}, nil
}
