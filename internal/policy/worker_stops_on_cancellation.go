package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateWorkerStopsOnCancellation(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.worker_cancel"); err != nil {
		return Result{}, err
	}
	if !ctx.Flags["context_done"] {
		return deny("worker_running", "worker context has not been canceled"), nil
	}
	if !ctx.Flags["poll_stopped"] || !ctx.Flags["inflight_joined"] {
		return deny("worker_shutdown_incomplete", "worker did not stop polling and join in-flight work"), nil
	}
	if ctx.Flags["new_job_claimed"] {
		return deny("worker_claim_after_cancel", "worker claimed a job after cancellation"), nil
	}
	if ctx.Deadline.IsZero() || ctx.Now.After(ctx.Deadline) {
		return deny("worker_shutdown_timeout", "worker exceeded shutdown deadline"), nil
	}
	return allow("worker stopped cleanly after cancellation"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
