package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateDispatchDepartureIsUnique(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.dispatch_unique"); err != nil {
		return Result{}, err
	}
	if ctx.ResourceID == "" {
		return Result{}, fault.New(fault.CodeInvalid, "policy.dispatch_unique", "dispatch ID is required")
	}
	duplicates := []string{}
	for _, id := range ctx.ExistingIDs {
		if id != ctx.ResourceID {
			duplicates = append(duplicates, id)
		}
	}
	if len(duplicates) > 0 {
		result := deny("duplicate_departure", "vessel already has another departure in the window")
		result.IDs = cloneStrings(duplicates)
		return result, nil
	}
	return allow("vessel departure is unique in the window"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
