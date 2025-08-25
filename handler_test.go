package ddmetrics

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockStatsDClient para simular o comportamento do cliente StatsD
type MockStatsDClient struct {
	mock.Mock
}

func (m *MockStatsDClient) Gauge(name string, value float64, tags []string, rate float64) error {
	args := m.Called(name, value, tags, rate)
	return args.Error(0)
}

func (m *MockStatsDClient) Count(name string, value int64, tags []string, rate float64) error {
	args := m.Called(name, value, tags, rate)
	return args.Error(0)
}

func (m *MockStatsDClient) Histogram(name string, value float64, tags []string, rate float64) error {
	args := m.Called(name, value, tags, rate)
	return args.Error(0)
}

func (m *MockStatsDClient) Close() error {
	args := m.Called()
	return args.Error(0)
}

// MockSlogHandler para simular o comportamento do handler slog
type MockSlogHandler struct {
	mock.Mock
}

func (m *MockSlogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	args := m.Called(ctx, level)
	return args.Bool(0)
}

func (m *MockSlogHandler) Handle(ctx context.Context, r slog.Record) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockSlogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	args := m.Called(attrs)
	return args.Get(0).(slog.Handler)
}

func (m *MockSlogHandler) WithGroup(name string) slog.Handler {
	args := m.Called(name)
	return args.Get(0).(slog.Handler)
}

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

func TestCustomHandler_NewBaseHandler(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	h := NewBaseHandler(mockHandler)

	assert.NotNil(t, h)
	assert.Equal(t, mockHandler, h.delegate)
	assert.NotNil(t, h.ch)
	assert.NotNil(t, h.ctx)
	assert.NotNil(t, h.cancel)
}

func TestCustomHandler_Handle_NonMetricLevel(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("Handle", mock.Anything, mock.Anything).Return(nil)

	h := NewBaseHandler(mockHandler)
	record := slog.Record{Level: slog.LevelInfo}
	err := h.Handle(context.Background(), record)

	assert.NoError(t, err)
	mockHandler.AssertCalled(t, "Handle", mock.Anything, record)
	assert.Len(t, h.ch, 1) // Verifica se o record foi enfileirado
}

func TestCustomHandler_Handle_MetricLevel(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	h := NewBaseHandler(mockHandler)
	record := slog.Record{Level: MetricLevel}
	err := h.Handle(context.Background(), record)

	assert.NoError(t, err)
	mockHandler.AssertNotCalled(t, "Handle", mock.Anything, mock.Anything)
	assert.Len(t, h.ch, 1) // Verifica se o record foi enfileirado
}

func TestCustomHandler_Handle_ChannelFull(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	h := NewBaseHandler(mockHandler)

	// Preenche o canal
	for i := 0; i < 100; i++ {
		h.ch <- slog.Record{}
	}

	// Tenta enfileirar mais um (canal cheio)
	record := slog.Record{Level: slog.LevelInfo}
	err := h.Handle(context.Background(), record)

	assert.NoError(t, err)
	assert.Len(t, h.ch, 100) // Canal ainda tem 100 registros
}

func TestCustomHandler_Enabled(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("Enabled", mock.Anything, slog.LevelInfo).Return(true)

	h := NewBaseHandler(mockHandler)
	enabled := h.Enabled(context.Background(), slog.LevelInfo)

	assert.True(t, enabled)
	mockHandler.AssertCalled(t, "Enabled", mock.Anything, slog.LevelInfo)
}

func TestCustomHandler_WithAttrs(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("WithAttrs", mock.Anything).Return(mockHandler)

	h := NewBaseHandler(mockHandler)
	attrs := []slog.Attr{slog.String("key", "value")}
	newHandler := h.WithAttrs(attrs)

	assert.IsType(t, &CustomHandler{}, newHandler)
	assert.Equal(t, h.ch, newHandler.(*CustomHandler).ch)
	assert.Equal(t, h.ctx, newHandler.(*CustomHandler).ctx)
	assert.Equal(t, h.cancel, newHandler.(*CustomHandler).cancel)
	mockHandler.AssertCalled(t, "WithAttrs", attrs)
}

func TestCustomHandler_WithGroup(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("WithGroup", "group").Return(mockHandler)

	h := NewBaseHandler(mockHandler)
	newHandler := h.WithGroup("group")

	assert.IsType(t, &CustomHandler{}, newHandler)
	assert.Equal(t, h.ch, newHandler.(*CustomHandler).ch)
	assert.Equal(t, h.ctx, newHandler.(*CustomHandler).ctx)
	assert.Equal(t, h.cancel, newHandler.(*CustomHandler).cancel)
	mockHandler.AssertCalled(t, "WithGroup", "group")
}

