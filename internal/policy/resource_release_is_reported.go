package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateResourceReleaseIsReported(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.resource_release"); err != nil {
		return Result{}, err
	}
	if !ctx.Flags["operation_attempted"] {
		return Result{}, fault.New(fault.CodeInvalid, "policy.resource_release", "operation was not attempted")
	}
	if ctx.Flags["operation_failed"] {
		if !ctx.Flags["rollback_completed"] {
			return deny("rollback_failed", "failed operation did not release its reservation"), nil
		}
		if ctx.Flags["success_reported"] {
			return deny("false_success", "failed operation was reported as successful"), nil
		}
	}
	if ctx.Flags["close_failed"] && !ctx.Flags["close_error_returned"] {
		return deny("close_error_lost", "resource close failure was not returned"), nil
	}
	if !ctx.Flags["resource_closed"] {
		return deny("resource_leak", "resource remains open"), nil
	}
	return allow("resource release and failure reporting are complete"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
