package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

const (
	restaurantURL = "/api/v1/analytics/restaurants/9999/days?from=2024-01-01&to=2024-01-31"
	platformURL   = "/api/v1/analytics/platform/days?from=2024-01-01&to=2024-01-31"
)

// TestAuth_NoToken_Returns401 verifies that missing credentials return 401.
func TestAuth_NoToken_Returns401(t *testing.T) {
	resp, err := http.Get(apiServer.URL + restaurantURL)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized)
	assertErrorCode(t, resp, "UNAUTHENTICATED")
}

// TestAuth_InvalidToken_Returns401 verifies that a malformed JWT returns 401.
func TestAuth_InvalidToken_Returns401(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+restaurantURL, nil)
	req.Header.Set("Authorization", "Bearer not.a.valid.jwt")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusUnauthorized)
}

// TestAuth_NoPermission_Returns403 verifies that a valid JWT whose role has no
// analytics:read permission returns 403 Forbidden.
func TestAuth_NoPermission_Returns403(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+restaurantURL, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("staff")) // mock RBAC returns no perms
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusForbidden)
	assertErrorCode(t, resp, "FORBIDDEN")
}

// TestAuth_SystemAdmin_Returns200 verifies that system_admin bypasses RBAC.
func TestAuth_SystemAdmin_Returns200(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+restaurantURL, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("system_admin"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)
}

// TestAuth_ManagerWithPermission_Returns200 verifies that a role with
// analytics:read (as returned by the mock core service) gets through.
func TestAuth_ManagerWithPermission_Returns200(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+restaurantURL, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("manager")) // mock RBAC returns analytics:read
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)
}

// TestAuth_CookieToken_Returns200 verifies that the access_token cookie is
// accepted as an alternative to the Authorization header.
func TestAuth_CookieToken_Returns200(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+restaurantURL, nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: mintJWT("system_admin")})
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)
}

// TestAPI_MissingFromParam_Returns400 verifies validation of the 'from' query param.
func TestAPI_MissingFromParam_Returns400(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/analytics/restaurants/1/days?to=2024-01-31", apiServer.URL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("system_admin"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

// TestAPI_InvalidRestaurantID_Returns400 verifies that a non-numeric restaurant ID
// returns 400.
func TestAPI_InvalidRestaurantID_Returns400(t *testing.T) {
	url := fmt.Sprintf("%s/api/v1/analytics/restaurants/abc/days?from=2024-01-01&to=2024-01-31", apiServer.URL)
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("system_admin"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusBadRequest)
}

// TestAPI_PlatformDays_Returns200 verifies the platform days endpoint is reachable.
func TestAPI_PlatformDays_Returns200(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, apiServer.URL+platformURL, nil)
	req.Header.Set("Authorization", "Bearer "+mintJWT("system_admin"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	assertStatus(t, resp, http.StatusOK)
}

// assertStatus is a helper that fails the test if the response status does not match.
func assertStatus(t *testing.T, resp *http.Response, want int) {
	t.Helper()
	if resp.StatusCode != want {
		t.Errorf("status = %d, want %d", resp.StatusCode, want)
	}
}

// assertErrorCode reads the error response body and verifies the error code field.
func assertErrorCode(t *testing.T, resp *http.Response, wantCode string) {
	t.Helper()
	var body struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if body.Success {
		t.Error("success = true, want false for error response")
	}
	if body.Error.Code != wantCode {
		t.Errorf("error.code = %q, want %q", body.Error.Code, wantCode)
	}
}
