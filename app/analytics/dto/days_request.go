package dto

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/quickbite/analytics-service/app/analytics"
)

type GetRestaurantDaysRequest struct {
	RestaurantID int
	From         string
	To           string
}

type GetBranchDaysRequest struct {
	RestaurantID int // from URL: {restaurantId}
	BranchID     int // from URL: {branchId}
	From         string
	To           string
}

func ParseGetBranchDaysRequest(r *http.Request) (*GetBranchDaysRequest, error) {
	// parse restaurantId (same as existing pattern)
	idStr := chi.URLParam(r, "restaurantId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return nil, analytics.ErrInvalidID
	}
	// parse branchId     (same pattern, different chi.URLParam key)
	branchIdStr := chi.URLParam(r, "branchId")
	branchId, err := strconv.Atoi(branchIdStr)
	if err != nil || branchId <= 0 {
		return nil, analytics.ErrInvalidID
	}
	// parse from / to    (identical to existing)
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
	return &GetBranchDaysRequest{
		RestaurantID: id,
		BranchID:     branchId,
		From:         from,
		To:           to,
	}, nil
}

func ParseGetRestaurantDaysRequest(r *http.Request) (*GetRestaurantDaysRequest, error) {
	idStr := chi.URLParam(r, "restaurantId")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		return nil, analytics.ErrInvalidID
	}

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

	return &GetRestaurantDaysRequest{
		RestaurantID: id,
		From:         from,
		To:           to,
	}, nil
}
