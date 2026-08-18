package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateRetryDoesNotRepeatCommittedStep(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.retry_checkpoint"); err != nil {
		return Result{}, err
	}
	completed := cloneMap(ctx.Metadata)
	steps := []string{"reserve", "notify", "audit"}
	pending := []string{}
	for _, step := range steps {
		if completed[step] != "committed" {
			pending = append(pending, step)
		}
	}
	if len(pending) == 0 {
		return deny("nothing_to_retry", "all workflow steps are already committed"), nil
	}
	if ctx.Attempts >= ctx.MaxAttempts {
		return deny("retry_exhausted", "workflow exhausted retry attempts"), nil
	}
	result := allow("workflow can resume from uncommitted steps")
	result.IDs = pending
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
