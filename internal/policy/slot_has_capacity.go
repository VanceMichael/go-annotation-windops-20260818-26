package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateSlotHasCapacity(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.slot_capacity"); err != nil {
		return Result{}, err
	}
	if ctx.Capacity < 0 || ctx.Used < 0 || ctx.Requested <= 0 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.slot_capacity", "capacity values are invalid")
	}
	if ctx.Used > ctx.Capacity {
		return Result{}, fault.New(fault.CodePrecondition, "policy.slot_capacity", "stored usage already exceeds capacity")
	}
	remaining := ctx.Capacity - ctx.Used
	if ctx.Requested > remaining {
		result := deny("slot_full", "maintenance slot does not have enough capacity")
		result.Quantity = remaining
		return result, nil
	}
	result := allow("maintenance slot has capacity")
	result.Quantity = remaining - ctx.Requested
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
