package otelpostcode

import (
	"context"
	"net/http"
	"time"

	postcode "github.com/abcubed3/postcode"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const instrumentationName = "github.com/abcubed3/postcode/otelpostcode"

// NewTelemetry creates an adapter fulfilling postcode.Telemetry using global OTel providers.
func NewTelemetry() (postcode.Telemetry, error) {
	mp := otel.GetMeterProvider()
	meter := mp.Meter(instrumentationName)

	duration, err := meter.Float64Histogram(
		"http.client.request.duration",
		metric.WithDescription("Duration of HTTP client requests."),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10),
	)
	if err != nil {
		return postcode.Telemetry{}, err
	}

	activeReqs, err := meter.Int64UpDownCounter(
		"http.client.active_requests",
		metric.WithDescription("Number of active in-flight requests."),
		metric.WithUnit("{request}"),
	)
	if err != nil {
		return postcode.Telemetry{}, err
	}

	tracer := otel.GetTracerProvider().Tracer(instrumentationName)

	return postcode.Telemetry{
		Tracer:  &otelTracer{tracer: tracer},
		Metrics: &otelMetrics{duration: duration, activeReqs: activeReqs},
	}, nil
}

type otelTracer struct {
	tracer trace.Tracer
}

func (t *otelTracer) Start(ctx context.Context, op string) (context.Context, postcode.Span) {
	ctx, span := t.tracer.Start(ctx, op, trace.WithSpanKind(trace.SpanKindClient))
	return ctx, &otelSpan{span: span}
}

type otelSpan struct {
	span trace.Span
}

func (s *otelSpan) End()                     { s.span.End() }
func (s *otelSpan) RecordError(err error)    { s.span.RecordError(err) }
func (s *otelSpan) SetAttribute(k, v string) { s.span.SetAttributes(attribute.String(k, v)) }
func (s *otelSpan) SetStatus(code int, msg string) {
	if code >= http.StatusBadRequest {
		s.span.SetStatus(codes.Error, msg)
	} else {
		s.span.SetStatus(codes.Ok, msg)
	}
}

type otelMetrics struct {
	duration   metric.Float64Histogram
	activeReqs metric.Int64UpDownCounter
}

func (m *otelMetrics) RecordDuration(ctx context.Context, method, endpoint string, statusCode int, duration time.Duration) {
	attrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(method),
		semconv.URLPath(endpoint),
		semconv.HTTPResponseStatusCode(statusCode),
	}
	m.duration.Record(ctx, duration.Seconds(), metric.WithAttributes(attrs...))
}

func (m *otelMetrics) AddInflight(ctx context.Context, delta int64) {
	m.activeReqs.Add(ctx, delta)
}
