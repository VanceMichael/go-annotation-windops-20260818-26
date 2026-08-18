package policy

import (
	"testing"
	"time"
)

func baseline() Context {
	now := time.Date(2026, 8, 18, 9, 0, 0, 0, time.UTC)
	return Context{TenantID: "tenant-a", ResourceID: "resource-1", RelatedID: "tenant-a", ActorID: "crew-1", Status: "planned", RelatedStatus: "planned", Kind: "offshore-lifting", Region: "east", Method: "POST", Path: "/api/campaigns", Key: "key-1", PayloadHash: "0123456789abcdef0123456789abcdef", ExpectedHash: "0123456789abcdef0123456789abcdef", StartsAt: now.Add(time.Hour), EndsAt: now.Add(6 * time.Hour), Now: now, Deadline: now.Add(4 * time.Hour), ValidFrom: now.AddDate(-1, 0, 0), ValidUntil: now.AddDate(1, 0, 0), ExistingStarts: []time.Time{}, ExistingEnds: []time.Time{}, ExistingIDs: []string{}, Values: []int64{}, Labels: []string{}, Capacity: 100, Used: 20, Requested: 10, Budget: 1000, Cost: 100, Version: 2, ExpectedVersion: 2, Attempts: 1, MaxAttempts: 5, Confidence: 90, WindKPH: 40, WaveCM: 150, RestMinutes: 720, PlannedMinutes: 480, Flags: map[string]bool{"window_confirmed": true, "slots_booked": true, "budget_reviewed": true, "risk_assessed": true, "inspection_valid": true, "insurance_valid": true, "medical_clearance": true, "campaign_approved": true, "vessel_reserved": true, "crew_complete": true, "vessel_ready": true, "crew_checked_in": true, "permit_active": true, "crew_on_site": true, "turbine_isolated": true, "before_photo": true, "after_photo": true, "torque_record": true, "technician_signoff": true, "all_work_terminal": true, "dispatch_released": true, "permit_closed": true, "audit_complete": true, "business_written": true, "audit_written": true, "same_transaction": true, "cancel_propagates": true, "context_done": true, "poll_stopped": true, "inflight_joined": true, "operation_attempted": true, "resource_closed": true}, Metadata: map[string]string{"tenant": "tenant-a", "method": "POST", "path": "/api/campaigns", "key": "key-1", "count_region": "east", "audit_tenant": "tenant-a", "request_id": "request-1"}}
}

func expectAllowed(t *testing.T, result Result, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Allowed {
		t.Fatalf("expected allowed, got %+v", result)
	}
}
func expectDenied(t *testing.T, result Result, err error, code string) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Allowed || result.Code != code {
		t.Fatalf("expected denial %s, got %+v", code, result)
	}
}

type checker struct {
	t    *testing.T
	code string
}

func check(t *testing.T) checker                    { return checker{t: t} }
func checkDenied(t *testing.T, code string) checker { return checker{t: t, code: code} }
func (c checker) allowed(result Result, err error)  { expectAllowed(c.t, result, err) }
func (c checker) denied(result Result, err error)   { expectDenied(c.t, result, err, c.code) }

