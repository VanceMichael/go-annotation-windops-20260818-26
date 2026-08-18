package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateRepositoryResultIsIsolated(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.repository_isolation"); err != nil {
		return Result{}, err
	}
	first := cloneStrings(ctx.ExistingIDs)
	second := cloneStrings(ctx.ExistingIDs)
	if len(first) > 0 {
		first[0] = "caller-mutated"
	}
	if len(first) > 0 && second[0] == first[0] {
		return deny("shared_result", "repository results share mutable storage"), nil
	}
	metadata := cloneMap(ctx.Metadata)
	metadata["caller"] = "mutated"
	if ctx.Metadata["caller"] == "mutated" {
		return deny("shared_metadata", "repository metadata was returned by reference"), nil
	}
	result := allow("repository result is isolated from caller mutation")
	result.IDs = second
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
