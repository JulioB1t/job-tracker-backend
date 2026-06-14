package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterAllowsLANClientPreflight(t *testing.T) {
	assertPreflightAllowed(t, "http://192.168.1.167:3000")
}

func TestRouterAllowsLocalhostPreflight(t *testing.T) {
	assertPreflightAllowed(t, "http://localhost:3000")
}

func assertPreflightAllowed(t *testing.T, origin string) {
	t.Helper()

	router := NewRouter(nil)

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/applications", nil)
	req.Header.Set("Origin", origin)
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)
	req.Header.Set("Access-Control-Request-Headers", "content-type")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Fatalf("expected Access-Control-Allow-Origin %q, got %q", origin, got)
	}

	if got := rec.Header().Get("Access-Control-Allow-Methods"); got != http.MethodPost {
		t.Fatalf("expected Access-Control-Allow-Methods %q, got %q", http.MethodPost, got)
	}
}
