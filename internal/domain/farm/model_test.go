package farm

import (
	"testing"
	"time"
)

func validFarm(now time.Time) Farm {
	return Farm{
		ID:       "farm-1",
		TenantID: "tenant-a",
		Name:     "North Shoal", Timezone: "Asia/Shanghai", CoastZone: "north", CapacityMW: 300,
		Status:    StatusDraft,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestFarmValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validFarm(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestFarmValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Farm)
	}{
		{name: "missing id", mutate: func(value *Farm) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Farm) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Farm) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Farm) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Farm) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Farm) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validFarm(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestFarmTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validFarm(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusActive, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusActive {
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

func TestFarmTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validFarm(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
