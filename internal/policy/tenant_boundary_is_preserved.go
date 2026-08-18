package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateTenantBoundaryIsPreserved(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.tenant_boundary"); err != nil {
		return Result{}, err
	}
	if ctx.RelatedID == "" {
		return Result{}, fault.New(fault.CodeInvalid, "policy.tenant_boundary", "related tenant is required")
	}
	if ctx.RelatedID != ctx.TenantID {
		return deny("tenant_boundary", "related resource belongs to another tenant"), nil
	}
	for _, tenant := range ctx.Labels {
		if tenant != ctx.TenantID {
			return deny("tenant_boundary", "query returned a resource from another tenant"), nil
		}
	}
	if ctx.Metadata["audit_tenant"] != ctx.TenantID {
		return deny("tenant_boundary", "audit event tenant does not match mutation tenant"), nil
	}
	return allow("tenant boundary is preserved"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
