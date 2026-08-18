package outbox

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusPending          Status = "pending"
	StatusRunning          Status = "running"
	StatusSucceeded        Status = "succeeded"
	StatusRetry            Status = "retry"
	StatusPermanentFailure Status = "permanent_failure"
)

type Job struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenant_id"`
	Topic       string    `json:"topic"`
	ObjectID    string    `json:"object_id"`
	Payload     string    `json:"payload"`
	Attempts    int       `json:"attempts"`
	AvailableAt time.Time `json:"available_at"`
	LastError   string    `json:"last_error"`
	Status      Status    `json:"status"`
	Version     int64     `json:"version"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (v Job) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "outbox.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "outbox.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Topic) == "" {
		return fault.New(fault.CodeInvalid, "outbox.validate", "Topic is required")
	}
	if strings.TrimSpace(v.ObjectID) == "" {
		return fault.New(fault.CodeInvalid, "outbox.validate", "ObjectID is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "outbox.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "outbox.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "outbox.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusRunning, StatusSucceeded, StatusRetry, StatusPermanentFailure:
		return true
	default:
		return false
	}
}

func (v Job) WithStatus(next Status, now time.Time) (Job, error) {
	if !next.Valid() {
		return Job{}, fault.New(fault.CodeInvalid, "outbox.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Job) Clone() Job { return v }
