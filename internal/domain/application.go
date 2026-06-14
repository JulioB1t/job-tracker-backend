package domain

import (
	"errors"
	"time"
)

var ErrApplicationNotFound = errors.New("application not found")

type ApplicationStatus string

const (
	StatusSaved              ApplicationStatus = "SAVED"
	StatusApplied            ApplicationStatus = "APPLIED"
	StatusRecruiterScreen    ApplicationStatus = "RECRUITER_SCREEN"
	StatusOnlineAssessment   ApplicationStatus = "ONLINE_ASSESSMENT"
	StatusTechnicalInterview ApplicationStatus = "TECHNICAL_INTERVIEW"
	StatusFinalRound         ApplicationStatus = "FINAL_ROUND"
	StatusOffer              ApplicationStatus = "OFFER"
	StatusRejected           ApplicationStatus = "REJECTED"
	StatusWithdrawn          ApplicationStatus = "WITHDRAWN"
	StatusNoResponse         ApplicationStatus = "NO_RESPONSE"
)

type Salary struct {
	Min      *int   `json:"min,omitempty"`
	Max      *int   `json:"max,omitempty"`
	Currency string `json:"currency,omitempty"`
}

type JobDescription struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type Application struct {
	ID                     string            `json:"id"`
	CompanyName            string            `json:"companyName"`
	Title                  string            `json:"title"`
	Source                 string            `json:"source,omitempty"`
	JobURL                 string            `json:"jobUrl,omitempty"`
	Location               string            `json:"location,omitempty"`
	JobDescription         *JobDescription   `json:"jobDescription,omitempty"`
	Salary                 *Salary           `json:"salary,omitempty"`
	Sponsorship            string            `json:"sponsorship,omitempty"`
	CurrentStatus          ApplicationStatus `json:"currentStatus"`
	ApplicationSubmittedAt *time.Time        `json:"applicationSubmittedAt,omitempty"`
	Notes                  string            `json:"notes,omitempty"`
	CreatedAt              time.Time         `json:"createdAt"`
	UpdatedAt              time.Time         `json:"updatedAt"`
}

type StatusTransition struct {
	ID            string            `json:"id"`
	ApplicationID string            `json:"applicationId"`
	FromStatus    ApplicationStatus `json:"fromStatus,omitempty"`
	ToStatus      ApplicationStatus `json:"toStatus"`
	Note          string            `json:"note,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
}
