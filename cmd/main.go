package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
)

func main() {
	// Configuração
	jsonHandler := slog.NewJSONHandler(os.Stdout, nil)
	metricsHandler := NewMetricsHandler(jsonHandler, "127.0.0.1:8125")
	logger := slog.New(metricsHandler)
	slog.SetDefault(logger)

	// Wrapper para métricas
	metricLogger := &MetricLogger{Logger: logger}

	// Exemplo: Log normal (emite JSON e processa se tiver attrs de métrica)
	slog.Info("Evento normal", "key", "value")

	// Exemplo: Somente métrica (não emite JSON, só processa assincronamente)
	metricLogger.Metric("Evento métrica", "metric_name", "app.latency", "metric_type", "gauge", "metric_value", 150.5)

	// Dê tempo para processar
	time.Sleep(1 * time.Second)

	// Shutdown graceful
	metricsHandler.Shutdown()
}