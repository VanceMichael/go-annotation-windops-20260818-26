package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateEvidenceChecksumIsUnique(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.evidence_checksum"); err != nil {
		return Result{}, err
	}
	if ctx.PayloadHash == "" {
		return Result{}, fault.New(fault.CodeInvalid, "policy.evidence_checksum", "checksum is required")
	}
	matches := []string{}
	for i, label := range ctx.Labels {
		if label == ctx.PayloadHash && i < len(ctx.ExistingIDs) {
			matches = append(matches, ctx.ExistingIDs[i])
		}
	}
	if len(matches) > 0 {
		result := deny("evidence_duplicate", "evidence checksum is already registered")
		result.IDs = matches
		return result, nil
	}
	if len(ctx.PayloadHash) < 32 {
		return deny("checksum_invalid", "evidence checksum is too short"), nil
	}
	return allow("evidence checksum is unique"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
