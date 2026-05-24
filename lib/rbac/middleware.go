package rbac

import (
	"net/http"

	"github.com/quickbite/analytics-service/lib/appcontext"
	apperr "github.com/quickbite/analytics-service/lib/errors"
	apphttp "github.com/quickbite/analytics-service/lib/http"
)

func Require(pc *PermissionCache, perm string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := appcontext.GetClaims(r.Context())
			if claims == nil {
				apphttp.SendError(w, apperr.ErrUnauthenticated.StatusCode, apperr.ErrUnauthenticated.Code, apperr.ErrUnauthenticated.Message)
				return
			}

			if claims.Role == "system_admin" {
				next.ServeHTTP(w, r)
				return
			}

			roleName := claims.Role
			if claims.RestaurantRole != nil {
				roleName = *claims.RestaurantRole
			}

			permissions, err := pc.GetPermissions(r.Context(), roleName)
			if err != nil {
				apphttp.SendError(w, http.StatusInternalServerError, "RBAC_LOOKUP_FAILED", "Failed to verify permissions")
				return
			}

			if !pc.HasPermission(permissions, perm) {
				apphttp.SendError(w, apperr.ErrForbidden.StatusCode, apperr.ErrForbidden.Code, apperr.ErrForbidden.Message)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
