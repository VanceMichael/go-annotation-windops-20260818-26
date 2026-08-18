package store

import (
	"context"
	"database/sql"
	"fmt"
)

type migration struct {
	version    int
	statements []string
}

var migrations = []migration{
	{version: 1, statements: []string{
		`CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`,
		`CREATE TABLE IF NOT EXISTS farms(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_farms_tenant_status ON farms(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS turbines(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_turbines_tenant_status ON turbines(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS weather_windows(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_weather_windows_tenant_status ON weather_windows(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS campaigns(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_campaigns_tenant_status ON campaigns(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS crew_members(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_crew_members_tenant_status ON crew_members(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS qualifications(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_qualifications_tenant_status ON qualifications(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS vessels(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_vessels_tenant_status ON vessels(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS permits(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_permits_tenant_status ON permits(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS work_orders(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_work_orders_tenant_status ON work_orders(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS dispatches(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_dispatches_tenant_status ON dispatches(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS evidence(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_evidence_tenant_status ON evidence(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS audit_events(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_audit_events_tenant_status ON audit_events(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS outbox_jobs(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_outbox_jobs_tenant_status ON outbox_jobs(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS idempotency_records(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_idempotency_records_tenant_status ON idempotency_records(tenant_id,status,updated_at)`,
		`CREATE TABLE IF NOT EXISTS maintenance_slots(id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, data_json TEXT NOT NULL, status TEXT NOT NULL, version INTEGER NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL)`,
		`CREATE INDEX IF NOT EXISTS idx_maintenance_slots_tenant_status ON maintenance_slots(tenant_id,status,updated_at)`,
	}},
	{version: 2, statements: []string{
		`CREATE TABLE IF NOT EXISTS entity_links(tenant_id TEXT NOT NULL, source_type TEXT NOT NULL, source_id TEXT NOT NULL, relation TEXT NOT NULL, target_type TEXT NOT NULL, target_id TEXT NOT NULL, created_at TEXT NOT NULL, PRIMARY KEY(tenant_id,source_type,source_id,relation,target_type,target_id))`,
		`CREATE INDEX IF NOT EXISTS idx_entity_links_target ON entity_links(tenant_id,target_type,target_id,relation)`,
		`CREATE TABLE IF NOT EXISTS workflow_locks(tenant_id TEXT NOT NULL, resource_type TEXT NOT NULL, resource_id TEXT NOT NULL, owner_id TEXT NOT NULL, expires_at TEXT NOT NULL, version INTEGER NOT NULL, PRIMARY KEY(tenant_id,resource_type,resource_id))`,
		`CREATE TABLE IF NOT EXISTS request_log(request_id TEXT PRIMARY KEY, tenant_id TEXT NOT NULL, method TEXT NOT NULL, path TEXT NOT NULL, outcome TEXT NOT NULL, created_at TEXT NOT NULL)`,
	}},
}

func (d *Database) Migrate(ctx context.Context) error {
	if _, err := d.db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS schema_migrations(version INTEGER PRIMARY KEY, applied_at TEXT NOT NULL)`); err != nil {
		return err
	}
	return d.WithTx(ctx, func(tx *sql.Tx) error {
		for _, m := range migrations {
			var count int
			if err := tx.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE version=?`, m.version).Scan(&count); err != nil {
				return err
			}
			if count > 0 {
				continue
			}
			for _, statement := range m.statements {
				if _, err := tx.ExecContext(ctx, statement); err != nil {
					return fmt.Errorf("migration %d: %w", m.version, err)
				}
			}
			if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(version,applied_at) VALUES(?,strftime('%Y-%m-%dT%H:%M:%fZ','now'))`, m.version); err != nil {
				return err
			}
		}
		return nil
	})
}
