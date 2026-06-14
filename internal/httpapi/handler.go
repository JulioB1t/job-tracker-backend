package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/JulioB1t/job-tracker-backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

type ApplicationStore interface {
	Create(app domain.Application) domain.Application
	List() []domain.Application
	Get(id string) (domain.Application, error)
	Update(id string, app domain.Application) (domain.Application, error)
	UpdateStatus(id string, toStatus domain.ApplicationStatus, note string) (domain.StatusTransition, error)
	ListTransitions(applicationID string) ([]domain.StatusTransition, error)
}

type Handler struct {
	store ApplicationStore
}

func NewHandler(store ApplicationStore) *Handler {
	return &Handler{store: store}
}

type createApplicationRequest struct {
	CompanyName            string                   `json:"companyName"`
	Title                  string                   `json:"title"`
	Source                 string                   `json:"source,omitempty"`
	JobURL                 string                   `json:"jobUrl,omitempty"`
	Location               string                   `json:"location,omitempty"`
	JobDescription         *domain.JobDescription   `json:"jobDescription,omitempty"`
	Salary                 *domain.Salary           `json:"salary,omitempty"`
	Sponsorship            string                   `json:"sponsorship,omitempty"`
	CurrentStatus          domain.ApplicationStatus `json:"currentStatus"`
	ApplicationSubmittedAt *flexibleDateTime        `json:"applicationSubmittedAt,omitempty"`
	Notes                  string                   `json:"notes,omitempty"`
}

type createStatusTransitionRequest struct {
	ToStatus domain.ApplicationStatus `json:"toStatus"`
	Note     string                   `json:"note"`
}

type flexibleDateTime struct {
	value time.Time
}

func (dt *flexibleDateTime) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	value, err := strconv.Unquote(string(data))
	if err != nil {
		return err
	}
	if value == "" {
		return nil
	}

	for _, layout := range []string{time.RFC3339, "2006-01-02"} {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			dt.value = parsed.UTC()
			return nil
		}
	}

	return fmt.Errorf("expected date-time in RFC3339 or date in YYYY-MM-DD format")
}

func (dt *flexibleDateTime) timePtr() *time.Time {
	if dt == nil || dt.value.IsZero() {
		return nil
	}

	value := dt.value
	return &value
}

func (req createApplicationRequest) toApplication() domain.Application {
	return domain.Application{
		CompanyName:            req.CompanyName,
		Title:                  req.Title,
		Source:                 req.Source,
		JobURL:                 req.JobURL,
		Location:               req.Location,
		JobDescription:         req.JobDescription,
		Salary:                 req.Salary,
		Sponsorship:            req.Sponsorship,
		CurrentStatus:          req.CurrentStatus,
		ApplicationSubmittedAt: req.ApplicationSubmittedAt.timePtr(),
		Notes:                  req.Notes,
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("write JSON response failed: status=%d error=%v", status, err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func (h *Handler) createApplication(w http.ResponseWriter, r *http.Request) {
	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("create application failed: invalid JSON: %v", err)
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	if req.CompanyName == "" {
		log.Print("create application failed: missing companyName")
		writeError(w, http.StatusBadRequest, "companyName is required")
		return
	}

	if req.Title == "" {
		log.Printf("create application failed: missing title company=%q", req.CompanyName)
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	app := h.store.Create(req.toApplication())
	log.Printf("application created id=%s company=%q title=%q status=%s", app.ID, app.CompanyName, app.Title, app.CurrentStatus)

	writeJSON(w, http.StatusCreated, app)
}

func (h *Handler) updateApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req createApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("update application failed: invalid JSON id=%s error=%v", id, err)
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	if req.CompanyName == "" {
		log.Printf("update application failed: missing companyName id=%s", id)
		writeError(w, http.StatusBadRequest, "companyName is required")
		return
	}

	if req.Title == "" {
		log.Printf("update application failed: missing title id=%s company=%q", id, req.CompanyName)
		writeError(w, http.StatusBadRequest, "title is required")
		return
	}

	app, err := h.store.Update(id, req.toApplication())
	if errors.Is(err, domain.ErrApplicationNotFound) {
		log.Printf("update application failed: not found id=%s", id)
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	log.Printf("application updated id=%s company=%q title=%q status=%s", app.ID, app.CompanyName, app.Title, app.CurrentStatus)
	writeJSON(w, http.StatusOK, app)
}

func (h *Handler) listApplications(w http.ResponseWriter, r *http.Request) {
	apps := h.store.List()
	log.Printf("applications listed count=%d", len(apps))

	writeJSON(w, http.StatusOK, map[string]any{
		"applications": apps,
		"total":        len(apps),
	})
}

func (h *Handler) getApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	app, err := h.store.Get(id)
	if errors.Is(err, domain.ErrApplicationNotFound) {
		log.Printf("get application failed: not found id=%s", id)
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	log.Printf("application fetched id=%s status=%s", app.ID, app.CurrentStatus)
	writeJSON(w, http.StatusOK, app)
}

func (h *Handler) createStatusTransition(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req createStatusTransitionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("create status transition failed: invalid JSON applicationID=%s error=%v", id, err)
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	if req.ToStatus == "" {
		log.Printf("create status transition failed: missing toStatus applicationID=%s", id)
		writeError(w, http.StatusBadRequest, "toStatus is required")
		return
	}

	transition, err := h.store.UpdateStatus(id, req.ToStatus, req.Note)
	if errors.Is(err, domain.ErrApplicationNotFound) {
		log.Printf("create status transition failed: application not found id=%s toStatus=%s", id, req.ToStatus)
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	log.Printf(
		"status transition created id=%s applicationID=%s from=%s to=%s",
		transition.ID,
		transition.ApplicationID,
		transition.FromStatus,
		transition.ToStatus,
	)

	writeJSON(w, http.StatusCreated, transition)
}

func (h *Handler) listStatusTransitions(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	transitions, err := h.store.ListTransitions(id)
	if errors.Is(err, domain.ErrApplicationNotFound) {
		log.Printf("list status transitions failed: application not found id=%s", id)
		writeError(w, http.StatusNotFound, "application not found")
		return
	}

	log.Printf("status transitions listed applicationID=%s count=%d", id, len(transitions))
	writeJSON(w, http.StatusOK, map[string]any{
		"items": transitions,
		"total": len(transitions),
	})
}
