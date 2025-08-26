package ddmetrics

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetricLogger_Metric(t *testing.T) {
	// Configura um buffer para capturar a saída do logger
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := &MetricLogger{slog.New(handler)}

	// Testa o método Metric
	logger.Metric("test_metric", slog.String("metric_name", "test.counter"), slog.String("metric_type", "counter"), slog.Float64("metric_value", 1.0))

	// Aguarda um pouco para garantir que o log seja processado
	time.Sleep(10 * time.Millisecond)

	// Verifica se o log foi gerado com o nível correto
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)
	assert.Equal(t, float64(MetricLevel), logOutput["level"])
	assert.Equal(t, "test_metric", logOutput["msg"])
	assert.Equal(t, "test.counter", logOutput["metric_name"])
	assert.Equal(t, "counter", logOutput["metric_type"])
	assert.Equal(t, 1.0, logOutput["metric_value"])
}

func TestMetricLogger_Span(t *testing.T) {
	// Configura um buffer para capturar a saída do logger
	var buf bytes.Buffer
	handler := slog.NewJSONHandler(&buf, nil)
	logger := &MetricLogger{slog.New(handler)}

	// Cria um contexto para o teste
	ctx := context.WithValue(context.Background(), "test_key", "test_value")

	// Testa o método Span
	returnedCtx := logger.Span(ctx, "test-span", slog.String("key", "value"))

	// Aguarda um pouco para garantir que o log seja processado
	time.Sleep(10 * time.Millisecond)

	// Verifica se o log foi gerado com o nível correto
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	assert.NoError(t, err)
	assert.Equal(t, float64(SpanLevel), logOutput["level"])
	assert.Equal(t, "test-span", logOutput["msg"])
	assert.Equal(t, "value", logOutput["key"])

	// Verifica se o contexto retornado é o mesmo
	assert.Equal(t, ctx, returnedCtx)
}
