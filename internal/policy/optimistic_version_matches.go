package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateOptimisticVersionMatches(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.optimistic_version"); err != nil {
		return Result{}, err
	}
	if ctx.Version < 1 || ctx.ExpectedVersion < 1 {
		return Result{}, fault.New(fault.CodeInvalid, "policy.optimistic_version", "versions must be positive")
	}
	if ctx.Version != ctx.ExpectedVersion {
		result := deny("version_conflict", "record changed since it was read")
		result.Quantity = ctx.Version
		return result, nil
	}
	if ctx.Flags["deleted"] {
		return deny("record_deleted", "record was deleted before update"), nil
	}
	result := allow("record version matches")
	result.Quantity = ctx.Version + 1
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