func TestWeatherWindowSafetyRejectsMarineLimit(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateWeatherWindowIsSafe(ctx))
	ctx.WindKPH = 71
	checkDenied(t, "unsafe_weather").denied(EvaluateWeatherWindowIsSafe(ctx))
}
func TestWeatherWindowOverlapReportsConflictingIDs(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateWeatherWindowDoesNotOverlap(ctx))
	ctx.ExistingStarts = []time.Time{ctx.StartsAt.Add(time.Minute)}
	ctx.ExistingEnds = []time.Time{ctx.EndsAt}
	ctx.ExistingIDs = []string{"window-2"}
	result, err := EvaluateWeatherWindowDoesNotOverlap(ctx)
	expectDenied(t, result, err, "window_conflict")
	if len(result.IDs) != 1 || result.IDs[0] != "window-2" {
		t.Fatalf("missing conflict: %+v", result)
	}
}
func TestCampaignBudgetIncludesLineCosts(t *testing.T) {
	ctx := baseline()
	ctx.Values = []int64{300, 400}
	check(t).allowed(EvaluateCampaignFitsBudget(ctx))
	ctx.Values = append(ctx.Values, 300)
	checkDenied(t, "budget_exceeded").denied(EvaluateCampaignFitsBudget(ctx))
}
func TestSlotCapacityAcceptsExactFit(t *testing.T) {
	ctx := baseline()
	ctx.Capacity = 30
	ctx.Used = 20
	ctx.Requested = 10
	check(t).allowed(EvaluateSlotHasCapacity(ctx))
	ctx.Requested = 11
	checkDenied(t, "slot_full").denied(EvaluateSlotHasCapacity(ctx))
}
func TestCampaignApprovalRequiresAllPrerequisites(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateCampaignCanBeApproved(ctx))
	ctx.Flags["risk_assessed"] = false
	checkDenied(t, "campaign_incomplete").denied(EvaluateCampaignCanBeApproved(ctx))
}
func TestVesselReservationChecksDocuments(t *testing.T) {
	ctx := baseline()
	ctx.Status = "available"
	check(t).allowed(EvaluateVesselCanBeReserved(ctx))
	ctx.Flags["insurance_valid"] = false
	checkDenied(t, "vessel_documents").denied(EvaluateVesselCanBeReserved(ctx))
}
func TestCrewQualificationUsesHalfOpenValidity(t *testing.T) {
	ctx := baseline()
	ctx.Status = "valid"
	check(t).allowed(EvaluateCrewMemberIsQualified(ctx))
	ctx.Now = ctx.ValidUntil
	checkDenied(t, "qualification_outside_validity").denied(EvaluateCrewMemberIsQualified(ctx))
}
func TestCrewRestUsesNightShiftMinimum(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateCrewMemberHasRested(ctx))
	ctx.Flags["night_shift"] = true
	ctx.RestMinutes = 719
	checkDenied(t, "crew_rest").denied(EvaluateCrewMemberHasRested(ctx))
}
func TestCrewAssignmentsRejectOverlap(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateCrewAssignmentsDoNotOverlap(ctx))
	ctx.ExistingStarts = []time.Time{ctx.StartsAt}
	ctx.ExistingEnds = []time.Time{ctx.EndsAt}
	ctx.ExistingIDs = []string{"dispatch-2"}
	checkDenied(t, "crew_double_booked").denied(EvaluateCrewAssignmentsDoNotOverlap(ctx))
}
func TestPermitSubmissionRequiresCompleteManifest(t *testing.T) {
	ctx := baseline()
	ctx.Status = "draft"
	check(t).allowed(EvaluatePermitCanBeSubmitted(ctx))
	ctx.Flags["crew_complete"] = false
	checkDenied(t, "permit_incomplete").denied(EvaluatePermitCanBeSubmitted(ctx))
}
func TestPermitActivationOnlyInsideWindow(t *testing.T) {
	ctx := baseline()
	ctx.Status = "approved"
	ctx.Now = ctx.StartsAt
	check(t).allowed(EvaluatePermitCanBeActivated(ctx))
	ctx.Now = ctx.EndsAt
	checkDenied(t, "permit_window").denied(EvaluatePermitCanBeActivated(ctx))
}
func TestPermitExpiryCannotExceedWindow(t *testing.T) {
	ctx := baseline()
	ctx.Deadline = ctx.EndsAt
	check(t).allowed(EvaluatePermitExpiryIsBounded(ctx))
	ctx.Deadline = ctx.EndsAt.Add(time.Second)
	checkDenied(t, "permit_expiry").denied(EvaluatePermitExpiryIsBounded(ctx))
}
func TestDispatchCapacityIncludesManifestLines(t *testing.T) {
	ctx := baseline()
	ctx.Values = []int64{30, 40}
	check(t).allowed(EvaluateDispatchFitsVessel(ctx))
	ctx.Values = append(ctx.Values, 1)
	checkDenied(t, "dispatch_over_capacity").denied(EvaluateDispatchFitsVessel(ctx))
}
func TestDispatchDepartureRejectsOtherIDs(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateDispatchDepartureIsUnique(ctx))
	ctx.ExistingIDs = []string{"dispatch-2"}
	checkDenied(t, "duplicate_departure").denied(EvaluateDispatchDepartureIsUnique(ctx))
}
func TestWorkStartRequiresIsolation(t *testing.T) {
	ctx := baseline()
	ctx.Status = "assigned"
	check(t).allowed(EvaluateWorkOrderCanStart(ctx))
	ctx.Flags["turbine_isolated"] = false
	checkDenied(t, "work_preconditions").denied(EvaluateWorkOrderCanStart(ctx))
}
func TestWorkCompletionRequiresEvidence(t *testing.T) {
	ctx := baseline()
	ctx.Status = "awaiting_evidence"
	check(t).allowed(EvaluateWorkOrderCanComplete(ctx))
	ctx.Flags["torque_record"] = false
	checkDenied(t, "evidence_incomplete").denied(EvaluateWorkOrderCanComplete(ctx))
}
func TestEvidenceChecksumRejectsDuplicate(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateEvidenceChecksumIsUnique(ctx))
	ctx.Labels = []string{ctx.PayloadHash}
	ctx.ExistingIDs = []string{"evidence-2"}
	checkDenied(t, "evidence_duplicate").denied(EvaluateEvidenceChecksumIsUnique(ctx))
}
func TestCampaignCloseRequiresTerminalDependencies(t *testing.T) {
	ctx := baseline()
	ctx.Status = "in_progress"
	check(t).allowed(EvaluateCampaignCanClose(ctx))
	ctx.Flags["dispatch_released"] = false
	checkDenied(t, "campaign_open_dependencies").denied(EvaluateCampaignCanClose(ctx))
}
func TestIdempotencyScopeIncludesTenantMethodAndPath(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateIdempotencyScopeMatches(ctx))
	ctx.Metadata["path"] = "/api/permits"
	checkDenied(t, "idempotency_scope_mismatch").denied(EvaluateIdempotencyScopeMatches(ctx))
}
func TestOptimisticVersionRejectsStaleWriter(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateOptimisticVersionMatches(ctx))
	ctx.ExpectedVersion--
	checkDenied(t, "version_conflict").denied(EvaluateOptimisticVersionMatches(ctx))
}
func TestOutboxRetryStopsAtMaximum(t *testing.T) {
	ctx := baseline()
	ctx.Now = ctx.Deadline
	check(t).allowed(EvaluateOutboxRetryIsDue(ctx))
	ctx.Attempts = ctx.MaxAttempts
	checkDenied(t, "permanent_failure").denied(EvaluateOutboxRetryIsDue(ctx))
}
func TestRetryOnlyReturnsUncommittedSteps(t *testing.T) {
	ctx := baseline()
	ctx.Metadata = map[string]string{"reserve": "committed"}
	result, err := EvaluateRetryDoesNotRepeatCommittedStep(ctx)
	expectAllowed(t, result, err)
	if len(result.IDs) != 2 || result.IDs[0] != "notify" {
		t.Fatalf("unexpected pending steps: %+v", result.IDs)
	}
}
func TestQueryCountRequiresSameFilters(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateQueryFilterAndCountMatch(ctx))
	ctx.Metadata["count_region"] = "west"
	checkDenied(t, "query_filter_mismatch").denied(EvaluateQueryFilterAndCountMatch(ctx))
}
func TestBatchReturnsFailedItemIDs(t *testing.T) {
	ctx := baseline()
	ctx.ExistingIDs = []string{"a", "b", "c"}
	ctx.Labels = []string{"ok", "failed", "ok"}
	result, err := EvaluateBatchReportsPartialFailure(ctx)
	expectAllowed(t, result, err)
	if result.Code != "partial_failure" || len(result.IDs) != 1 || result.IDs[0] != "b" {
		t.Fatalf("unexpected batch result: %+v", result)
	}
}
func TestAuditMustShareBusinessTransaction(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateAuditBelongsToTransaction(ctx))
	ctx.Flags["same_transaction"] = false
	checkDenied(t, "audit_atomicity").denied(EvaluateAuditBelongsToTransaction(ctx))
}
func TestAuditSequenceIsStrictlyIncreasing(t *testing.T) {
	ctx := baseline()
	ctx.Values = []int64{1, 2, 3}
	ctx.ExistingIDs = []string{"a", "b", "c"}
	check(t).allowed(EvaluateAuditSequenceIsStable(ctx))
	ctx.Values[2] = 2
	checkDenied(t, "audit_order").denied(EvaluateAuditSequenceIsStable(ctx))
}
func TestTenantBoundaryRejectsMixedResults(t *testing.T) {
	ctx := baseline()
	ctx.Labels = []string{"tenant-a", "tenant-a"}
	check(t).allowed(EvaluateTenantBoundaryIsPreserved(ctx))
	ctx.Labels[1] = "tenant-b"
	checkDenied(t, "tenant_boundary").denied(EvaluateTenantBoundaryIsPreserved(ctx))
}
func TestContextDeadlineRequiresCancellationPropagation(t *testing.T) {
	ctx := baseline()
	ctx.Flags["cancel_propagates"] = false
	checkDenied(t, "cancel_not_forwarded").denied(EvaluateContextDeadlineIsForwarded(ctx))
}
func TestRepositoryResultClonesSlicesAndMaps(t *testing.T) {
	ctx := baseline()
	ctx.ExistingIDs = []string{"first", "second"}
	check(t).allowed(EvaluateRepositoryResultIsIsolated(ctx))
	if ctx.ExistingIDs[0] != "first" {
		t.Fatalf("caller input mutated: %+v", ctx.ExistingIDs)
	}
}
func TestRecoveredMetadataAddsDefaults(t *testing.T) {
	ctx := baseline()
	ctx.Metadata = map[string]string{}
	result, err := EvaluateRecoveredMetadataIsInitialized(ctx)
	expectAllowed(t, result, err)
	if len(result.IDs) != 3 {
		t.Fatalf("missing defaults: %+v", result)
	}
}
func TestWorkerCancellationJoinsInflightWork(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateWorkerStopsOnCancellation(ctx))
	ctx.Flags["inflight_joined"] = false
	checkDenied(t, "worker_shutdown_incomplete").denied(EvaluateWorkerStopsOnCancellation(ctx))
}
func TestResourceReleaseReturnsCloseFailure(t *testing.T) {
	ctx := baseline()
	check(t).allowed(EvaluateResourceReleaseIsReported(ctx))
	ctx.Flags["close_failed"] = true
	checkDenied(t, "close_error_lost").denied(EvaluateResourceReleaseIsReported(ctx))
}
func TestDispatchPriorityUsesIDTieBreak(t *testing.T) {
	ctx := baseline()
	ctx.Values = []int64{5, 8, 8}
	ctx.ExistingIDs = []string{"z", "b", "a"}
	result, err := EvaluateDispatchPriorityIsStable(ctx)
	expectAllowed(t, result, err)
	if result.IDs[0] != "a" || result.IDs[1] != "b" {
		t.Fatalf("unstable order: %+v", result.IDs)
	}
}
