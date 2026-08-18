package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/dispatch"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type DispatchRepository struct{ db *Database }

func NewDispatchRepository(db *Database) *DispatchRepository { return &DispatchRepository{db: db} }

func (r *DispatchRepository) Create(ctx context.Context, value dispatch.Dispatch) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "dispatch.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO dispatches(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "dispatch.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *DispatchRepository) Get(ctx context.Context, tenantID, id string) (dispatch.Dispatch, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM dispatches WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dispatch.Dispatch{}, fault.New(fault.CodeNotFound, "dispatch.get", "record was not found")
		}
		return dispatch.Dispatch{}, fault.Wrap(fault.CodeInternal, "dispatch.get", "query record", err)
	}
	var value dispatch.Dispatch
	if err := json.Unmarshal(raw, &value); err != nil {
		return dispatch.Dispatch{}, fault.Wrap(fault.CodeInternal, "dispatch.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *DispatchRepository) Update(ctx context.Context, value dispatch.Dispatch, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "dispatch.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE dispatches SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "dispatch.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "dispatch.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "dispatch.update", "record version changed")
	}
	return nil
}

func (r *DispatchRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[dispatch.Dispatch], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[dispatch.Dispatch]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM dispatches WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[dispatch.Dispatch]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM dispatches WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[dispatch.Dispatch]{}, err
	}
	defer rows.Close()
	items := make([]dispatch.Dispatch, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[dispatch.Dispatch]{}, err
		}
		var value dispatch.Dispatch
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[dispatch.Dispatch]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[dispatch.Dispatch]{}, err
	}
	return page.Result[dispatch.Dispatch]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
