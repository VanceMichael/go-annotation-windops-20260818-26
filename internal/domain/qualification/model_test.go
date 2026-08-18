package qualification

import (
	"testing"
	"time"
)

func validQualification(now time.Time) Qualification {
	return Qualification{
		ID:       "qualification-1",
		TenantID: "tenant-a",
		MemberID: "crew-1", Kind: "offshore-lifting", ValidFrom: now.AddDate(-1, 0, 0), ValidUntil: now.AddDate(1, 0, 0),
		Status:    StatusValid,
		Version:   1,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func TestQualificationValidationAcceptsCompleteValue(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validQualification(now)
	if err := value.Validate(); err != nil {
		t.Fatalf("complete value should validate: %v", err)
	}
}

func TestQualificationValidationRequiresIdentityTenantStatusAndTimestamps(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name   string
		mutate func(*Qualification)
	}{
		{name: "missing id", mutate: func(value *Qualification) { value.ID = "" }},
		{name: "missing tenant", mutate: func(value *Qualification) { value.TenantID = "" }},
		{name: "invalid status", mutate: func(value *Qualification) { value.Status = Status("unknown") }},
		{name: "invalid version", mutate: func(value *Qualification) { value.Version = 0 }},
		{name: "missing created timestamp", mutate: func(value *Qualification) { value.CreatedAt = time.Time{} }},
		{name: "missing updated timestamp", mutate: func(value *Qualification) { value.UpdatedAt = time.Time{} }},
	}
	for _, item := range cases {
		t.Run(item.name, func(t *testing.T) {
			value := validQualification(now)
			item.mutate(&value)
			if err := value.Validate(); err == nil {
				t.Fatal("invalid value unexpectedly passed validation")
			}
		})
	}
}

func TestQualificationTransitionUpdatesVersionAndTimestamp(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validQualification(now)
	nextAt := now.Add(15 * time.Minute)
	transitioned, err := value.WithStatus(StatusSuspended, nextAt)
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if transitioned.Status != StatusSuspended {
		t.Fatalf("unexpected status: %s", transitioned.Status)
	}
	if transitioned.Version != value.Version+1 {
		t.Fatalf("version did not increment: %d", transitioned.Version)
	}
	if !transitioned.UpdatedAt.Equal(nextAt) {
		t.Fatalf("updated timestamp mismatch: %s", transitioned.UpdatedAt)
	}
	if value.Status != StatusValid || value.Version != 1 {
		t.Fatalf("transition mutated original value: %+v", value)
	}
}

func TestQualificationTransitionRejectsUnknownStatus(t *testing.T) {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := validQualification(now)
	if _, err := value.WithStatus(Status("unknown"), now.Add(time.Minute)); err == nil {
		t.Fatal("unknown target status should be rejected")
	}
}
