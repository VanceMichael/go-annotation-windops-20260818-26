package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateRecoveredMetadataIsInitialized(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.recovery_metadata"); err != nil {
		return Result{}, err
	}
	metadata := cloneMap(ctx.Metadata)
	defaults := map[string]string{"source": "recovered", "schema": "2", "priority": "normal"}
	for key, value := range defaults {
		if metadata[key] == "" {
			metadata[key] = value
		}
	}
	if metadata["source"] == "" || metadata["schema"] == "" || metadata["priority"] == "" {
		return deny("recovery_metadata", "recovered record metadata is incomplete"), nil
	}
	result := allow("recovered record metadata is initialized")
	result.IDs = []string{metadata["source"], metadata["schema"], metadata["priority"]}
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
