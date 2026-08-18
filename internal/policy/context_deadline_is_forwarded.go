package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateContextDeadlineIsForwarded(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.context_deadline"); err != nil {
		return Result{}, err
	}
	if ctx.Deadline.IsZero() {
		return deny("deadline_missing", "downstream operation did not receive a deadline"), nil
	}
	if ctx.Now.After(ctx.Deadline) || ctx.Now.Equal(ctx.Deadline) {
		return deny("deadline_exceeded", "downstream deadline has expired"), nil
	}
	if ctx.Metadata["request_id"] == "" {
		return deny("request_context_missing", "downstream context lost request metadata"), nil
	}
	if !cancellationForwarded(ctx.Flags, !ctx.Deadline.IsZero()) {
		return deny("cancel_not_forwarded", "downstream cancellation is detached"), nil
	}
	result := allow("deadline and cancellation are forwarded")
	result.RetryAt = ctx.Deadline
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
