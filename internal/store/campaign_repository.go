package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"windops/internal/domain/campaign"
	"windops/internal/fault"
	"windops/internal/platform/page"
)

type CampaignRepository struct{ db *Database }

func NewCampaignRepository(db *Database) *CampaignRepository { return &CampaignRepository{db: db} }

func (r *CampaignRepository) Create(ctx context.Context, value campaign.Campaign) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "campaign.create", "encode record", err)
	}
	_, err = r.db.db.ExecContext(ctx, `INSERT INTO campaigns(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, value.ID, value.TenantID, data, value.Status, value.Version, utc(value.CreatedAt), utc(value.UpdatedAt))
	if err != nil {
		return fault.Wrap(fault.CodeConflict, "campaign.create", "record already exists or violates a constraint", err)
	}
	return nil
}

func (r *CampaignRepository) Get(ctx context.Context, tenantID, id string) (campaign.Campaign, error) {
	var raw []byte
	if err := r.db.db.QueryRowContext(ctx, `SELECT data_json FROM campaigns WHERE tenant_id=? AND id=?`, tenantID, id).Scan(&raw); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return campaign.Campaign{}, fault.New(fault.CodeNotFound, "campaign.get", "record was not found")
		}
		return campaign.Campaign{}, fault.Wrap(fault.CodeInternal, "campaign.get", "query record", err)
	}
	var value campaign.Campaign
	if err := json.Unmarshal(raw, &value); err != nil {
		return campaign.Campaign{}, fault.Wrap(fault.CodeInternal, "campaign.get", "decode record", err)
	}
	return value.Clone(), nil
}

func (r *CampaignRepository) Update(ctx context.Context, value campaign.Campaign, expectedVersion int64) error {
	if err := value.Validate(); err != nil {
		return err
	}
	data, err := json.Marshal(value)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "campaign.update", "encode record", err)
	}
	result, err := r.db.db.ExecContext(ctx, `UPDATE campaigns SET data_json=?,status=?,version=?,updated_at=? WHERE tenant_id=? AND id=? AND version=?`, data, value.Status, value.Version, utc(value.UpdatedAt), value.TenantID, value.ID, expectedVersion)
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "campaign.update", "update record", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fault.Wrap(fault.CodeInternal, "campaign.update", "read update result", err)
	}
	if changed != 1 {
		return fault.New(fault.CodeConflict, "campaign.update", "record version changed")
	}
	return nil
}

func (r *CampaignRepository) List(ctx context.Context, tenantID, status string, request page.Request) (page.Result[campaign.Campaign], error) {
	normalized, err := request.Normalize(map[string]bool{"updated_at": true, "created_at": true, "id": true}, "updated_at")
	if err != nil {
		return page.Result[campaign.Campaign]{}, err
	}
	where, args := "tenant_id=?", []any{tenantID}
	if status != "" {
		where += " AND status=?"
		args = append(args, status)
	}
	var total int
	if err := r.db.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM campaigns WHERE "+where, args...).Scan(&total); err != nil {
		return page.Result[campaign.Campaign]{}, err
	}
	direction := "ASC"
	if normalized.Desc {
		direction = "DESC"
	}
	args = append(args, normalized.Limit, normalized.Offset)
	rows, err := r.db.db.QueryContext(ctx, fmt.Sprintf("SELECT data_json FROM campaigns WHERE %s ORDER BY %s %s,id ASC LIMIT ? OFFSET ?", where, normalized.Sort, direction), args...)
	if err != nil {
		return page.Result[campaign.Campaign]{}, err
	}
	defer rows.Close()
	items := make([]campaign.Campaign, 0, normalized.Limit)
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return page.Result[campaign.Campaign]{}, err
		}
		var value campaign.Campaign
		if err := json.Unmarshal(raw, &value); err != nil {
			return page.Result[campaign.Campaign]{}, err
		}
		items = append(items, value.Clone())
	}
	if err := rows.Err(); err != nil {
		return page.Result[campaign.Campaign]{}, err
	}
	return page.Result[campaign.Campaign]{Items: items, Total: total, Limit: normalized.Limit, Offset: normalized.Offset}, nil
}
