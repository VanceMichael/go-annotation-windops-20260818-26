package idempotency

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

type Record struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Key          string    `json:"key"`
	PayloadHash  string    `json:"payload_hash"`
	ResponseCode int       `json:"response_code"`
	ResponseBody string    `json:"response_body"`
	Status       Status    `json:"status"`
	Version      int64     `json:"version"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (v Record) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "tenant is required")
	}
	if strings.TrimSpace(v.Method) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "Method is required")
	}
	if strings.TrimSpace(v.Path) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "Path is required")
	}
	if strings.TrimSpace(v.Key) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "Key is required")
	}
	if strings.TrimSpace(v.PayloadHash) == "" {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "PayloadHash is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "idempotency.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusProcessing, StatusCompleted, StatusFailed:
		return true
	default:
		return false
	}
}

func (v Record) WithStatus(next Status, now time.Time) (Record, error) {
	if !next.Valid() {
		return Record{}, fault.New(fault.CodeInvalid, "idempotency.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Record) Clone() Record { return v }
