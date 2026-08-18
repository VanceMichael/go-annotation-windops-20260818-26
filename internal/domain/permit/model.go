package permit

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusSubmitted Status = "submitted"
	StatusApproved  Status = "approved"
	StatusActivated Status = "activated"
	StatusClosed    Status = "closed"
	StatusRejected  Status = "rejected"
)

type Permit struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	CampaignID  string    `json:"campaign_id"`
	VesselID    string    `json:"vessel_id"`
	WindowID    string    `json:"window_id"`
	RequestedBy string    `json:"requested_by"`
	ExpiresAt   time.Time `json:"expires_at"`
	Status      Status    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (v Permit) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "tenant is required")
	}
	if strings.TrimSpace(v.CampaignID) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "CampaignID is required")
	}
	if strings.TrimSpace(v.VesselID) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "VesselID is required")
	}
	if strings.TrimSpace(v.WindowID) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "WindowID is required")
	}
	if strings.TrimSpace(v.RequestedBy) == "" {
		return fault.New(fault.CodeInvalid, "permit.validate", "RequestedBy is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "permit.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "permit.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "permit.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusSubmitted, StatusApproved, StatusActivated, StatusClosed, StatusRejected:
		return true
	default:
		return false
	}
}

func (v Permit) WithStatus(next Status, now time.Time) (Permit, error) {
	if !next.Valid() {
		return Permit{}, fault.New(fault.CodeInvalid, "permit.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Permit) Clone() Permit { return v }
