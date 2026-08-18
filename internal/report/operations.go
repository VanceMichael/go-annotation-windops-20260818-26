package report

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"time"
	"windops/internal/store"
)

type StatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}
type Operations struct {
	TenantID      string        `json:"tenant_id"`
	GeneratedAt   time.Time     `json:"generated_at"`
	Campaigns     []StatusCount `json:"campaigns"`
	Permits       []StatusCount `json:"permits"`
	WorkOrders    []StatusCount `json:"work_orders"`
	Dispatches    []StatusCount `json:"dispatches"`
	PendingOutbox int           `json:"pending_outbox"`
	OpenLocks     int           `json:"open_locks"`
	LatestAuditAt *time.Time    `json:"latest_audit_at,omitempty"`
}

type Builder struct{ db *store.Database }

func New(db *store.Database) *Builder { return &Builder{db: db} }

func (b *Builder) Build(ctx context.Context, tenant string, now time.Time) (Operations, error) {
	if tenant == "" {
		return Operations{}, fmt.Errorf("tenant is required")
	}
	result := Operations{TenantID: tenant, GeneratedAt: now.UTC()}
	var err error
	if result.Campaigns, err = b.statuses(ctx, "campaigns", tenant); err != nil {
		return Operations{}, err
	}
	if result.Permits, err = b.statuses(ctx, "permits", tenant); err != nil {
		return Operations{}, err
	}
	if result.WorkOrders, err = b.statuses(ctx, "work_orders", tenant); err != nil {
		return Operations{}, err
	}
	if result.Dispatches, err = b.statuses(ctx, "dispatches", tenant); err != nil {
		return Operations{}, err
	}
	if err = b.db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM outbox_jobs WHERE tenant_id=? AND status IN ('pending','retry','running')`, tenant).Scan(&result.PendingOutbox); err != nil {
		return Operations{}, fmt.Errorf("count outbox: %w", err)
	}
	if err = b.db.SQL().QueryRowContext(ctx, `SELECT COUNT(*) FROM workflow_locks WHERE tenant_id=? AND expires_at>?`, tenant, now.UTC().Format(time.RFC3339Nano)).Scan(&result.OpenLocks); err != nil {
		return Operations{}, fmt.Errorf("count locks: %w", err)
	}
	var raw sql.NullString
	if err = b.db.SQL().QueryRowContext(ctx, `SELECT MAX(created_at) FROM audit_events WHERE tenant_id=?`, tenant).Scan(&raw); err != nil {
		return Operations{}, fmt.Errorf("latest audit: %w", err)
	}
	if raw.Valid {
		parsed, err := time.Parse(time.RFC3339Nano, raw.String)
		if err != nil {
			return Operations{}, fmt.Errorf("parse latest audit: %w", err)
		}
		result.LatestAuditAt = &parsed
	}
	return result, nil
}

func (b *Builder) statuses(ctx context.Context, table, tenant string) ([]StatusCount, error) {
	allowed := map[string]bool{"campaigns": true, "permits": true, "work_orders": true, "dispatches": true}
	if !allowed[table] {
		return nil, fmt.Errorf("unsupported report table %q", table)
	}
	rows, err := b.db.SQL().QueryContext(ctx, "SELECT status,COUNT(*) FROM "+table+" WHERE tenant_id=? GROUP BY status", tenant)
	if err != nil {
		return nil, fmt.Errorf("query %s statuses: %w", table, err)
	}
	defer rows.Close()
	result := []StatusCount{}
	for rows.Next() {
		var item StatusCount
		if err := rows.Scan(&item.Status, &item.Count); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Status < result[j].Status })
	return result, nil
}

func Total(items []StatusCount) int {
	total := 0
	for _, item := range items {
		if item.Count > 0 {
			total += item.Count
		}
	}
	return total
}
func Find(items []StatusCount, status string) int {
	for _, item := range items {
		if item.Status == status {
			return item.Count
		}
	}
	return 0
}
