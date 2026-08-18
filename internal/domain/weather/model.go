package weather

import (
	"strings"
	"time"
	"windops/internal/fault"
)

type Status string

const (
	StatusForecast  Status = "forecast"
	StatusConfirmed Status = "confirmed"
	StatusExpired   Status = "expired"
)

type Window struct {
	ID         string    `json:"id"`
	TenantID   string    `json:"tenant_id"`
	FarmID     string    `json:"farm_id"`
	StartsAt   time.Time `json:"starts_at"`
	EndsAt     time.Time `json:"ends_at"`
	MaxWaveCM  int       `json:"max_wave_cm"`
	MaxWindKPH int       `json:"max_wind_kph"`
	Confidence int       `json:"confidence"`
	Status     Status    `json:"status"`
	Version    int64     `json:"version"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func (v Window) Validate() error {
	if strings.TrimSpace(v.ID) == "" {
		return fault.New(fault.CodeInvalid, "weather.validate", "ID is required")
	}
	if strings.TrimSpace(v.TenantID) == "" {
		return fault.New(fault.CodeInvalid, "weather.validate", "tenant is required")
	}
	if strings.TrimSpace(v.FarmID) == "" {
		return fault.New(fault.CodeInvalid, "weather.validate", "FarmID is required")
	}
	if !v.Status.Valid() {
		return fault.New(fault.CodeInvalid, "weather.validate", "status is invalid")
	}
	if v.Version < 1 {
		return fault.New(fault.CodeInvalid, "weather.validate", "version must be positive")
	}
	if v.CreatedAt.IsZero() || v.UpdatedAt.IsZero() {
		return fault.New(fault.CodeInvalid, "weather.validate", "timestamps are required")
	}
	return nil
}

func (s Status) Valid() bool {
	switch s {
	case StatusForecast, StatusConfirmed, StatusExpired:
		return true
	default:
		return false
	}
}

func (v Window) WithStatus(next Status, now time.Time) (Window, error) {
	if !next.Valid() {
		return Window{}, fault.New(fault.CodeInvalid, "weather.transition", "target status is invalid")
	}
	if v.Status == next {
		return v, nil
	}
	v.Status = next
	v.Version++
	v.UpdatedAt = now.UTC()
	return v, v.Validate()
}

func (v Window) Clone() Window { return v }
