package evidence

import (
	"testing"
	"time"
)

func validEvidence(now time.Time) Evidence {
	return Evidence{
		ID:          "evidence-1",
		TenantID:    "tenant-a",
		WorkOrderID: "work-1", Kind: "photo", ObjectKey: "evidence/photo-1.jpg", Checksum: "0123456789abcdef0123456789abcdef", CapturedAt: now,
		Status:    StatusPending,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestEvidenceValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvidence(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestEvidenceValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Evidence)
	}{
		{name: "missing id", mutate: func(value *Evidence) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Evidence) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Evidence) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Evidence) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Evidence) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Evidence) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validEvidence(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestEvidenceTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvidence(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusAccepted, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusAccepted {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusPending || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestEvidenceTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvidence(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
