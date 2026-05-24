package auth

import (
	"net/http"
	"strings"

	"github.com/quickbite/analytics-service/lib/appcontext"
	apperr "github.com/quickbite/analytics-service/lib/errors"
	apphttp "github.com/quickbite/analytics-service/lib/http"
)

func Authenticate(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := extractToken(r)
			if token == "" {
				apphttp.SendError(w, apperr.ErrUnauthenticated.StatusCode, apperr.ErrUnauthenticated.Code, apperr.ErrUnauthenticated.Message)
				return
			}

			claims, err := VerifyToken(token, secret)
			if err != nil {
				apphttp.SendError(w, apperr.ErrUnauthenticated.StatusCode, apperr.ErrUnauthenticated.Code, apperr.ErrUnauthenticated.Message)
				return
			}

			ctx := appcontext.SetClaims(r.Context(), claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractToken(r *http.Request) string {
	if cookie, err := r.Cookie("access_token"); err == nil && cookie.Value != "" {
		return cookie.Value
	}

	if auth := r.Header.Get("Authorization"); strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}
