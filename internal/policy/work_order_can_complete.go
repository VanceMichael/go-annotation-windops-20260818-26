package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateWorkOrderCanComplete(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.work_complete"); err != nil {
		return Result{}, err
	}
	if ctx.Status != "awaiting_evidence" {
		return deny("work_state", "work must await evidence before completion"), nil
	}
	required := []string{"before_photo", "after_photo", "torque_record", "technician_signoff"}
	missing := []string{}
	for _, key := range required {
		if !ctx.Flags[key] {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		result := deny("evidence_incomplete", "required completion evidence is missing")
		result.IDs = missing
		return result, nil
	}
	if ctx.Flags["open_safety_observation"] {
		return deny("safety_observation_open", "open safety observations block completion"), nil
	}
	return allow("work order can complete"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
