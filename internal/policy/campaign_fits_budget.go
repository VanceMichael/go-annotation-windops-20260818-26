package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateCampaignFitsBudget(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.campaign_budget"); err != nil {
		return Result{}, err
	}
	if ctx.Budget < 0 || ctx.Cost < 0 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.campaign_budget", "budget and cost cannot be negative")
	}
	total := ctx.Cost
	for _, value := range ctx.Values {
		if value < 0 {
			return Result{}, fault.New(fault.CodeInvalid, "policy.campaign_budget", "line cost cannot be negative")
		}
		total += value
	}
	if total > ctx.Budget {
		result := deny("budget_exceeded", "campaign cost exceeds approved budget")
		result.Quantity = total - ctx.Budget
		return result, nil
	}
	result := allow("campaign is within approved budget")
	result.Quantity = ctx.Budget - total
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
