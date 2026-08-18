package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/turbine"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type TurbineRepository struct{ db *Database }

func NewTurbineRepository(db *Database) *TurbineRepository { return &TurbineRepository{db: db} }

func (r *TurbineRepository) Create(ctx context.Context, value turbine.Turbine) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "turbine.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO turbines(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "turbine.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *TurbineRepository) Get(ctx context.Context, tenantID, id string) (turbine.Turbine, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM turbines WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return turbine.Turbine{}, fault.New(fault.CodeNotFound, "turbine.get", "record was not found")
		}
		return turbine.Turbine{}, fault.Wrap(fault.CodeInternal, "turbine.get", "query record", err)
	}
	var value turbine.Turbine
	if err := json.Unmarshal(raw, &value); err != nil {
		return turbine.Turbine{}, fault.Wrap(fault.CodeInternal, "turbine.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *TurbineRepository) Update(ctx context.Context, value turbine.Turbine, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "turbine.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE turbines SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "turbine.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "turbine.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "turbine.update", "record version changed")
	}
	return nil
}

func (r *TurbineRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[turbine.Turbine], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[turbine.Turbine]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM turbines WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[turbine.Turbine]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM turbines WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[turbine.Turbine]{}, err
	}
	defer rows.Close()
	items := make([]turbine.Turbine, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[turbine.Turbine]{}, err
		}
		var value turbine.Turbine
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[turbine.Turbine]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[turbine.Turbine]{}, err
	}
	return page.Result[turbine.Turbine]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
