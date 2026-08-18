package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluatePermitCanBeSubmitted(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.permit_submit"); err != nil {
		return Result{}, err
	}
	if ctx.Status != "draft" {
		return deny("permit_state", "only draft permits can be submitted"), nil
	}
	required := []string{"campaign_approved", "vessel_reserved", "crew_complete", "window_confirmed"}
	missing := []string{}
	for _, key := range required {
		if !ctx.Flags[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		result := deny("permit_incomplete", "permit prerequisites are incomplete")
		result.IDs = missing
		return result, nil
	}
	return allow("permit can be submitted"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
