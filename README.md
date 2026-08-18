# WindOps

WindOps is an offshore wind maintenance-window and permit coordination platform. It connects farm operators, marine coordinators, qualified crews, vessels and turbine engineers through a persisted workflow: forecast windows are confirmed, campaigns are approved, vessel permits are activated, work is dispatched, evidence is accepted, and audit/outbox records survive restarts.

The backend uses Go 1.22 and SQLite with versioned migrations. The operations workspace uses React, TypeScript, Vite and Ant Design. No online service is required for tests.

## Run

```bash
GOTOOLCHAIN=local go run ./cmd/server
cd web && npm ci && npm run dev
```

Backend defaults to `:8080`; the Vite development server proxies `/api` to it. Configure `WINDOPS_DB`, `WINDOPS_ADDR`, `WINDOPS_WEB`, and `WINDOPS_TENANT` as needed.

## Workflows

- Planning: confirm a safe weather window, create a maintenance campaign, reserve turbine slots, and approve a constrained budget.
- Marine dispatch: validate vessel inspection, seat/cargo capacity, crew qualifications and rest periods before activating a permit.
- Field execution: assign work, record start/finish state, require evidence, close campaigns, and publish durable outbox notifications.
- Recovery: SQLite migrations, optimistic versions, idempotency scope, audit records, retry state and resource locks survive process restarts.

## Verification

```bash
make test
make race
make vet
make build
make web-install
make web-test
make web-typecheck
make web-build
docker build -t windops .
```

