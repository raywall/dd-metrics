package main

import (
	"log/slog"
	"os"
	"time"

	ddmetrics "github.com/raywall/dd-metrics"
)

func main() {
	// Configuração
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	metricsHandler := ddmetrics.NewMetricsHandler(jsonHandler, "127.0.0.1:8125")
	logger := slog.New(metricsHandler)
	slog.SetDefault(logger)

	// Wrapper para métricas
	metricLogger := &ddmetrics.MetricLogger{Logger: logger}

	// Exemplo: Log normal (emite JSON e processa se tiver attrs de métrica)
	slog.Info("Evento normal", "key", "value")

	// Exemplo: Somente métrica (não emite JSON, só processa assincronamente)
	metricLogger.Metric("Evento métrica", "metric_name", "app.latency", "metric_type", "gauge", "metric_value", 150.5)

	// Dê tempo para processar
	time.Sleep(1 * time.Second)

	// Shutdown graceful
	metricsHandler.Shutdown()

	// // Configura o OpenTelemetry (exemplo com exportador stdout)
	// exporter, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	// if err != nil {
	// 	panic(err)
	// }
	// tp := trace.NewTracerProvider(trace.WithBatcher(exporter))
	// otel.SetTracerProvider(tp)
	// defer tp.Shutdown(context.Background())

	// // Configura o logger com TracingHandler
	// handler := ddmetrics.NewTracingHandler(slog.NewJSONHandler(os.Stdout, nil), tp)
	// logger := &ddmetrics.MetricLogger{slog.New(handler)}

	// // Usa o logger para criar um span
	// ctx := context.Background()
	// ctx = logger.Span(ctx, "my-operation", slog.String("user_id", "123"), slog.String("action", "process"))
}
