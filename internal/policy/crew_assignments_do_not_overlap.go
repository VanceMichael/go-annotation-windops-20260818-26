package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCrewAssignmentsDoNotOverlap(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.crew_overlap"); err != nil {
		return Result{}, err
	}
	if len(ctx.ExistingStarts) != len(ctx.ExistingEnds) || len(ctx.ExistingStarts) != len(ctx.ExistingIDs) {
		return Result{}, fault.New(fault.CodeInvalid, "policy.crew_overlap", "assignment vectors must align")
	}
	conflicts := []string{}
	for i := range ctx.ExistingIDs {
		if overlap(ctx.StartsAt, ctx.EndsAt, ctx.ExistingStarts[i], ctx.ExistingEnds[i]) {
			conflicts = append(conflicts, ctx.ExistingIDs[i])
		}
	}
	if len(conflicts) > 0 {
		result := deny("crew_double_booked", "crew member already has an overlapping assignment")
		result.IDs = conflicts
		return result, nil
	}
	return allow("crew assignment does not overlap"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
