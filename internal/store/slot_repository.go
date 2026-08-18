package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/slot"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type SlotRepository struct{ db *Database }

func NewSlotRepository(db *Database) *SlotRepository { return &SlotRepository{db: db} }

func (r *SlotRepository) Create(ctx context.Context, value slot.Slot) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "slot.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO maintenance_slots(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "slot.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *SlotRepository) Get(ctx context.Context, tenantID, id string) (slot.Slot, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM maintenance_slots WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return slot.Slot{}, fault.New(fault.CodeNotFound, "slot.get", "record was not found")
		}
		return slot.Slot{}, fault.Wrap(fault.CodeInternal, "slot.get", "query record", err)
	}
	var value slot.Slot
	if err := json.Unmarshal(raw, &value); err != nil {
		return slot.Slot{}, fault.Wrap(fault.CodeInternal, "slot.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *SlotRepository) Update(ctx context.Context, value slot.Slot, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "slot.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE maintenance_slots SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "slot.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "slot.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "slot.update", "record version changed")
	}
	return nil
}

func (r *SlotRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[slot.Slot], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[slot.Slot]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM maintenance_slots WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[slot.Slot]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM maintenance_slots WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[slot.Slot]{}, err
	}
	defer rows.Close()
	items := make([]slot.Slot, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[slot.Slot]{}, err
		}
		var value slot.Slot
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[slot.Slot]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[slot.Slot]{}, err
	}
	return page.Result[slot.Slot]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
