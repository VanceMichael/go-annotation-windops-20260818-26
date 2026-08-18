package store

import (
	"context"
	"path/filepath"
	"testing"
	"time"
	"windops/internal/domain/crew"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

func TestMemberRepositoryPersistsAndUpdatesWithOptimisticVersion(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewMemberRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	value := crew.Member{ID: "crew-1", TenantID: "tenant-a", Name: "Lin Yue", HomePort: "Donggang", MaxHours: 12, LastShiftEnd: now.Add(-12 * time.Hour), Status: crew.StatusAvailable, Version: 1, CreatedAt: now, UpdatedAt: now}
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
	transitioned, err := loaded.WithStatus(crew.StatusAssigned, now.Add(time.Minute))
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
	if updated.Status != crew.StatusAssigned || updated.Version != 2 {
		t.Fatalf("update not persisted: %+v", updated)
	}
	if err := repo.Update(ctx, transitioned, loaded.Version); !fault.IsCode(err, fault.CodeConflict) {
		t.Fatalf("stale update should conflict, got %v", err)
	}
}

func TestMemberRepositoryFiltersTenantStatusAndPaginates(t *testing.T) {
	ctx := context.Background()
	db := newTestDatabase(t)
	repo := NewMemberRepository(db)
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	makeValue := func(id, tenant string, status crew.Status, offset int) crew.Member {
		return crew.Member{ID: id, TenantID: tenant, Name: "Lin Yue", HomePort: "Donggang", MaxHours: 12, LastShiftEnd: now.Add(-12 * time.Hour), Status: status, Version: 1, CreatedAt: now.Add(time.Duration(offset) * time.Minute), UpdatedAt: now.Add(time.Duration(offset) * time.Minute)}
	}
	fixtures := []crew.Member{makeValue("crew-1", "tenant-a", crew.StatusAvailable, 1), makeValue("crew-2", "tenant-a", crew.StatusAvailable, 2), makeValue("crew-3", "tenant-a", crew.StatusAssigned, 3), makeValue("crew-4", "tenant-b", crew.StatusAvailable, 4)}
	for _, fixture := range fixtures {
		if err := repo.Create(ctx, fixture); err != nil {
			t.Fatalf("create %s: %v", fixture.ID, err)
		}
	}
	result, err := repo.List(ctx, "tenant-a", string(crew.StatusAvailable), page.Request{Limit: 1, Offset: 0, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if result.Total != 2 || len(result.Items) != 1 || result.Items[0].ID != "crew-2" {
		t.Fatalf("unexpected first page: %+v", result)
	}
	second, err := repo.List(ctx, "tenant-a", string(crew.StatusAvailable), page.Request{Limit: 1, Offset: 1, Sort: "updated_at", Desc: true})
	if err != nil {
		t.Fatalf("second page: %v", err)
	}
	if second.Total != 2 || len(second.Items) != 1 || second.Items[0].ID != "crew-1" {
		t.Fatalf("unexpected second page: %+v", second)
	}
	if _, err := repo.Get(ctx, "tenant-b", "crew-1"); !fault.IsCode(err, fault.CodeNotFound) {
		t.Fatalf("cross-tenant get should not find record: %v", err)
	}
}

func TestMemberRepositorySurvivesDatabaseRestart(t *testing.T) {
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "restart.db")
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	db, err := Open(ctx, path)
	if err != nil {
		t.Fatal(err)
	}
	repo := NewMemberRepository(db)
	value := crew.Member{ID: "crew-restart", TenantID: "tenant-a", Name: "Lin Yue", HomePort: "Donggang", MaxHours: 12, LastShiftEnd: now.Add(-12 * time.Hour), Status: crew.StatusAvailable, Version: 1, CreatedAt: now, UpdatedAt: now}
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
	loaded, err := NewMemberRepository(reopened).Get(ctx, "tenant-a", value.ID)
	if err != nil {
		t.Fatal(err)
	}
	if loaded.ID != value.ID || loaded.Version != 1 {
		t.Fatalf("restart lost value: %+v", loaded)
	}
}
