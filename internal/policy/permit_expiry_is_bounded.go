package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluatePermitExpiryIsBounded(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.permit_expiry"); err != nil {
		return Result{}, err
	}
	if ctx.StartsAt.IsZero() || ctx.EndsAt.IsZero() || ctx.Deadline.IsZero() {
		return Result{}, fault.New(fault.CodeInvalid, "policy.permit_expiry", "window and expiry are required")
	}
	if ctx.Deadline.After(ctx.EndsAt) {
		return deny("permit_expiry", "permit expiry cannot exceed weather window"), nil
	}
	if ctx.Deadline.Before(ctx.StartsAt) {
		return deny("permit_expiry", "permit expiry cannot precede weather window"), nil
	}
	if ctx.Deadline.Sub(ctx.StartsAt) > 12*time.Hour {
		return deny("permit_duration", "permit duration cannot exceed twelve hours"), nil
	}
	return allow("permit expiry is bounded by the operating window"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
