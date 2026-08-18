package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateBatchReportsPartialFailure(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.batch_result"); err != nil {
		return Result{}, err
	}
	if len(ctx.ExistingIDs) != len(ctx.Labels) {
		return Result{}, fault.New(fault.CodeInvalid, "policy.batch_result", "batch IDs and outcomes must align")
	}
	failed := []string{}
	succeeded := int64(0)
	for i, outcome := range ctx.Labels {
		switch outcome {
		case "ok":
			succeeded++
		case "failed":
			failed = append(failed, ctx.ExistingIDs[i])
		default:
			return Result{}, fault.New(fault.CodeInvalid, "policy.batch_result", "unknown batch outcome")
		}
	}
	result := allow("batch completed")
	result.Quantity = succeeded
	result.IDs = failed
	if len(failed) > 0 {
		result.Code = "partial_failure"
		result.Message = "batch completed with item failures"
	}
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
