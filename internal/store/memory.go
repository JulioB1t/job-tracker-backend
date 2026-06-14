package store

import (
	"sync"
	"time"

	"github.com/JulioB1t/job-tracker-backend/internal/domain"
	"github.com/google/uuid"
)

type MemoryApplicationStore struct {
	mu          sync.RWMutex
	apps        map[string]domain.Application
	transitions map[string][]domain.StatusTransition
}

func NewMemoryApplicationStore() *MemoryApplicationStore {
	return &MemoryApplicationStore{
		apps:        make(map[string]domain.Application),
		transitions: make(map[string][]domain.StatusTransition),
	}
}

func (s *MemoryApplicationStore) Create(app domain.Application) domain.Application {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	app.ID = uuid.NewString()
	app.CreatedAt = now
	app.UpdatedAt = now

	if app.CurrentStatus == "" {
		app.CurrentStatus = domain.StatusSaved
	}

	normalizeApplication(&app)
	s.apps[app.ID] = app

	initialTransition := domain.StatusTransition{
		ID:            uuid.NewString(),
		ApplicationID: app.ID,
		FromStatus:    "",
		ToStatus:      app.CurrentStatus,
		CreatedAt:     now,
	}
	s.transitions[app.ID] = append(s.transitions[app.ID], initialTransition)

	return app
}

func (s *MemoryApplicationStore) List() []domain.Application {
	s.mu.RLock()
	defer s.mu.RUnlock()

	apps := make([]domain.Application, 0, len(s.apps))
	for _, app := range s.apps {
		apps = append(apps, app)
	}

	return apps
}

func (s *MemoryApplicationStore) Get(id string) (domain.Application, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	app, ok := s.apps[id]
	if !ok {
		return domain.Application{}, domain.ErrApplicationNotFound
	}

	return app, nil
}

func (s *MemoryApplicationStore) Update(id string, app domain.Application) (domain.Application, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.apps[id]
	if !ok {
		return domain.Application{}, domain.ErrApplicationNotFound
	}

	now := time.Now().UTC()
	app.ID = existing.ID
	app.CreatedAt = existing.CreatedAt
	app.UpdatedAt = now

	if app.CurrentStatus == "" {
		app.CurrentStatus = domain.StatusSaved
	}

	normalizeApplication(&app)
	s.apps[id] = app

	return app, nil
}

func (s *MemoryApplicationStore) UpdateStatus(id string, toStatus domain.ApplicationStatus, note string) (domain.StatusTransition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	app, ok := s.apps[id]
	if !ok {
		return domain.StatusTransition{}, domain.ErrApplicationNotFound
	}

	now := time.Now().UTC()
	transition := domain.StatusTransition{
		ID:            uuid.NewString(),
		ApplicationID: id,
		FromStatus:    app.CurrentStatus,
		ToStatus:      toStatus,
		Note:          note,
		CreatedAt:     now,
	}

	app.CurrentStatus = toStatus
	app.UpdatedAt = now

	s.apps[id] = app
	s.transitions[id] = append(s.transitions[id], transition)

	return transition, nil
}

func (s *MemoryApplicationStore) ListTransitions(applicationID string) ([]domain.StatusTransition, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.apps[applicationID]; !ok {
		return nil, domain.ErrApplicationNotFound
	}

	return s.transitions[applicationID], nil
}

func normalizeApplication(app *domain.Application) {
	if app.Salary != nil && app.Salary.Currency == "" {
		app.Salary.Currency = "USD"
	}
}
