package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/outbox"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type JobRepository struct{ db *Database }

func NewJobRepository(db *Database) *JobRepository { return &JobRepository{db: db} }

func (r *JobRepository) Create(ctx context.Context, value outbox.Job) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "outbox.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO outbox_jobs(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "outbox.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *JobRepository) Get(ctx context.Context, tenantID, id string) (outbox.Job, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM outbox_jobs WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return outbox.Job{}, fault.New(fault.CodeNotFound, "outbox.get", "record was not found")
		}
		return outbox.Job{}, fault.Wrap(fault.CodeInternal, "outbox.get", "query record", err)
	}
	var value outbox.Job
	if err := json.Unmarshal(raw, &value); err != nil {
		return outbox.Job{}, fault.Wrap(fault.CodeInternal, "outbox.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *JobRepository) Update(ctx context.Context, value outbox.Job, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "outbox.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE outbox_jobs SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "outbox.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "outbox.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "outbox.update", "record version changed")
	}
	return nil
}

func (r *JobRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[outbox.Job], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[outbox.Job]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM outbox_jobs WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[outbox.Job]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM outbox_jobs WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[outbox.Job]{}, err
	}
	defer rows.Close()
	items := make([]outbox.Job, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[outbox.Job]{}, err
		}
		var value outbox.Job
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[outbox.Job]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[outbox.Job]{}, err
	}
	return page.Result[outbox.Job]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
