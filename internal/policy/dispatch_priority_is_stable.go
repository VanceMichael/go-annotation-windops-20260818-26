package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateDispatchPriorityIsStable(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.dispatch_priority"); err != nil {
		return Result{}, err
	}
	if len(ctx.Values) != len(ctx.ExistingIDs) {
		return Result{}, fault.New(fault.CodeInvalid, "policy.dispatch_priority", "priority values and IDs must align")
	}
	ids := cloneStrings(ctx.ExistingIDs)
	values := append([]int64(nil), ctx.Values...)
	for i := 1; i < len(ids); i++ {
		for j := i; j > 0 && (values[j] > values[j-1] || (values[j] == values[j-1] && ids[j] < ids[j-1])); j-- {
			values[j], values[j-1] = values[j-1], values[j]
			ids[j], ids[j-1] = ids[j-1], ids[j]
		}
	}
	result := allow("dispatch priority order is stable")
	result.IDs = ids
	if len(values) > 0 {
		result.Quantity = values[0]
	}
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
