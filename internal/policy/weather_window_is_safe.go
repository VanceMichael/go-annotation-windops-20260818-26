package policy

import (
	"time"
	"windops/internal/fault"
)

func EvaluateWeatherWindowIsSafe(ctx Context) (Result, error) {
	if err := requireTenant(ctx, "policy.weather_window"); err != nil {
		return Result{}, err
	}
	if ctx.StartsAt.IsZero() || ctx.EndsAt.IsZero() || !ctx.StartsAt.Before(ctx.EndsAt) {
		return deny("invalid_window", "weather window must have an increasing time range"), nil
	}
	if ctx.Confidence < 70 {
		return deny("forecast_uncertain", "forecast confidence is below the operating threshold"), nil
	}
	if ctx.WindKPH > 70 || ctx.WaveCM > 250 {
		return deny("unsafe_weather", "wind or wave conditions exceed the marine limit"), nil
	}
	result := allow("weather window is safe for planning")
	result.Quantity = int64(ctx.Confidence)
	return result, nil
}

var _ = time.Second
var _ = fault.CodeInvalid
