package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateDispatchFitsVessel(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.dispatch_capacity"); err != nil {
		return Result{}, err
	}
	if ctx.Requested < 0 || ctx.Used < 0 || ctx.Capacity < 0 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.dispatch_capacity", "capacity cannot be negative")
	}
	total := ctx.Used + ctx.Requested
	for _, value := range ctx.Values {
		if value < 0 {
			return Result{}, fault.New(fault.CodeInvalid, "policy.dispatch_capacity", "manifest quantity cannot be negative")
		}
		total += value
	}
	if total > ctx.Capacity {
		result := deny("dispatch_over_capacity", "dispatch manifest exceeds vessel capacity")
		result.Quantity = total - ctx.Capacity
		return result, nil
	}
	result := allow("dispatch manifest fits vessel capacity")
	result.Quantity = ctx.Capacity - total
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
