package postcode

import (
	"errors"
	"sync/atomic"
)

// ErrAgentBudgetExceeded is returned when an autonomous agent reaches its configured call ceiling.
var ErrAgentBudgetExceeded = errors.New("postcode: agent call budget exceeded for session")

// AgentGuardConfig configures autonomous guardrails for AI agents to prevent runaway costs
// and infinite loops during tool-calling execution.
type AgentGuardConfig struct {
	// MaxCommercialCallsPerRun limits the number of paid/commercial calls (Level 2/3) in a session.
	// 0 means unlimited.
	MaxCommercialCallsPerRun int

	// MaxTotalCallsPerRun limits the total number of all gateway calls in a session.
	// 0 means unlimited.
	MaxTotalCallsPerRun int

	// AutoDowngradeToLevel1 automatically downgrades Level 2/3 commercial requests to
	// free Level 1 validity checks if commercial limits are reached or insufficient credits occur.
	AutoDowngradeToLevel1 bool

	// AutoFallbackToOffline automatically falls back to offline reference geocoding
	// if gateway calls fail with 429 (rate limit) or 402 (insufficient credits).
	AutoFallbackToOffline bool
}

// AgentGuardMetrics reports live telemetry for calls executed under agent guardrails.
type AgentGuardMetrics struct {
	TotalCalls       int64 `json:"total_calls"`
	CommercialCalls  int64 `json:"commercial_calls"`
	DowngradedCalls  int64 `json:"downgraded_calls"`
	OfflineFallbacks int64 `json:"offline_fallbacks"`
}

type agentGuardState struct {
	cfg              AgentGuardConfig
	totalCalls       atomic.Int64
	commercialCalls  atomic.Int64
	downgradedCalls  atomic.Int64
	offlineFallbacks atomic.Int64
}

func newAgentGuardState(cfg AgentGuardConfig) *agentGuardState {
	return &agentGuardState{cfg: cfg}
}

// checkAndRecord checks limits before executing a call.
// Returns the effective LookupLevel to use, or ErrAgentBudgetExceeded if limits are hit.
func (g *agentGuardState) checkAndRecord(requestedLevel LookupLevel) (LookupLevel, error) {
	if g == nil {
		return requestedLevel, nil
	}

	// 1. Atomically check and reserve total call ceiling
	if g.cfg.MaxTotalCallsPerRun > 0 {
		for {
			currentTotal := g.totalCalls.Load()
			if currentTotal >= int64(g.cfg.MaxTotalCallsPerRun) {
				return 0, ErrAgentBudgetExceeded
			}
			if g.totalCalls.CompareAndSwap(currentTotal, currentTotal+1) {
				break
			}
		}
	} else {
		g.totalCalls.Add(1)
	}

	effectiveLevel := requestedLevel

	// 2. Atomically check and reserve commercial call ceiling for Level 2 and Level 3
	if requestedLevel > Level1 {
		if g.cfg.MaxCommercialCallsPerRun > 0 {
			for {
				currentCommercial := g.commercialCalls.Load()
				if currentCommercial >= int64(g.cfg.MaxCommercialCallsPerRun) {
					if g.cfg.AutoDowngradeToLevel1 {
						g.downgradedCalls.Add(1)
						effectiveLevel = Level1
						break
					} else {
						// Rollback total calls reservation
						g.totalCalls.Add(-1)
						return 0, ErrAgentBudgetExceeded
					}
				}
				if g.commercialCalls.CompareAndSwap(currentCommercial, currentCommercial+1) {
					break
				}
			}
		} else {
			g.commercialCalls.Add(1)
		}
	}

	return effectiveLevel, nil
}

func (g *agentGuardState) recordFallback() {
	if g != nil {
		g.offlineFallbacks.Add(1)
	}
}

func (g *agentGuardState) recordDowngrade() {
	if g != nil {
		g.downgradedCalls.Add(1)
	}
}

func (g *agentGuardState) metrics() AgentGuardMetrics {
	if g == nil {
		return AgentGuardMetrics{}
	}
	return AgentGuardMetrics{
		TotalCalls:       g.totalCalls.Load(),
		CommercialCalls:  g.commercialCalls.Load(),
		DowngradedCalls:  g.downgradedCalls.Load(),
		OfflineFallbacks: g.offlineFallbacks.Load(),
	}
}
