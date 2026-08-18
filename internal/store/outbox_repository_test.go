package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
	"windops/internal/domain/outbox"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

func TestJobRepositoryPersistsAndUpdatesWithOptimisticVersion(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewJobRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := outbox.Job{ID: "outbox-1", TenantID: "tenant-a", Topic: "campaign.approved", ObjectID: "campaign-1", Payload: "{}", Attempts: 0, AvailableAt: now, Status: outbox.StatusPending, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(ctx, value); err != nil {
		t.Fatalf("create: %v", err)
	}
	loaded, err := repo.Get(ctx, "tenant-a", value.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if loaded.ID != value.ID || loaded.TenantID != "tenant-a" || loaded.Version != 1 {
		t.Fatalf("unexpected loaded value: %+v", loaded)
	}
	transitioned, err := loaded.WithStatus(outbox.StatusRunning, now.Add(time.Minute))
	if err != nil {
		t.Fatalf("transition: %v", err)
	}
	if err := repo.Update(ctx, transitioned, loaded.Version); err != nil {
		t.Fatalf("update: %v", err)
	}
	updated, err := repo.Get(ctx, "tenant-a", value.ID)
	if err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if updated.Status != outbox.StatusRunning || updated.Version != 2 {
		t.Fatalf("update not persisted: %+v", updated)
	}
	if err := repo.Update(ctx, transitioned, loaded.Version); !fault.IsCode(err, fault.CodeConflict) {
		t.Fatalf("stale update should conflict, got %v", err)
	}
}

func TestJobRepositoryFiltersTenantStatusAndPaginates(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewJobRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	makeValue := func(id, tenant string, status outbox.Status, offset int) outbox.Job {
		return outbox.Job{ID: id, TenantID: tenant, Topic: "campaign.approved", ObjectID: "campaign-1", Payload: "{}", Attempts: 0, AvailableAt: now, Status: status, Version: 1, CreatedAt: now.Add(time.Duration(offset) * time.Minute), UpdatedAt: now.Add(time.Duration(offset) * time.Minute)}
	}
	fixtures := []outbox.Job{makeValue("outbox-1", "tenant-a", outbox.StatusPending, 1), makeValue("outbox-2", "tenant-a", outbox.StatusPending, 2), makeValue("outbox-3", "tenant-a", outbox.StatusRunning, 3), makeValue("outbox-4", "tenant-b", outbox.StatusPending, 4)}
	for _, fixture := range fixtures {
		if err := repo.Create(ctx, fixture); err != nil {
			t.Fatalf("create %s: %v", fixture.ID, err)
		}
	}
	result, err := repo.List(ctx, "tenant-a", string(outbox.StatusPending), page.Request{Limit: 1, Offset: 0, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Total != 2 || len(result.Items) != 1 || result.Items[0].ID != "outbox-2" {
		t.Fatalf("unexpected first page: %+v", result)
	}
	second, err := repo.List(ctx, "tenant-a", string(outbox.StatusPending), page.Request{Limit: 1, Offset: 1, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if second.Total != 2 || len(second.Items) != 1 || second.Items[0].ID != "outbox-1" {
		t.Fatalf("unexpected second page: %+v", second)
	}
	if _, err := repo.Get(ctx, "tenant-b", "outbox-1"); !fault.IsCode(err, fault.CodeNotFound) {
		t.Fatalf("cross-tenant get should not find record: %v", err)
	}
}

func TestJobRepositorySurvivesDatabaseRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "restart.db")
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewJobRepository(db)
	value := outbox.Job{ID: "outbox-restart", TenantID: "tenant-a", Topic: "campaign.approved", ObjectID: "campaign-1", Payload: "{}", Attempts: 0, AvailableAt: now, Status: outbox.StatusPending, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := repo.Create(ctx, value); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	reopened, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	defer reopened.Close()
	loaded, err := NewJobRepository(reopened).Get(ctx, "tenant-a", value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != value.ID || loaded.Version != 1 {
		t.Fatalf("restart lost value: %+v", loaded)
	}
}