func TestCustomHandler_Shutdown(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	h := NewBaseHandler(mockHandler)

	h.Shutdown()
	select {
	case <-h.ctx.Done():
		// Sucesso: contexto foi cancelado
	default:
		t.Error("Shutdown não cancelou o contexto")
	}
}

func TestMetricsHandler_NewMetricsHandler(t *testing.T) {
	// Simula um endereço StatsD válido para evitar panic
	t.Setenv("STATSD_ADDR", "127.0.0.1:8125")
	mockHandler := &MockSlogHandler{}
	h := NewMetricsHandler(mockHandler, "127.0.0.1:8125")

	assert.NotNil(t, h)
	assert.NotNil(t, h.CustomHandler)
	assert.NotNil(t, h.client)
}

func TestMetricsHandler_ProcessMetrics_Gauge(t *testing.T) {
	mockStatsD := &MockStatsDClient{}
	mockStatsD.On("Gauge", "test.gauge", 1.0, []string{"env:production"}, 1.0).Return(nil)

	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
		client:        mockStatsD,
	}
	go h.processMetrics()

	// Envia um registro com métrica
	record := slog.Record{Level: MetricLevel}
	record.AddAttrs(
		slog.String("metric_name", "test.gauge"),
		slog.String("metric_type", "gauge"),
		slog.Float64("metric_value", 1.0),
	)
	h.ch <- record

	// Aguarda processamento
	time.Sleep(10 * time.Millisecond)

	mockStatsD.AssertCalled(t, "Gauge", "test.gauge", 1.0, []string{"env:production"}, 1.0)
}

func TestMetricsHandler_ProcessMetrics_Counter(t *testing.T) {
	mockStatsD := &MockStatsDClient{}
	mockStatsD.On("Count", "test.counter", int64(1), []string{"env:production"}, 1.0).Return(nil)

	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
		client:        mockStatsD,
	}
	go h.processMetrics()

	// Envia um registro com métrica
	record := slog.Record{Level: MetricLevel}
	record.AddAttrs(
		slog.String("metric_name", "test.counter"),
		slog.String("metric_type", "counter"),
		slog.Float64("metric_value", 1.0),
	)
	h.ch <- record

	// Aguarda processamento
	time.Sleep(10 * time.Millisecond)

	mockStatsD.AssertCalled(t, "Count", "test.counter", int64(1), []string{"env:production"}, 1.0)
}

func TestMetricsHandler_ProcessMetrics_Histogram(t *testing.T) {
	mockStatsD := &MockStatsDClient{}
	mockStatsD.On("Histogram", "test.histogram", 1.0, []string{"env:production"}, 1.0).Return(nil)

	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
		client:        mockStatsD,
	}
	go h.processMetrics()

	// Envia um registro com métrica
	record := slog.Record{Level: MetricLevel}
	record.AddAttrs(
		slog.String("metric_name", "test.histogram"),
		slog.String("metric_type", "histogram"),
		slog.Float64("metric_value", 1.0),
	)
	h.ch <- record

	// Aguarda processamento
	time.Sleep(10 * time.Millisecond)

	mockStatsD.AssertCalled(t, "Histogram", "test.histogram", 1.0, []string{"env:production"}, 1.0)
}

func TestMetricsHandler_ProcessMetrics_NoMetric(t *testing.T) {
	mockStatsD := &MockStatsDClient{}
	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
		client:        mockStatsD,
	}
	go h.processMetrics()

	// Envia um registro sem métrica
	record := slog.Record{Level: MetricLevel}
	record.AddAttrs(slog.String("key", "value"))
	h.ch <- record

	// Aguarda processamento
	time.Sleep(10 * time.Millisecond)

	mockStatsD.AssertNotCalled(t, "Gauge", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStatsD.AssertNotCalled(t, "Count", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockStatsD.AssertNotCalled(t, "Histogram", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestMetricsHandler_Shutdown(t *testing.T) {
	mockStatsD := &MockStatsDClient{}
	mockStatsD.On("Close").Return(nil)

	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
		client:        mockStatsD,
	}
	go h.processMetrics()

	h.Shutdown()
	time.Sleep(10 * time.Millisecond)

	select {
	case <-h.ctx.Done():
		// Sucesso: contexto foi cancelado
	default:
		t.Error("Shutdown não cancelou o contexto")
	}
	mockStatsD.AssertCalled(t, "Close")
}