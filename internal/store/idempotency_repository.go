package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/idempotency"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type RecordRepository struct{ db *Database }

func NewRecordRepository(db *Database) *RecordRepository { return &RecordRepository{db: db} }

func (r *RecordRepository) Create(ctx context.Context, value idempotency.Record) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "idempotency.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO idempotency_records(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "idempotency.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *RecordRepository) Get(ctx context.Context, tenantID, id string) (idempotency.Record, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM idempotency_records WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return idempotency.Record{}, fault.New(fault.CodeNotFound, "idempotency.get", "record was not found")
		}
		return idempotency.Record{}, fault.Wrap(fault.CodeInternal, "idempotency.get", "query record", err)
	}
	var value idempotency.Record
	if err := json.Unmarshal(raw, &value); err != nil {
		return idempotency.Record{}, fault.Wrap(fault.CodeInternal, "idempotency.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *RecordRepository) Update(ctx context.Context, value idempotency.Record, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "idempotency.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE idempotency_records SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "idempotency.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "idempotency.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "idempotency.update", "record version changed")
	}
	return nil
}

func (r *RecordRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[idempotency.Record], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[idempotency.Record]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM idempotency_records WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[idempotency.Record]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM idempotency_records WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[idempotency.Record]{}, err
	}
	defer rows.Close()
	items := make([]idempotency.Record, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[idempotency.Record]{}, err
		}
		var value idempotency.Record
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[idempotency.Record]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[idempotency.Record]{}, err
	}
	return page.Result[idempotency.Record]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
