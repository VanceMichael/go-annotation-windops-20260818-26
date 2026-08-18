package crew

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusAvailable Status = "available"
	StatusAssigned  Status = "assigned"
	StatusResting   Status = "resting"
	StatusInactive  Status = "inactive"
)

type Member struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	HomePort     string    `json:"home_port"`
	MaxHours     int       `json:"max_hours"`
	LastShiftEnd time.Time `json:"last_shift_end"`
	Status       Status    `json:"status"`
	Version      int64     `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (v Member) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "crew.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "crew.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Name) == "" {
		return fault.New(fault.CodeInvalid, "crew.validate", "Name is required")
	}
	if strings.TrimSpace(v.HomePort) == "" {
		return fault.New(fault.CodeInvalid, "crew.validate", "HomePort is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "crew.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "crew.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "crew.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusAvailable, StatusAssigned, StatusResting, StatusInactive:
		return true
	default:
		return false
	}
}

func (v Member) WithStatus(next Status, now time.Time) (Member, error) {
	if !next.Valid() {
		return Member{}, fault.New(fault.CodeInvalid, "crew.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Member) Clone() Member { return v }
