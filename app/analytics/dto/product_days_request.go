package dto

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/quickbite/analytics-service/app/analytics"
)

type GetProductDaysRequest struct {
	RestaurantID int
	From         string
	To           string
}

func ParseGetProductDaysRequest(r *http.Request) (*GetProductDaysRequest, error) {
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

	return &GetProductDaysRequest{
		RestaurantID: id,
		From:         from,
		To:           to,
	}, nil
}
