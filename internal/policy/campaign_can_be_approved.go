package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCampaignCanBeApproved(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.campaign_approval"); err != nil {
		return Result{}, err
	}
	required := []string{"window_confirmed", "slots_booked", "budget_reviewed", "risk_assessed"}
	missing := make([]string, 0)
	for _, key := range required {
		if !ctx.Flags[key] {
			missing = append(missing, key)
		}
	}
	if ctx.Status != "planned" {
		return deny("campaign_state", "only planned campaigns can be approved"), nil
	}
	if len(missing) > 0 {
		result := deny("campaign_incomplete", "campaign prerequisites are incomplete")
		result.IDs = missing
		return result, nil
	}
	return allow("campaign can be approved"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
