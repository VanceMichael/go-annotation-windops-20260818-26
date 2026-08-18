package turbine

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusCommissioning Status = "commissioning"
	StatusAvailable     Status = "available"
	StatusRestricted    Status = "restricted"
	StatusOffline       Status = "offline"
)

type Turbine struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenant_id"`
	FarmID        string    `json:"farm_id"`
	Code          string    `json:"code"`
	RatedKW       int       `json:"rated_kw"`
	LastServiceAt time.Time `json:"last_service_at"`
	Status        Status    `json:"status"`
	Version       int64     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (v Turbine) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "turbine.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "turbine.validate", "tenant is required")
	}
	if strings.TrimSpace(v.FarmID) == "" {
		return fault.New(fault.CodeInvalid, "turbine.validate", "FarmID is required")
	}
	if strings.TrimSpace(v.Code) == "" {
		return fault.New(fault.CodeInvalid, "turbine.validate", "Code is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "turbine.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "turbine.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "turbine.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusCommissioning, StatusAvailable, StatusRestricted, StatusOffline:
		return true
	default:
		return false
	}
}

func (v Turbine) WithStatus(next Status, now time.Time) (Turbine, error) {
	if !next.Valid() {
		return Turbine{}, fault.New(fault.CodeInvalid, "turbine.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Turbine) Clone() Turbine { return v }
