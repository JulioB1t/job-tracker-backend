package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestMaskRequestBodyMasksPIIAndKeepsJobDescription(t *testing.T) {
	body := []byte(`{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"location": "New York",
		"notes": "Call Jane at jane@example.com",
		"jobDescription": {
			"contentType": "text/plain",
			"content": "Email jobs@example.com with questions."
		}
	}`)

	masked := maskRequestBody(body)

	for _, value := range []string{"New York", "Call Jane"} {
		if strings.Contains(masked, value) {
			t.Fatalf("expected %q to be masked in %s", value, masked)
		}
	}

	for _, value := range []string{"Acme", "Backend Engineer", "Email jobs@example.com with questions."} {
		if !strings.Contains(masked, value) {
			t.Fatalf("expected %q to remain visible in %s", value, masked)
		}
	}
}

func TestLogRequestBodyRestoresBodyForHandler(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			t.Fatal("expected request body to be restored")
		}
		w.WriteHeader(http.StatusCreated)
	})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", strings.NewReader(`{"notes":"private"}`))
	rec := httptest.NewRecorder()

	logRequestBody(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}
