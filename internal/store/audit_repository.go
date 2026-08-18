package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/audit"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type EventRepository struct{ db *Database }

func NewEventRepository(db *Database) *EventRepository { return &EventRepository{db: db} }

func (r *EventRepository) Create(ctx context.Context, value audit.Event) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "audit.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO audit_events(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "audit.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *EventRepository) Get(ctx context.Context, tenantID, id string) (audit.Event, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM audit_events WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return audit.Event{}, fault.New(fault.CodeNotFound, "audit.get", "record was not found")
		}
		return audit.Event{}, fault.Wrap(fault.CodeInternal, "audit.get", "query record", err)
	}
	var value audit.Event
	if err := json.Unmarshal(raw, &value); err != nil {
		return audit.Event{}, fault.Wrap(fault.CodeInternal, "audit.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *EventRepository) Update(ctx context.Context, value audit.Event, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "audit.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE audit_events SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "audit.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "audit.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "audit.update", "record version changed")
	}
	return nil
}

func (r *EventRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[audit.Event], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[audit.Event]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM audit_events WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[audit.Event]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM audit_events WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[audit.Event]{}, err
	}
	defer rows.Close()
	items := make([]audit.Event, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[audit.Event]{}, err
		}
		var value audit.Event
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[audit.Event]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[audit.Event]{}, err
	}
	return page.Result[audit.Event]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
