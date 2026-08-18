package application

import (
	"context"
	"time"
	"windops/internal/domain/farm"
	"windops/internal/domain/turbine"
	"windops/internal/domain/vessel"
	"windops/internal/platform/identity"
	"windops/internal/platform/page"
	"windops/internal/store"
)

func Bootstrap(ctx context.Context, db *store.Database, ids identity.Generator, tenant string, now time.Time) error {
	farms := store.NewFarmRepository(db)
	list, err := farms.List(ctx, tenant, "", pageRequest())
	if err != nil {
		return err
	}
	if list.Total > 0 {
		return nil
	}
	farmID := ids.New("farm")
	f := farm.Farm{ID: farmID, TenantID: tenant, Name: "East Shoal Array", Timezone: "Asia/Shanghai", CoastZone: "east-shoal", CapacityMW: 420, Status: farm.StatusActive, Version: 1, CreatedAt: now, UpdatedAt: now}
	if err := farms.Create(ctx, f); err != nil {
		return err
	}
	turbines := store.NewTurbineRepository(db)
	for i, code := range []string{"ES-A01", "ES-A02", "ES-B01", "ES-B02"} {
		value := turbine.Turbine{ID: ids.New("turbine"), TenantID: tenant, FarmID: farmID, Code: code, RatedKW: 8000 + i*500, LastServiceAt: now.AddDate(0, -2, 0), Status: turbine.StatusAvailable, Version: 1, CreatedAt: now, UpdatedAt: now}
		if err := turbines.Create(ctx, value); err != nil {
			return err
		}
	}
	vessels := store.NewVesselRepository(db)
	for _, item := range []struct {
		name         string
		seats, cargo int
	}{{"Hai Gong 7", 18, 12000}, {"Lan Jing 3", 12, 8000}} {
		value := vessel.Vessel{ID: ids.New("vessel"), TenantID: tenant, Name: item.name, HomePort: "Donggang", SeatCapacity: item.seats, CargoCapacityKG: item.cargo, InspectionDue: now.AddDate(1, 0, 0), Status: vessel.StatusAvailable, Version: 1, CreatedAt: now, UpdatedAt: now}
		if err := vessels.Create(ctx, value); err != nil {
			return err
		}
	}
	return nil
}

func pageRequest() page.Request { return page.Request{Limit: 1, Sort: "id"} }
