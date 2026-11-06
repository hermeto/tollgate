// Package telemetry owns the gateway's OpenTelemetry providers and instrument names.
package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	otelmetric "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.43.0"
	oteltrace "go.opentelemetry.io/otel/trace"
)

const (
	RequestsTotal   = "tollgate_requests_total"
	RequestDuration = "tollgate_request_duration_seconds"
	GatewayOverhead = "tollgate_gateway_overhead_seconds"
	StageDuration   = "tollgate_stage_duration_seconds"
	PhaseRequest    = "request"
	PhaseResponse   = "response"
	PhaseChunk      = "chunk"
)

// Telemetry is safe to use even when the collector is down. Export happens on
// bounded background queues; HTTP handlers only record local measurements.
type Telemetry struct {
	TracerProvider *sdktrace.TracerProvider
	LoggerProvider *log.LoggerProvider
	MeterProvider  *sdkmetric.MeterProvider
	Tracer         oteltrace.Tracer
	Requests       otelmetric.Int64Counter
	Duration       otelmetric.Float64Histogram
	Overhead       otelmetric.Float64Histogram
	Stages         otelmetric.Float64Histogram
	Logger         *slog.Logger
}

type fanoutHandler struct{ handlers []slog.Handler }

func (h fanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}
func (h fanoutHandler) Handle(ctx context.Context, record slog.Record) error {
	var first error
	for _, handler := range h.handlers {
		if err := handler.Handle(ctx, record); err != nil && first == nil {
			first = err
		}
	}
	return first
}
func (h fanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	out := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		out[i] = handler.WithAttrs(attrs)
	}
	return fanoutHandler{out}
}
func (h fanoutHandler) WithGroup(name string) slog.Handler {
	out := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		out[i] = handler.WithGroup(name)
	}
	return fanoutHandler{out}
}

func New(ctx context.Context, serviceName, endpoint string) (*Telemetry, error) {
	if serviceName == "" {
		serviceName = "tollgate"
	}
	res := resource.NewSchemaless(semconv.ServiceName(serviceName))
	endpoint = strings.TrimSpace(endpoint)
	traceOpts := []otlptracehttp.Option{otlptracehttp.WithTimeout(5 * time.Second)}
	logOpts := []otlploghttp.Option{otlploghttp.WithTimeout(5 * time.Second)}
	if endpoint != "" {
		traceOpts = append(traceOpts, otlptracehttp.WithEndpointURL(endpoint+"/v1/traces"))
		logOpts = append(logOpts, otlploghttp.WithEndpointURL(endpoint+"/v1/logs"))
		if strings.HasPrefix(endpoint, "http://") {
			traceOpts = append(traceOpts, otlptracehttp.WithInsecure())
			logOpts = append(logOpts, otlploghttp.WithInsecure())
		}
	}
	traceExporter, err := otlptracehttp.New(ctx, traceOpts...)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(sdktrace.WithResource(res), sdktrace.WithSampler(sdktrace.AlwaysSample()), sdktrace.WithBatcher(traceExporter, sdktrace.WithMaxQueueSize(2048), sdktrace.WithMaxExportBatchSize(512), sdktrace.WithBatchTimeout(time.Second)))
	logExporter, err := otlploghttp.New(ctx, logOpts...)
	if err != nil {
		_ = tp.Shutdown(ctx)
		return nil, err
	}
	lp := log.NewLoggerProvider(log.WithResource(res), log.WithProcessor(log.NewBatchProcessor(logExporter, log.WithMaxQueueSize(2048), log.WithExportMaxBatchSize(512))))
	promExporter, err := prometheus.New()
	if err != nil {
		_ = tp.Shutdown(ctx)
		_ = lp.Shutdown(ctx)
		return nil, err
	}
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithResource(res), sdkmetric.WithReader(promExporter))
	meter := mp.Meter("github.com/hermeto/tollgate")
	requests, err := meter.Int64Counter(RequestsTotal)
	if err != nil {
		return nil, err
	}
	duration, err := meter.Float64Histogram(RequestDuration, otelmetric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	overhead, err := meter.Float64Histogram(GatewayOverhead, otelmetric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	stages, err := meter.Float64Histogram(StageDuration, otelmetric.WithUnit("s"))
	if err != nil {
		return nil, err
	}
	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetLoggerProvider(lp)
	otel.SetTextMapPropagator(propagation.TraceContext{})
	otelLogger := otelslog.NewHandler("github.com/hermeto/tollgate", otelslog.WithLoggerProvider(lp))
	logger := slog.New(fanoutHandler{handlers: []slog.Handler{slog.NewJSONHandler(os.Stdout, nil), otelLogger}})
	return &Telemetry{TracerProvider: tp, LoggerProvider: lp, MeterProvider: mp, Tracer: tp.Tracer("github.com/hermeto/tollgate"), Requests: requests, Duration: duration, Overhead: overhead, Stages: stages, Logger: logger}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	if t == nil {
		return nil
	}
	if err := t.TracerProvider.Shutdown(ctx); err != nil {
		return err
	}
	if err := t.LoggerProvider.Shutdown(ctx); err != nil {
		return err
	}
	return t.MeterProvider.Shutdown(ctx)
}

func (t *Telemetry) Attributes(reqID, method, path string) []attribute.KeyValue {
	return []attribute.KeyValue{attribute.String("tollgate.request_id", reqID), attribute.String("http.request.method", method), attribute.String("url.path", path)}
}

func Inject(ctx context.Context, h http.Header) {
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(h))
}

// MetricsHandler exposes the Prometheus scrape endpoint for the configured meter provider.
func MetricsHandler() http.Handler { return promhttp.Handler() }
