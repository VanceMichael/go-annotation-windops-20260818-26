package evidence

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusPending  Status = "pending"
	StatusAccepted Status = "accepted"
	StatusRejected Status = "rejected"
)

type Evidence struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	WorkOrderID string    `json:"work_order_id"`
	Kind        string    `json:"kind"`
	ObjectKey   string    `json:"object_key"`
	Checksum    string    `json:"checksum"`
	CapturedAt  time.Time `json:"captured_at"`
	Status      Status    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (v Evidence) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "tenant is required")
	}
	if strings.TrimSpace(v.WorkOrderID) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "WorkOrderID is required")
	}
	if strings.TrimSpace(v.Kind) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "Kind is required")
	}
	if strings.TrimSpace(v.ObjectKey) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "ObjectKey is required")
	}
	if strings.TrimSpace(v.Checksum) == "" {
		return fault.New(fault.CodeInvalid, "evidence.validate", "Checksum is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "evidence.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "evidence.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "evidence.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusAccepted, StatusRejected:
		return true
	default:
		return false
	}
}

func (v Evidence) WithStatus(next Status, now time.Time) (Evidence, error) {
	if !next.Valid() {
		return Evidence{}, fault.New(fault.CodeInvalid, "evidence.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Evidence) Clone() Evidence { return v }
