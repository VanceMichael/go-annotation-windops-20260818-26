package dispatch

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusProposed Status = "proposed"
	StatusReserved Status = "reserved"
	StatusDeparted Status = "departed"
	StatusArrived  Status = "arrived"
	StatusReleased Status = "released"
)

type Dispatch struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	CampaignID  string    `json:"campaign_id"`
	VesselID    string    `json:"vessel_id"`
	PermitID    string    `json:"permit_id"`
	CrewCount   int       `json:"crew_count"`
	CargoKG     int       `json:"cargo_kg"`
	DepartureAt time.Time `json:"departure_at"`
	Status      Status    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (v Dispatch) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "tenant is required")
	}
	if strings.TrimSpace(v.CampaignID) == "" {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "CampaignID is required")
	}
	if strings.TrimSpace(v.VesselID) == "" {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "VesselID is required")
	}
	if strings.TrimSpace(v.PermitID) == "" {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "PermitID is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "dispatch.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusProposed, StatusReserved, StatusDeparted, StatusArrived, StatusReleased:
		return true
	default:
		return false
	}
}

func (v Dispatch) WithStatus(next Status, now time.Time) (Dispatch, error) {
	if !next.Valid() {
		return Dispatch{}, fault.New(fault.CodeInvalid, "dispatch.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Dispatch) Clone() Dispatch { return v }
