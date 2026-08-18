package vessel

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusAvailable   Status = "available"
	StatusReserved    Status = "reserved"
	StatusUnderway    Status = "underway"
	StatusMaintenance Status = "maintenance"
)

type Vessel struct {
	ID              string    `json:"id"`
	TenantID        string    `json:"tenant_id"`
	Name            string    `json:"name"`
	HomePort        string    `json:"home_port"`
	SeatCapacity    int       `json:"seat_capacity"`
	CargoCapacityKG int       `json:"cargo_capacity_kg"`
	InspectionDue   time.Time `json:"inspection_due"`
	Status          Status    `json:"status"`
	Version         int64     `json:"version"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (v Vessel) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "vessel.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "vessel.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Name) == "" {
		return fault.New(fault.CodeInvalid, "vessel.validate", "Name is required")
	}
	if strings.TrimSpace(v.HomePort) == "" {
		return fault.New(fault.CodeInvalid, "vessel.validate", "HomePort is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "vessel.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "vessel.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "vessel.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusAvailable, StatusReserved, StatusUnderway, StatusMaintenance:
		return true
	default:
		return false
	}
}

func (v Vessel) WithStatus(next Status, now time.Time) (Vessel, error) {
	if !next.Valid() {
		return Vessel{}, fault.New(fault.CodeInvalid, "vessel.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Vessel) Clone() Vessel { return v }
