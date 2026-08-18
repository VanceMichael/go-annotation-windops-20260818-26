package campaign

import (
	"testing"
	"time"
)

func validCampaign(now time.Time) Campaign {
	return Campaign{
		ID:       "campaign-1",
		TenantID: "tenant-a",
		FarmID:   "farm-1", Name: "Gearbox inspection", WindowID: "window-1", Priority: 2, BudgetCents: 500000,
		Status:    StatusDraft,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestCampaignValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validCampaign(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestCampaignValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Campaign)
	}{
		{name: "missing id", mutate: func(value *Campaign) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Campaign) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Campaign) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Campaign) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Campaign) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Campaign) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validCampaign(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestCampaignTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validCampaign(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusPlanned, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusPlanned {
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

func TestCampaignTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validCampaign(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
