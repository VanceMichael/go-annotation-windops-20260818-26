package idempotency

import (
	"testing"
	"time"
)

func validRecord(now time.Time) Record {
	return Record{
		ID:       "idempotency-1",
		TenantID: "tenant-a",
		Method:   "POST", Path: "/api/campaigns", Key: "key-1", PayloadHash: "0123456789abcdef", ResponseCode: 201, ResponseBody: "{}",
		Status:    StatusProcessing,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestRecordValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validRecord(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestRecordValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Record)
	}{
		{name: "missing id", mutate: func(value *Record) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Record) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Record) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Record) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Record) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Record) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validRecord(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestRecordTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validRecord(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusCompleted, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusCompleted {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusProcessing || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestRecordTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validRecord(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
