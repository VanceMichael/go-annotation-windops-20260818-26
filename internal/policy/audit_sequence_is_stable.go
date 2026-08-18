package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateAuditSequenceIsStable(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.audit_sequence"); err != nil {
		return Result{}, err
	}
	if len(ctx.Values) != len(ctx.ExistingIDs) {
		return Result{}, fault.New(fault.CodeInvalid, "policy.audit_sequence", "audit sequences and IDs must align")
	}
	previous := int64(-1)
	seen := map[string]bool{}
	for i, sequence := range ctx.Values {
		if sequence <= previous {
			return deny("audit_order", "audit sequence is not strictly increasing"), nil
		}
		if seen[ctx.ExistingIDs[i]] {
			return deny("audit_duplicate", "audit event appears more than once"), nil
		}
		seen[ctx.ExistingIDs[i]] = true
		previous = sequence
	}
	result := allow("audit sequence is stable")
	result.IDs = cloneStrings(ctx.ExistingIDs)
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
