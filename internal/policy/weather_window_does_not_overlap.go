package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateWeatherWindowDoesNotOverlap(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.window_overlap"); err != nil {
		return Result{}, err
	}
	if len(ctx.ExistingStarts) != len(ctx.ExistingEnds) || len(ctx.ExistingStarts) != len(ctx.ExistingIDs) {
		return Result{}, fault.New(fault.CodeInvalid, "policy.window_overlap", "existing window vectors must align")
	}
	conflicts := make([]string, 0)
	for i := range ctx.ExistingStarts {
		if overlap(ctx.StartsAt, ctx.EndsAt, ctx.ExistingStarts[i], ctx.ExistingEnds[i]) {
			conflicts = append(conflicts, ctx.ExistingIDs[i])
		}
	}
	if len(conflicts) > 0 {
		result := deny("window_conflict", "confirmed weather windows overlap")
		result.IDs = cloneStrings(conflicts)
		return result, nil
	}
	return allow("weather window does not overlap another confirmed window"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
