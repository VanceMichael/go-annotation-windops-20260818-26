package crew

import (
	"testing"
	"time"
)

func validMember(now time.Time) Member {
	return Member{
		ID:       "crew-1",
		TenantID: "tenant-a",
		Name:     "Lin Yue", HomePort: "Donggang", MaxHours: 12, LastShiftEnd: now.Add(-12 * time.Hour),
		Status:    StatusAvailable,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestMemberValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validMember(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestMemberValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Member)
	}{
		{name: "missing id", mutate: func(value *Member) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Member) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Member) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Member) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Member) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Member) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validMember(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestMemberTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validMember(now)
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
	if value.Status != StatusAvailable || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestMemberTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validMember(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
