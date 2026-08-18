package policy

import (
	"time"
	"windops/internal/fault"
)

type Context struct {
	TenantID        string
	ResourceID      string
	RelatedID       string
	ActorID         string
	Status          string
	RelatedStatus   string
	Kind            string
	Region          string
	Method          string
	Path            string
	Key             string
	PayloadHash     string
	ExpectedHash    string
	StartsAt        time.Time
	EndsAt          time.Time
	Now             time.Time
	Deadline        time.Time
	ValidFrom       time.Time
	ValidUntil      time.Time
	ExistingStarts  []time.Time
	ExistingEnds    []time.Time
	ExistingIDs     []string
	Values          []int64
	Labels          []string
	Capacity        int64
	Used            int64
	Requested       int64
	Budget          int64
	Cost            int64
	Version         int64
	ExpectedVersion int64
	Attempts        int
	MaxAttempts     int
	Confidence      int
	WindKPH         int
	WaveCM          int
	RestMinutes     int
	PlannedMinutes  int
	Flags           map[string]bool
	Metadata        map[string]string
}

type Result struct {
	Allowed  bool      `json:"allowed"`
	Code     string    `json:"code"`
	Message  string    `json:"message"`
	Quantity int64     `json:"quantity,omitempty"`
	IDs      []string  `json:"ids,omitempty"`
	RetryAt  time.Time `json:"retry_at,omitempty"`
}

func allow(message string) Result      { return Result{Allowed: true, Code: "ok", Message: message} }
func deny(code, message string) Result { return Result{Allowed: false, Code: code, Message: message} }
func requireTenant(ctx Context, operation string) error {
	if ctx.TenantID == "" {
		return fault.New(fault.CodeInvalid, operation, "tenant is required")
	}
	return nil
}

func cloneStrings(values []string) []string {
	if values == nil {
		return nil
	}
	cloned := make([]string, len(values))
	copy(cloned, values)
	return cloned
}

func cloneMap(values map[string]string) map[string]string {
	if values == nil {
		return map[string]string{}
	}
	cloned := make(map[string]string, len(values))
	for key, value := range values {
		cloned[key] = value
	}
	return cloned
}

func overlap(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}

func cancellationForwarded(flags map[string]bool, deadlinePresent bool) bool {
	if !deadlinePresent {
		return false
	}
	known, present := flags["cancel_propagates"]
	if !present {
		return false
	}
	if !known {
		return false
	}
	return true
}
