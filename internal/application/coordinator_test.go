package application

import (
	"context"
	"path/filepath"
	"testing"
	"time"
	"windops/internal/platform/clock"
	"windops/internal/platform/identity"
	"windops/internal/policy"
	"windops/internal/store"
)

func TestCoordinatorPersistsAuditAndOutboxAtomically(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "coordinator.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	coordinator := NewCoordinator(db, clock.Fixed{Value: now}, &identity.Sequence{})
	request := DecisionRequest{Rule: "WeatherWindowIsSafe", Actor: "planner-1", RequestID: "request-1", Context: policy.Context{TenantID: "tenant-a", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Confidence: 90, WindKPH: 40, WaveCM: 150}}
	response, err := coordinator.Evaluate(ctx, request)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if !response.Result.Allowed || response.AuditID == "" || response.OutboxID == "" {
		t.Fatalf("incomplete response: %+v", response)
	}
	overview, err := coordinator.Overview(ctx, "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if overview["audit_events"] != 1 || overview["outbox_jobs"] != 1 {
		t.Fatalf("decision records not atomic: %+v", overview)
	}
	var requests int
	if err := db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM request_log WHERE request_id='request-1'`).Scan(&requests); err != nil {
		t.Fatal(err)
	}
	if requests != 1 {
		t.Fatalf("request log count: %d", requests)
	}
}

func TestCoordinatorDenialPersistsAuditWithoutOutbox(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "denial.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	coordinator := NewCoordinator(db, clock.Fixed{Value: now}, &identity.Sequence{})
	request := DecisionRequest{Rule: "WeatherWindowIsSafe", Actor: "planner-1", RequestID: "request-denied", Context: policy.Context{TenantID: "tenant-a", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Confidence: 90, WindKPH: 90, WaveCM: 150}}
	response, err := coordinator.Evaluate(ctx, request)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if response.Result.Allowed || response.Result.Code != "unsafe_weather" {
		t.Fatalf("unexpected decision: %+v", response)
	}
	if response.AuditID == "" || response.OutboxID != "" {
		t.Fatalf("unexpected persistence IDs: %+v", response)
	}
	overview, err := coordinator.Overview(ctx, "tenant-a")
	if err != nil {
		t.Fatal(err)
	}
	if overview["audit_events"] != 1 || overview["outbox_jobs"] != 0 {
		t.Fatalf("denial persistence mismatch: %+v", overview)
	}
}
