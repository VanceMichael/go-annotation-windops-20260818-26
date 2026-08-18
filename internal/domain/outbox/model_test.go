package outbox

import (
	"testing"
	"time"
)

func validJob(now time.Time) Job {
	return Job{
		ID:       "outbox-1",
		TenantID: "tenant-a",
		Topic:    "campaign.approved", ObjectID: "campaign-1", Payload: "{}", Attempts: 0, AvailableAt: now,
		Status:    StatusPending,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestJobValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validJob(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestJobValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Job)
	}{
		{name: "missing id", mutate: func(value *Job) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Job) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Job) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Job) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Job) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Job) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validJob(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestJobTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validJob(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusRunning, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusRunning {
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

func TestJobTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validJob(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
