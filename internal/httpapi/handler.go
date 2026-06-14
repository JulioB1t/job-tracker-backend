package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/JulioB1t/job-tracker-backend/internal/domain"
	"github.com/go-chi/chi/v5"
)

type ApplicationStore interface {
	Create(app domain.Application) domain.Application
	List() []domain.Application
	Get(id string) (domain.Application, error)
	UpdateStatus(id string, toStatus domain.ApplicationStatus, note string) (domain.StatusTransition, error)
	ListTransitions(applicationID string) ([]domain.StatusTransition, error)
}

type Handler struct {
	store ApplicationStore
}

func NewHandler(store ApplicationStore) *Handler {
	return &Handler{store: store}
}

type createStatusTransitionRequest struct {
	ToStatus domain.ApplicationStatus `json:"toStatus"`
	Note     string                   `json:"note"`
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
	var req domain.Application
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

	app := h.store.Create(req)
	log.Printf("application created id=%s company=%q title=%q status=%s", app.ID, app.CompanyName, app.Title, app.CurrentStatus)

	writeJSON(w, http.StatusCreated, app)
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
