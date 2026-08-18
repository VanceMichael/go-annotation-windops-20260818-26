package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/vessel"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type VesselRepository struct{ db *Database }

func NewVesselRepository(db *Database) *VesselRepository { return &VesselRepository{db: db} }

func (r *VesselRepository) Create(ctx context.Context, value vessel.Vessel) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "vessel.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO vessels(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "vessel.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *VesselRepository) Get(ctx context.Context, tenantID, id string) (vessel.Vessel, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM vessels WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return vessel.Vessel{}, fault.New(fault.CodeNotFound, "vessel.get", "record was not found")
		}
		return vessel.Vessel{}, fault.Wrap(fault.CodeInternal, "vessel.get", "query record", err)
	}
	var value vessel.Vessel
	if err := json.Unmarshal(raw, &value); err != nil {
		return vessel.Vessel{}, fault.Wrap(fault.CodeInternal, "vessel.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *VesselRepository) Update(ctx context.Context, value vessel.Vessel, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "vessel.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE vessels SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "vessel.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "vessel.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "vessel.update", "record version changed")
	}
	return nil
}

func (r *VesselRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[vessel.Vessel], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[vessel.Vessel]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vessels WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[vessel.Vessel]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM vessels WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[vessel.Vessel]{}, err
	}
	defer rows.Close()
	items := make([]vessel.Vessel, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[vessel.Vessel]{}, err
		}
		var value vessel.Vessel
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[vessel.Vessel]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[vessel.Vessel]{}, err
	}
	return page.Result[vessel.Vessel]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
