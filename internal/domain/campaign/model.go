package campaign

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusDraft      Status = "draft"
	StatusPlanned    Status = "planned"
	StatusApproved   Status = "approved"
	StatusInProgress Status = "in_progress"
	StatusCompleted  Status = "completed"
	StatusCanceled   Status = "canceled"
)

type Campaign struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	FarmID      string    `json:"farm_id"`
	Name        string    `json:"name"`
	WindowID    string    `json:"window_id"`
	Priority    int       `json:"priority"`
	BudgetCents int64     `json:"budget_cents"`
	Status      Status    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (v Campaign) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "campaign.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "campaign.validate", "tenant is required")
	}
	if strings.TrimSpace(v.FarmID) == "" {
		return fault.New(fault.CodeInvalid, "campaign.validate", "FarmID is required")
	}
	if strings.TrimSpace(v.Name) == "" {
		return fault.New(fault.CodeInvalid, "campaign.validate", "Name is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "campaign.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "campaign.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "campaign.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusPlanned, StatusApproved, StatusInProgress, StatusCompleted, StatusCanceled:
		return true
	default:
		return false
	}
}

func (v Campaign) WithStatus(next Status, now time.Time) (Campaign, error) {
	if !next.Valid() {
		return Campaign{}, fault.New(fault.CodeInvalid, "campaign.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Campaign) Clone() Campaign { return v }
