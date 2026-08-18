package dispatch

import (
	"testing"
	"time"
)

func validDispatch(now time.Time) Dispatch {
	return Dispatch{
		ID:         "dispatch-1",
		TenantID:   "tenant-a",
		CampaignID: "campaign-1", VesselID: "vessel-1", PermitID: "permit-1", CrewCount: 8, CargoKG: 3000, DepartureAt: now.Add(time.Hour),
		Status:    StatusProposed,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestDispatchValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validDispatch(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestDispatchValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Dispatch)
	}{
		{name: "missing id", mutate: func(value *Dispatch) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Dispatch) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Dispatch) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Dispatch) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Dispatch) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Dispatch) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validDispatch(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestDispatchTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validDispatch(now)
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
	if value.Status != StatusProposed || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestDispatchTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validDispatch(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
