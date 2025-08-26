package ddmetrics

import (
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
