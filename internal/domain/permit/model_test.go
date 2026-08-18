package permit

import (
	"testing"
	"time"
)

func validPermit(now time.Time) Permit {
	return Permit{
		ID:         "permit-1",
		TenantID:   "tenant-a",
		CampaignID: "campaign-1", VesselID: "vessel-1", WindowID: "window-1", RequestedBy: "planner-1", ExpiresAt: now.Add(8 * time.Hour),
		Status:    StatusDraft,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestPermitValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validPermit(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestPermitValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Permit)
	}{
		{name: "missing id", mutate: func(value *Permit) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Permit) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Permit) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Permit) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Permit) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Permit) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validPermit(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestPermitTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validPermit(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusSubmitted, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusSubmitted {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusDraft || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestPermitTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validPermit(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
