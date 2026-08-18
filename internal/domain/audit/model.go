package audit

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusRecorded Status = "recorded"
	StatusExported Status = "exported"
)

type Event struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	Actor      string    `json:"actor"`
	Action     string    `json:"action"`
	ObjectType string    `json:"object_type"`
	ObjectID   string    `json:"object_id"`
	RequestID  string    `json:"request_id"`
	Detail     string    `json:"detail"`
	Status     Status    `json:"status"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (v Event) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Actor) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "Actor is required")
	}
	if strings.TrimSpace(v.Action) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "Action is required")
	}
	if strings.TrimSpace(v.ObjectType) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "ObjectType is required")
	}
	if strings.TrimSpace(v.ObjectID) == "" {
		return fault.New(fault.CodeInvalid, "audit.validate", "ObjectID is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "audit.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "audit.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "audit.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusRecorded, StatusExported:
		return true
	default:
		return false
	}
}

func (v Event) WithStatus(next Status, now time.Time) (Event, error) {
	if !next.Valid() {
		return Event{}, fault.New(fault.CodeInvalid, "audit.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Event) Clone() Event { return v }
