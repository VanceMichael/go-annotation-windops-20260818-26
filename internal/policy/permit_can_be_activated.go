package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluatePermitCanBeActivated(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.permit_activate"); err != nil {
		return Result{}, err
	}
	if ctx.Status != "approved" {
		return deny("permit_state", "only approved permits can be activated"), nil
	}
	if ctx.Now.Before(ctx.StartsAt) || !ctx.Now.Before(ctx.EndsAt) {
		return deny("permit_window", "permit can only activate inside its confirmed window"), nil
	}
	if !ctx.Flags["vessel_ready"] || !ctx.Flags["crew_checked_in"] {
		return deny("departure_not_ready", "vessel and crew must be ready"), nil
	}
	if ctx.Flags["weather_suspended"] {
		return deny("weather_suspended", "marine coordinator suspended departures"), nil
	}
	return allow("permit can be activated"), nil
}

var _ = time.Second
var _ = fault.CodeInvalid
