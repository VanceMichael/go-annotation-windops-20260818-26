package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateAuditBelongsToTransaction(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.audit_transaction"); err != nil {
		return Result{}, err
	}
	required := []string{"business_written", "audit_written", "same_transaction"}
	missing := []string{}
	for _, key := range required {
		if !ctx.Flags[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		result := deny("audit_atomicity", "business mutation and audit event are not atomic")
		result.IDs = missing
		return result, nil
	}
	if ctx.ResourceID == "" || ctx.ActorID == "" {
		return Result{}, fault.New(fault.CodeInvalid, "policy.audit_transaction", "resource and actor are required")
	}
	return allow("business mutation and audit event are atomic"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
