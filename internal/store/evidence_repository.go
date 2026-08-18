package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/evidence"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type EvidenceRepository struct{ db *Database }

func NewEvidenceRepository(db *Database) *EvidenceRepository { return &EvidenceRepository{db: db} }

func (r *EvidenceRepository) Create(ctx context.Context, value evidence.Evidence) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "evidence.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO evidence(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "evidence.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *EvidenceRepository) Get(ctx context.Context, tenantID, id string) (evidence.Evidence, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM evidence WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return evidence.Evidence{}, fault.New(fault.CodeNotFound, "evidence.get", "record was not found")
		}
		return evidence.Evidence{}, fault.Wrap(fault.CodeInternal, "evidence.get", "query record", err)
	}
	var value evidence.Evidence
	if err := json.Unmarshal(raw, &value); err != nil {
		return evidence.Evidence{}, fault.Wrap(fault.CodeInternal, "evidence.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *EvidenceRepository) Update(ctx context.Context, value evidence.Evidence, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "evidence.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE evidence SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "evidence.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "evidence.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "evidence.update", "record version changed")
	}
	return nil
}

func (r *EvidenceRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[evidence.Evidence], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[evidence.Evidence]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM evidence WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[evidence.Evidence]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM evidence WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[evidence.Evidence]{}, err
	}
	defer rows.Close()
	items := make([]evidence.Evidence, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[evidence.Evidence]{}, err
		}
		var value evidence.Evidence
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[evidence.Evidence]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[evidence.Evidence]{}, err
	}
	return page.Result[evidence.Evidence]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
