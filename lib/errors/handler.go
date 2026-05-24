package errors

import (
	"errors"
	"log/slog"
	"net/http"

	apphttp "github.com/quickbite/analytics-service/lib/http"
)

func Wrap(log *slog.Logger, h func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			var appErr *AppError
			if errors.As(err, &appErr) {
				log.Warn("app error",
					"code", appErr.Code,
					"status", appErr.StatusCode,
					"message", appErr.Message,
					"path", r.URL.Path,
				)
				apphttp.SendError(w, appErr.StatusCode, appErr.Code, appErr.Message)
				return
			}

			log.Error("unhandled error",
				"error", err.Error(),
				"path", r.URL.Path,
			)
			apphttp.SendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Something went wrong")
		}
	}
}
