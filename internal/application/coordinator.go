package application

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
	"windops/internal/domain/audit"
	"windops/internal/domain/outbox"
	"windops/internal/fault"
	"windops/internal/platform/clock"
	"windops/internal/platform/identity"
	"windops/internal/policy"
	"windops/internal/store"
)

type Coordinator struct {
	DB    *store.Database
	Clock clock.Clock
	IDs   identity.Generator
}

func NewCoordinator(db *store.Database, c clock.Clock, ids identity.Generator) *Coordinator {
	return &Coordinator{DB: db, Clock: c, IDs: ids}
}

type DecisionRequest struct {
	Rule      string         `json:"rule"`
	Context   policy.Context `json:"context"`
	Actor     string         `json:"actor"`
	RequestID string         `json:"request_id"`
}
type DecisionResponse struct {
	Rule     string        `json:"rule"`
	Result   policy.Result `json:"result"`
	AuditID  string        `json:"audit_id"`
	OutboxID string        `json:"outbox_id,omitempty"`
}

var evaluators = map[string]func(policy.Context) (policy.Result, error){
	"WeatherWindowIsSafe":             policy.EvaluateWeatherWindowIsSafe,
	"WeatherWindowDoesNotOverlap":     policy.EvaluateWeatherWindowDoesNotOverlap,
	"CampaignFitsBudget":              policy.EvaluateCampaignFitsBudget,
	"SlotHasCapacity":                 policy.EvaluateSlotHasCapacity,
	"CampaignCanBeApproved":           policy.EvaluateCampaignCanBeApproved,
	"VesselCanBeReserved":             policy.EvaluateVesselCanBeReserved,
	"CrewMemberIsQualified":           policy.EvaluateCrewMemberIsQualified,
	"CrewMemberHasRested":             policy.EvaluateCrewMemberHasRested,
	"CrewAssignmentsDoNotOverlap":     policy.EvaluateCrewAssignmentsDoNotOverlap,
	"PermitCanBeSubmitted":            policy.EvaluatePermitCanBeSubmitted,
	"PermitCanBeActivated":            policy.EvaluatePermitCanBeActivated,
	"PermitExpiryIsBounded":           policy.EvaluatePermitExpiryIsBounded,
	"DispatchFitsVessel":              policy.EvaluateDispatchFitsVessel,
	"DispatchDepartureIsUnique":       policy.EvaluateDispatchDepartureIsUnique,
	"WorkOrderCanStart":               policy.EvaluateWorkOrderCanStart,
	"WorkOrderCanComplete":            policy.EvaluateWorkOrderCanComplete,
	"EvidenceChecksumIsUnique":        policy.EvaluateEvidenceChecksumIsUnique,
	"CampaignCanClose":                policy.EvaluateCampaignCanClose,
	"IdempotencyScopeMatches":         policy.EvaluateIdempotencyScopeMatches,
	"OptimisticVersionMatches":        policy.EvaluateOptimisticVersionMatches,
	"OutboxRetryIsDue":                policy.EvaluateOutboxRetryIsDue,
	"RetryDoesNotRepeatCommittedStep": policy.EvaluateRetryDoesNotRepeatCommittedStep,
	"QueryFilterAndCountMatch":        policy.EvaluateQueryFilterAndCountMatch,
	"BatchReportsPartialFailure":      policy.EvaluateBatchReportsPartialFailure,
	"AuditBelongsToTransaction":       policy.EvaluateAuditBelongsToTransaction,
	"AuditSequenceIsStable":           policy.EvaluateAuditSequenceIsStable,
	"TenantBoundaryIsPreserved":       policy.EvaluateTenantBoundaryIsPreserved,
	"ContextDeadlineIsForwarded":      policy.EvaluateContextDeadlineIsForwarded,
	"RepositoryResultIsIsolated":      policy.EvaluateRepositoryResultIsIsolated,
	"RecoveredMetadataIsInitialized":  policy.EvaluateRecoveredMetadataIsInitialized,
	"WorkerStopsOnCancellation":       policy.EvaluateWorkerStopsOnCancellation,
	"ResourceReleaseIsReported":       policy.EvaluateResourceReleaseIsReported,
	"DispatchPriorityIsStable":        policy.EvaluateDispatchPriorityIsStable,
}

