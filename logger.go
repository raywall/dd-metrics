package ddmetrics

import (
	"context"
	"fmt"

	"github.com/DataDog/datadog-go/statsd"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type HandlerType string

const (
	JSON HandlerType = "json"
	TEXT HandlerType = "text"
)

type Handler struct {
	ctx    context.Context
	tags   []string
	span   trace.Span
	tracer *sdktrace.TracerProvider
	client *statsd.Client
}

func NewHandler(ctx context.Context, handlerType HandlerType, defaultTags []string) (*Handler, error) {
	return &Handler{
		ctx:  ctx,
		tags: defaultTags,
	}, nil
}

func (h *Handler) StartTrace(appName, domainName, serviceName, tracerHost, tracerPort string) error {
	// Create OTLP HTTP trace exporter to send data to OpenTelemetry Collector
	exporter, err := otlptracehttp.New(h.ctx,
		otlptracehttp.WithInsecure(), // Use insecure connection if testing locally
		otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%s", tracerHost, tracerPort)), // Collector OTLP HTTP endpoint
	)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %v", err)
	}

	// Create tracer provider with batch span processor and resource info
	h.tracer = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
		)),
	)
	otel.SetTracerProvider(h.tracer)

	// Obtain tracer
	tracer := otel.Tracer(appName)

	// Create a sample span
	h.ctx, h.span = tracer.Start(h.ctx, domainName)

	return nil
}

func (h *Handler) StartMetric(metricHost, metricPort, metricNamespace string) error {
	client, err := statsd.New(fmt.Sprintf("%s:%s", metricHost, metricPort),
		statsd.WithNamespace(metricNamespace),
		statsd.WithTags(h.tags),
	)
	if err != nil {
		return fmt.Errorf("failed to create statsd client: %v", err)
	}

	h.client = client
	return nil
}

func (h *Handler) SendMetric(metricName string, value float64, tags ...string) error {
	err := h.client.Distribution(metricName, value, tags, 1.0)
	if err != nil {
		return fmt.Errorf("failed to send metric: %v", err)
	}
	return nil
}

func (h *Handler) StopTrace() error {
	h.span.End()

	// Ensure all spans are flushed before exit
	if err := h.tracer.Shutdown(h.ctx); err != nil {
		return fmt.Errorf("failed to shutdown tracer provider: %v", err)
	}
	return nil
}

func (h *Handler) StopMetric() error {
	return h.client.Close()
}
