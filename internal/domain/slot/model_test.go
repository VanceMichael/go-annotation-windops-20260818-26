package slot

import (
	"testing"
	"time"
)

func validSlot(now time.Time) Slot {
	return Slot{
		ID:       "slot-1",
		TenantID: "tenant-a",
		FarmID:   "farm-1", TurbineID: "turbine-1", WindowID: "window-1", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Capacity: 2, Used: 0,
		Status:    StatusOpen,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestSlotValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validSlot(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestSlotValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Slot)
	}{
		{name: "missing id", mutate: func(value *Slot) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Slot) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Slot) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Slot) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Slot) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Slot) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validSlot(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestSlotTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validSlot(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusHeld, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusHeld {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusOpen || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestSlotTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validSlot(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
