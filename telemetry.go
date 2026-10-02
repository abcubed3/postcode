package postcode

import (
	"context"
	"time"
)

// Tracer defines the tracing contract for postcode operations.
type Tracer interface {
	Start(ctx context.Context, operation string) (context.Context, Span)
}

// Span encapsulates active span operations without coupling to any trace vendor.
type Span interface {
	End()
	RecordError(err error)
	SetAttribute(key, value string)
	SetStatus(code int, msg string)
}

// MetricsRecorder defines standard metrics contracts for client operations.
type MetricsRecorder interface {
	RecordDuration(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration)
	AddInflight(ctx context.Context, delta int64)
}

// Telemetry bundles tracing and metrics hooks.
type Telemetry struct {
	Tracer  Tracer
	Metrics MetricsRecorder
}

// NopTelemetry provides zero-overhead, no-op telemetry by default.
func NopTelemetry() Telemetry {
	return Telemetry{
		Tracer:  nopTracer{},
		Metrics: nopMetrics{},
	}
}

type nopTracer struct{}

func (nopTracer) Start(ctx context.Context, _ string) (context.Context, Span) {
	return ctx, nopSpan{}
}

type nopSpan struct{}

func (nopSpan) End()                      {}
func (nopSpan) RecordError(_ error)       {}
func (nopSpan) SetAttribute(_, _ string)  {}
func (nopSpan) SetStatus(_ int, _ string) {}

type nopMetrics struct{}

func (nopMetrics) RecordDuration(_ context.Context, _, _ string, _ int, _ time.Duration) {}
func (nopMetrics) AddInflight(_ context.Context, _ int64)                                {}
