package turbine

import (
	"testing"
	"time"
)

func validTurbine(now time.Time) Turbine {
	return Turbine{
		ID:       "turbine-1",
		TenantID: "tenant-a",
		FarmID:   "farm-1", Code: "NS-A01", RatedKW: 8000, LastServiceAt: now.AddDate(0, -2, 0),
		Status:    StatusCommissioning,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestTurbineValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validTurbine(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestTurbineValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Turbine)
	}{
		{name: "missing id", mutate: func(value *Turbine) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Turbine) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Turbine) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Turbine) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Turbine) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Turbine) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validTurbine(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestTurbineTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validTurbine(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusAvailable, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusAvailable {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusCommissioning || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestTurbineTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validTurbine(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
