package slot

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusOpen     Status = "open"
	StatusHeld     Status = "held"
	StatusBooked   Status = "booked"
	StatusReleased Status = "released"
)

type Slot struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	FarmID    string    `json:"farm_id"`
	TurbineID string    `json:"turbine_id"`
	WindowID  string    `json:"window_id"`
	StartsAt  time.Time `json:"starts_at"`
	EndsAt    time.Time `json:"ends_at"`
	Capacity  int       `json:"capacity"`
	Used      int       `json:"used"`
	Status    Status    `json:"status"`
	Version   int64     `json:"version"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (v Slot) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "slot.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "slot.validate", "tenant is required")
	}
	if strings.TrimSpace(v.FarmID) == "" {
		return fault.New(fault.CodeInvalid, "slot.validate", "FarmID is required")
	}
	if strings.TrimSpace(v.TurbineID) == "" {
		return fault.New(fault.CodeInvalid, "slot.validate", "TurbineID is required")
	}
	if strings.TrimSpace(v.WindowID) == "" {
		return fault.New(fault.CodeInvalid, "slot.validate", "WindowID is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "slot.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "slot.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "slot.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusOpen, StatusHeld, StatusBooked, StatusReleased:
		return true
	default:
		return false
	}
}

func (v Slot) WithStatus(next Status, now time.Time) (Slot, error) {
	if !next.Valid() {
		return Slot{}, fault.New(fault.CodeInvalid, "slot.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Slot) Clone() Slot { return v }
