package farm

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusActive    Status = "active"
	StatusSuspended Status = "suspended"
)

type Farm struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Name       string    `json:"name"`
	Timezone   string    `json:"timezone"`
	CoastZone  string    `json:"coast_zone"`
	CapacityMW int       `json:"capacity_mw"`
	Status     Status    `json:"status"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (v Farm) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "farm.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "farm.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Name) == "" {
		return fault.New(fault.CodeInvalid, "farm.validate", "Name is required")
	}
	if strings.TrimSpace(v.Timezone) == "" {
		return fault.New(fault.CodeInvalid, "farm.validate", "Timezone is required")
	}
	if strings.TrimSpace(v.CoastZone) == "" {
		return fault.New(fault.CodeInvalid, "farm.validate", "CoastZone is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "farm.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "farm.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "farm.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusDraft, StatusActive, StatusSuspended:
		return true
	default:
		return false
	}
}

func (v Farm) WithStatus(next Status, now time.Time) (Farm, error) {
	if !next.Valid() {
		return Farm{}, fault.New(fault.CodeInvalid, "farm.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Farm) Clone() Farm { return v }
