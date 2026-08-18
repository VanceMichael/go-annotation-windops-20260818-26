package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateOutboxRetryIsDue(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.outbox_retry"); err != nil {
		return Result{}, err
	}
	if ctx.Attempts < 0 || ctx.MaxAttempts < 1 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.outbox_retry", "attempt counts are invalid")
	}
	if ctx.Attempts >= ctx.MaxAttempts {
		return deny("permanent_failure", "outbox job exhausted retry attempts"), nil
	}
	if ctx.Now.Before(ctx.Deadline) {
		result := deny("retry_not_due", "outbox retry is not due")
		result.RetryAt = ctx.Deadline
		return result, nil
	}
	delay := time.Duration(1<<min(ctx.Attempts, 8)) * time.Second
	result := allow("outbox retry is due")
	result.RetryAt = ctx.Now.Add(delay)
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
