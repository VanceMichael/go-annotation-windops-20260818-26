package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCampaignCanClose(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.campaign_close"); err != nil {
		return Result{}, err
	}
	if ctx.Status != "in_progress" {
		return deny("campaign_state", "only in-progress campaigns can close"), nil
	}
	required := []string{"all_work_terminal", "dispatch_released", "permit_closed", "audit_complete"}
	missing := []string{}
	for _, key := range required {
		if !ctx.Flags[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		result := deny("campaign_open_dependencies", "campaign still has open dependencies")
		result.IDs = missing
		return result, nil
	}
	if ctx.Cost > ctx.Budget {
		return deny("campaign_cost_unsettled", "campaign actual cost exceeds budget without approval"), nil
	}
	return allow("campaign can close"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
