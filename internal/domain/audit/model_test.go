package audit

import (
	"testing"
	"time"
)

func validEvent(now time.Time) Event {
	return Event{
		ID:       "audit-1",
		TenantID: "tenant-a",
		Actor:    "planner-1", Action: "campaign.create", ObjectType: "campaign", ObjectID: "campaign-1", RequestID: "request-1", Detail: "{}",
		Status:    StatusRecorded,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestEventValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvent(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestEventValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Event)
	}{
		{name: "missing id", mutate: func(value *Event) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Event) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Event) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Event) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Event) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Event) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validEvent(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestEventTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvent(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusExported, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusExported {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusRecorded || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestEventTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validEvent(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
