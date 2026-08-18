package vessel

import (
	"testing"
	"time"
)

func validVessel(now time.Time) Vessel {
	return Vessel{
		ID:       "vessel-1",
		TenantID: "tenant-a",
		Name:     "Hai Gong 7", HomePort: "Donggang", SeatCapacity: 18, CargoCapacityKG: 12000, InspectionDue: now.AddDate(1, 0, 0),
		Status:    StatusAvailable,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestVesselValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validVessel(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestVesselValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Vessel)
	}{
		{name: "missing id", mutate: func(value *Vessel) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Vessel) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Vessel) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Vessel) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Vessel) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Vessel) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validVessel(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestVesselTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validVessel(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusReserved, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusReserved {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusAvailable || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestVesselTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validVessel(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
