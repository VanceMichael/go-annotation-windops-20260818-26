package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/workorder"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type WorkOrderRepository struct{ db *Database }

func NewWorkOrderRepository(db *Database) *WorkOrderRepository { return &WorkOrderRepository{db: db} }

func (r *WorkOrderRepository) Create(ctx context.Context, value workorder.WorkOrder) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "workorder.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO work_orders(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "workorder.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *WorkOrderRepository) Get(ctx context.Context, tenantID, id string) (workorder.WorkOrder, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM work_orders WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return workorder.WorkOrder{}, fault.New(fault.CodeNotFound, "workorder.get", "record was not found")
		}
		return workorder.WorkOrder{}, fault.Wrap(fault.CodeInternal, "workorder.get", "query record", err)
	}
	var value workorder.WorkOrder
	if err := json.Unmarshal(raw, &value); err != nil {
		return workorder.WorkOrder{}, fault.Wrap(fault.CodeInternal, "workorder.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *WorkOrderRepository) Update(ctx context.Context, value workorder.WorkOrder, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "workorder.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE work_orders SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "workorder.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "workorder.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "workorder.update", "record version changed")
	}
	return nil
}

func (r *WorkOrderRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[workorder.WorkOrder], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[workorder.WorkOrder]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM work_orders WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[workorder.WorkOrder]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM work_orders WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[workorder.WorkOrder]{}, err
	}
	defer rows.Close()
	items := make([]workorder.WorkOrder, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[workorder.WorkOrder]{}, err
		}
		var value workorder.WorkOrder
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[workorder.WorkOrder]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[workorder.WorkOrder]{}, err
	}
	return page.Result[workorder.WorkOrder]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
