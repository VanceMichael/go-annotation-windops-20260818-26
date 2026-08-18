package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCrewMemberIsQualified(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.crew_qualification"); err != nil {
		return Result{}, err
	}
	if ctx.ActorID == "" || ctx.Kind == "" {
		return Result{}, fault.New(fault.CodeInvalid, "policy.crew_qualification", "member and qualification kind are required")
	}
	if ctx.Status != "valid" {
		return deny("qualification_inactive", "required qualification is not active"), nil
	}
	if ctx.Now.Before(ctx.ValidFrom) || !ctx.Now.Before(ctx.ValidUntil) {
		return deny("qualification_outside_validity", "qualification is not valid at departure time"), nil
	}
	if !ctx.Flags["medical_clearance"] {
		return deny("medical_clearance", "medical clearance is missing"), nil
	}
	return allow("crew member holds a valid qualification"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