func Rules() []string {
	result := make([]string, 0, len(evaluators))
	for name := range evaluators {
		result = append(result, name)
	}
	sortStrings(result)
	return result
}

func (c *Coordinator) Evaluate(ctx context.Context, request DecisionRequest) (DecisionResponse, error) {
	evaluator, ok := evaluators[request.Rule]
	if !ok {
		return DecisionResponse{}, fault.New(fault.CodeNotFound, "coordinator.evaluate", "rule was not found")
	}
	if request.Actor == "" || request.RequestID == "" {
		return DecisionResponse{}, fault.New(fault.CodeInvalid, "coordinator.evaluate", "actor and request ID are required")
	}
	result, err := evaluator(request.Context)
	if err != nil {
		return DecisionResponse{}, err
	}
	now := c.Clock.Now().UTC()
	auditID := c.IDs.New("audit")
	outboxID := ""
	err = c.DB.WithTx(ctx, func(tx *sql.Tx) error {
		detail, _ := json.Marshal(map[string]any{"rule": request.Rule, "allowed": result.Allowed, "code": result.Code})
		event := audit.Event{ID: auditID, TenantID: request.Context.TenantID, Actor: request.Actor, Action: "policy.evaluate", ObjectType: "policy", ObjectID: request.Rule, RequestID: request.RequestID, Detail: string(detail), Status: audit.StatusRecorded, Version: 1, CreatedAt: now, UpdatedAt: now}
		raw, _ := json.Marshal(event)
		if _, err := tx.ExecContext(ctx, `INSERT INTO audit_events(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, event.ID, event.TenantID, raw, event.Status, event.Version, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("write audit: %w", err)
		}
		if result.Allowed {
			outboxID = c.IDs.New("job")
			job := outbox.Job{ID: outboxID, TenantID: request.Context.TenantID, Topic: "policy.allowed", ObjectID: request.Rule, Payload: string(detail), Attempts: 0, AvailableAt: now, Status: outbox.StatusPending, Version: 1, CreatedAt: now, UpdatedAt: now}
			payload, _ := json.Marshal(job)
			if _, err := tx.ExecContext(ctx, `INSERT INTO outbox_jobs(id,tenant_id,data_json,status,version,created_at,updated_at) VALUES(?,?,?,?,?,?,?)`, job.ID, job.TenantID, payload, job.Status, job.Version, now.Format(time.RFC3339Nano), now.Format(time.RFC3339Nano)); err != nil {
				return fmt.Errorf("write outbox: %w", err)
			}
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO request_log(request_id,tenant_id,method,path,outcome,created_at) VALUES(?,?,?,?,?,?)`, request.RequestID, request.Context.TenantID, "POST", "/api/decisions", result.Code, now.Format(time.RFC3339Nano)); err != nil {
			return fmt.Errorf("write request log: %w", err)
		}
		return nil
	})
	if err != nil {
		return DecisionResponse{}, fault.Wrap(fault.CodeInternal, "coordinator.evaluate", "persist decision", err)
	}
	return DecisionResponse{Rule: request.Rule, Result: result, AuditID: auditID, OutboxID: outboxID}, nil
}

func (c *Coordinator) Overview(ctx context.Context, tenant string) (map[string]int, error) {
	tables := []string{"farms", "turbines", "weather_windows", "campaigns", "crew_members", "vessels", "permits", "work_orders", "dispatches", "evidence", "outbox_jobs", "audit_events"}
	result := make(map[string]int, len(tables))
	for _, table := range tables {
		var count int
		query := "SELECT COUNT(*) FROM " + table + " WHERE tenant_id=?"
		if err := c.DB.SQL().QueryRowContext(ctx, query, tenant).Scan(&count); err != nil {
			return nil, err
		}
		result[table] = count
	}
	return result, nil
}

func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
