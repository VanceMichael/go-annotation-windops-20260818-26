package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/weather"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type WindowRepository struct{ db *Database }

func NewWindowRepository(db *Database) *WindowRepository { return &WindowRepository{db: db} }

func (r *WindowRepository) Create(ctx context.Context, value weather.Window) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "weather.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO weather_windows(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "weather.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *WindowRepository) Get(ctx context.Context, tenantID, id string) (weather.Window, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM weather_windows WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return weather.Window{}, fault.New(fault.CodeNotFound, "weather.get", "record was not found")
		}
		return weather.Window{}, fault.Wrap(fault.CodeInternal, "weather.get", "query record", err)
	}
	var value weather.Window
	if err := json.Unmarshal(raw, &value); err != nil {
		return weather.Window{}, fault.Wrap(fault.CodeInternal, "weather.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *WindowRepository) Update(ctx context.Context, value weather.Window, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "weather.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE weather_windows SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "weather.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "weather.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "weather.update", "record version changed")
	}
	return nil
}

func (r *WindowRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[weather.Window], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[weather.Window]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM weather_windows WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[weather.Window]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM weather_windows WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[weather.Window]{}, err
	}
	defer rows.Close()
	items := make([]weather.Window, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[weather.Window]{}, err
		}
		var value weather.Window
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[weather.Window]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[weather.Window]{}, err
	}
	return page.Result[weather.Window]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
