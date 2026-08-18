package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
	"windops/internal/domain/slot"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

func TestSlotRepositoryPersistsAndUpdatesWithOptimisticVersion(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewSlotRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := slot.Slot{ID: "slot-1", TenantID: "tenant-a", FarmID: "farm-1", TurbineID: "turbine-1", WindowID: "window-1", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Capacity: 2, Used: 0, Status: slot.StatusOpen, Version: 1, CreatedAt: now, UpdatedAt: now}
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
	transitioned, err := loaded.WithStatus(slot.StatusHeld, now.Add(time.Minute))
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
	if updated.Status != slot.StatusHeld || updated.Version != 2 {
		t.Fatalf("update not persisted: %+v", updated)
	}
	if err := repo.Update(ctx, transitioned, loaded.Version); !fault.IsCode(err, fault.CodeConflict) {
		t.Fatalf("stale update should conflict, got %v", err)
	}
}

func TestSlotRepositoryFiltersTenantStatusAndPaginates(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewSlotRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	makeValue := func(id, tenant string, status slot.Status, offset int) slot.Slot {
		return slot.Slot{ID: id, TenantID: tenant, FarmID: "farm-1", TurbineID: "turbine-1", WindowID: "window-1", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Capacity: 2, Used: 0, Status: status, Version: 1, CreatedAt: now.Add(time.Duration(offset) * time.Minute), UpdatedAt: now.Add(time.Duration(offset) * time.Minute)}
	}
	fixtures := []slot.Slot{makeValue("slot-1", "tenant-a", slot.StatusOpen, 1), makeValue("slot-2", "tenant-a", slot.StatusOpen, 2), makeValue("slot-3", "tenant-a", slot.StatusHeld, 3), makeValue("slot-4", "tenant-b", slot.StatusOpen, 4)}
	for _, fixture := range fixtures {
		if err := repo.Create(ctx, fixture); err != nil {
			t.Fatalf("create %s: %v", fixture.ID, err)
		}
	}
	result, err := repo.List(ctx, "tenant-a", string(slot.StatusOpen), page.Request{Limit: 1, Offset: 0, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Total != 2 || len(result.Items) != 1 || result.Items[0].ID != "slot-2" {
		t.Fatalf("unexpected first page: %+v", result)
	}
	second, err := repo.List(ctx, "tenant-a", string(slot.StatusOpen), page.Request{Limit: 1, Offset: 1, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if second.Total != 2 || len(second.Items) != 1 || second.Items[0].ID != "slot-1" {
		t.Fatalf("unexpected second page: %+v", second)
	}
	if _, err := repo.Get(ctx, "tenant-b", "slot-1"); !fault.IsCode(err, fault.CodeNotFound) {
		t.Fatalf("cross-tenant get should not find record: %v", err)
	}
}

func TestSlotRepositorySurvivesDatabaseRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "restart.db")
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewSlotRepository(db)
	value := slot.Slot{ID: "slot-restart", TenantID: "tenant-a", FarmID: "farm-1", TurbineID: "turbine-1", WindowID: "window-1", StartsAt: now.Add(time.Hour), EndsAt: now.Add(3 * time.Hour), Capacity: 2, Used: 0, Status: slot.StatusOpen, Version: 1, CreatedAt: now, UpdatedAt: now}
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
	loaded, err := NewSlotRepository(reopened).Get(ctx, "tenant-a", value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != value.ID || loaded.Version != 1 {
		t.Fatalf("restart lost value: %+v", loaded)
	}
}
