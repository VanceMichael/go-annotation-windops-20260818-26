package workorder

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusQueued           Status = "queued"
	StatusAssigned         Status = "assigned"
	StatusStarted          Status = "started"
	StatusAwaitingEvidence Status = "awaiting_evidence"
	StatusCompleted        Status = "completed"
	StatusFailed           Status = "failed"
	StatusCanceled         Status = "canceled"
)

type WorkOrder struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	CampaignID     string    `json:"campaign_id"`
	TurbineID      string    `json:"turbine_id"`
	PermitID       string    `json:"permit_id"`
	AssigneeID     string    `json:"assignee_id"`
	Summary        string    `json:"summary"`
	PlannedMinutes int       `json:"planned_minutes"`
	Status         Status    `json:"status"`
	Version        int64     `json:"version"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func (v WorkOrder) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "workorder.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "workorder.validate", "tenant is required")
	}
	if strings.TrimSpace(v.CampaignID) == "" {
		return fault.New(fault.CodeInvalid, "workorder.validate", "CampaignID is required")
	}
	if strings.TrimSpace(v.TurbineID) == "" {
		return fault.New(fault.CodeInvalid, "workorder.validate", "TurbineID is required")
	}
	if strings.TrimSpace(v.Summary) == "" {
		return fault.New(fault.CodeInvalid, "workorder.validate", "Summary is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "workorder.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "workorder.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "workorder.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusQueued, StatusAssigned, StatusStarted, StatusAwaitingEvidence, StatusCompleted, StatusFailed, StatusCanceled:
		return true
	default:
		return false
	}
}

func (v WorkOrder) WithStatus(next Status, now time.Time) (WorkOrder, error) {
	if !next.Valid() {
		return WorkOrder{}, fault.New(fault.CodeInvalid, "workorder.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v WorkOrder) Clone() WorkOrder { return v }
