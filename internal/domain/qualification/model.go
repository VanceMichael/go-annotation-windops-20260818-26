package qualification

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusValid     Status = "valid"
	StatusSuspended Status = "suspended"
	StatusExpired   Status = "expired"
)

type Qualification struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	MemberID   string    `json:"member_id"`
	Kind       string    `json:"kind"`
	ValidFrom  time.Time `json:"valid_from"`
	ValidUntil time.Time `json:"valid_until"`
	Status     Status    `json:"status"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (v Qualification) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "qualification.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "qualification.validate", "tenant is required")
	}
	if strings.TrimSpace(v.MemberID) == "" {
		return fault.New(fault.CodeInvalid, "qualification.validate", "MemberID is required")
	}
	if strings.TrimSpace(v.Kind) == "" {
		return fault.New(fault.CodeInvalid, "qualification.validate", "Kind is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "qualification.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "qualification.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "qualification.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusValid, StatusSuspended, StatusExpired:
		return true
	default:
		return false
	}
}

func (v Qualification) WithStatus(next Status, now time.Time) (Qualification, error) {
	if !next.Valid() {
		return Qualification{}, fault.New(fault.CodeInvalid, "qualification.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Qualification) Clone() Qualification { return v }
