package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JulioB1t/job-tracker-backend/internal/store"
)

type applicationResponse struct {
	ID                     string `json:"id"`
	CompanyName            string `json:"companyName"`
	Title                  string `json:"title"`
	Source                 string `json:"source,omitempty"`
	CurrentStatus          string `json:"currentStatus"`
	ApplicationSubmittedAt string `json:"applicationSubmittedAt,omitempty"`
	CreatedAt              string `json:"createdAt"`
	UpdatedAt              string `json:"updatedAt"`
	Salary                 struct {
		Currency string `json:"currency"`
	} `json:"salary"`
}

func TestCreateApplicationAcceptsNullableFields(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))

	body := strings.NewReader(`{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"jobDescription": null,
		"salary": null,
		"applicationSubmittedAt": null
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	for _, field := range []string{"jobDescription", "salary", "applicationSubmittedAt"} {
		if _, ok := response[field]; ok {
			t.Fatalf("expected nil field %q to be omitted from response, got %v", field, response[field])
		}
	}
}

func TestCreateApplicationAcceptsDateOnlySubmittedAt(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))

	body := strings.NewReader(`{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"applicationSubmittedAt": "2026-06-17"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response struct {
		ApplicationSubmittedAt string `json:"applicationSubmittedAt"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.ApplicationSubmittedAt != "2026-06-17T00:00:00Z" {
		t.Fatalf("expected submitted date at UTC midnight, got %q", response.ApplicationSubmittedAt)
	}
}

func TestCreateApplicationAcceptsDateTimeSubmittedAt(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))

	body := strings.NewReader(`{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"applicationSubmittedAt": "2026-06-17T12:30:00Z"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
}

func TestCreateApplicationDefaultsSalaryCurrencyToUSD(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))

	body := strings.NewReader(`{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"salary": {
			"min": 90000,
			"max": 130000
		}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response struct {
		Salary struct {
			Currency string `json:"currency"`
		} `json:"salary"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Salary.Currency != "USD" {
		t.Fatalf("expected salary currency to default to USD, got %q", response.Salary.Currency)
	}
}

func TestUpdateApplicationReplacesApplication(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))
	created := createApplicationForTest(t, router, `{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"source": "LinkedIn",
		"currentStatus": "APPLIED"
	}`)

	body := strings.NewReader(`{
		"companyName": "Acme Updated",
		"title": "Senior Backend Engineer",
		"source": "Company Site",
		"currentStatus": "TECHNICAL_INTERVIEW",
		"applicationSubmittedAt": "2026-06-17",
		"salary": {
			"min": 90000
		}
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/"+created.ID, body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var updated applicationResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if updated.ID != created.ID {
		t.Fatalf("expected id %q to be preserved, got %q", created.ID, updated.ID)
	}
	if updated.CreatedAt != created.CreatedAt {
		t.Fatalf("expected createdAt %q to be preserved, got %q", created.CreatedAt, updated.CreatedAt)
	}
	if updated.CompanyName != "Acme Updated" || updated.Title != "Senior Backend Engineer" {
		t.Fatalf("expected updated application fields, got company=%q title=%q", updated.CompanyName, updated.Title)
	}
	if updated.CurrentStatus != "TECHNICAL_INTERVIEW" {
		t.Fatalf("expected updated status, got %q", updated.CurrentStatus)
	}
	if updated.ApplicationSubmittedAt != "2026-06-17T00:00:00Z" {
		t.Fatalf("expected submitted date at UTC midnight, got %q", updated.ApplicationSubmittedAt)
	}
	if updated.Salary.Currency != "USD" {
		t.Fatalf("expected salary currency to default to USD, got %q", updated.Salary.Currency)
	}
}

func TestUpdateApplicationDoesNotKeepOmittedClientFields(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))
	created := createApplicationForTest(t, router, `{
		"companyName": "Acme",
		"title": "Backend Engineer",
		"source": "LinkedIn",
		"currentStatus": "APPLIED"
	}`)

	body := strings.NewReader(`{
		"companyName": "Acme Updated",
		"title": "Senior Backend Engineer"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/"+created.ID, body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var updated applicationResponse
	if err := json.NewDecoder(rec.Body).Decode(&updated); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if updated.Source != "" {
		t.Fatalf("expected omitted source to be cleared, got %q", updated.Source)
	}
	if updated.CurrentStatus != "SAVED" {
		t.Fatalf("expected omitted status to default to SAVED, got %q", updated.CurrentStatus)
	}
}

func TestUpdateApplicationReturnsNotFound(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))

	body := strings.NewReader(`{
		"companyName": "Acme",
		"title": "Backend Engineer"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/missing", body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusNotFound, rec.Code, rec.Body.String())
	}
}

func TestUpdateApplicationRequiresTitle(t *testing.T) {
	router := NewRouter(NewHandler(store.NewMemoryApplicationStore()))
	created := createApplicationForTest(t, router, `{
		"companyName": "Acme",
		"title": "Backend Engineer"
	}`)

	body := strings.NewReader(`{
		"companyName": "Acme Updated"
	}`)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/applications/"+created.ID, body)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d with body %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func createApplicationForTest(t *testing.T, router http.Handler, body string) applicationResponse {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/api/v1/applications", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected create status %d, got %d with body %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var response applicationResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	return response
}
