package weather

import (
	"testing"
	"time"
)

func validWindow(now time.Time) Window {
	return Window{
		ID:       "weather-1",
		TenantID: "tenant-a",
		FarmID:   "farm-1", StartsAt: now.Add(time.Hour), EndsAt: now.Add(5 * time.Hour), MaxWaveCM: 180, MaxWindKPH: 45, Confidence: 90,
		Status:    StatusForecast,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestWindowValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWindow(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestWindowValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Window)
	}{
		{name: "missing id", mutate: func(value *Window) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Window) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Window) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Window) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Window) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Window) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validWindow(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestWindowTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWindow(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusConfirmed, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusConfirmed {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusForecast || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestWindowTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWindow(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
