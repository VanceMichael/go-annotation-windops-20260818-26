package workorder

import (
	"testing"
	"time"
)

func validWorkOrder(now time.Time) WorkOrder {
	return WorkOrder{
		ID:         "workorder-1",
		TenantID:   "tenant-a",
		CampaignID: "campaign-1", TurbineID: "turbine-1", PermitID: "permit-1", AssigneeID: "crew-1", Summary: "Inspect gearbox vibration", PlannedMinutes: 180,
		Status:    StatusQueued,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestWorkOrderValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWorkOrder(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestWorkOrderValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*WorkOrder)
	}{
		{name: "missing id", mutate: func(value *WorkOrder) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *WorkOrder) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *WorkOrder) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *WorkOrder) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *WorkOrder) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *WorkOrder) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validWorkOrder(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestWorkOrderTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWorkOrder(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusAssigned, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusAssigned {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusQueued || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestWorkOrderTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validWorkOrder(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
