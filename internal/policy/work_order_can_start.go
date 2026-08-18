package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateWorkOrderCanStart(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.work_start"); err != nil {
		return Result{}, err
	}
	if ctx.Status != "assigned" {
		return deny("work_state", "only assigned work can start"), nil
	}
	if !ctx.Flags["permit_active"] || !ctx.Flags["crew_on_site"] || !ctx.Flags["turbine_isolated"] {
		return deny("work_preconditions", "permit, crew arrival and turbine isolation are required"), nil
	}
	if ctx.Flags["stop_work"] {
		return deny("stop_work", "a stop-work instruction is active"), nil
	}
	if !ctx.Deadline.IsZero() && !ctx.Now.Before(ctx.Deadline) {
		return deny("work_window_closed", "work cannot start after permit expiry"), nil
	}
	return allow("work order can start"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
