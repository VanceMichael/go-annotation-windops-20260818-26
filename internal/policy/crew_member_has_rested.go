package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCrewMemberHasRested(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.crew_rest"); err != nil {
		return Result{}, err
	}
	if ctx.RestMinutes < 0 || ctx.PlannedMinutes <= 0 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.crew_rest", "shift duration values are invalid")
	}
	minimum := 660
	if ctx.Flags["night_shift"] {
		minimum = 720
	}
	if ctx.RestMinutes < minimum {
		result := deny("crew_rest", "crew member has not completed minimum rest")
		result.Quantity = int64(minimum - ctx.RestMinutes)
		return result, nil
	}
	if ctx.PlannedMinutes > 720 {
		return deny("shift_too_long", "planned offshore shift exceeds twelve hours"), nil
	}
	return allow("crew rest and planned shift are valid"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
