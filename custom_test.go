package ddmetrics

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/trace"
)

// MockSlogHandler para simular o comportamento do handler slog
type MockSlogHandler struct {
	mock.Mock
}

func (m *MockSlogHandler) Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	args := m.Called(ctx, spanName, opts)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
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

func TestCustomHandler_Handle_SpanLevel(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	h := NewBaseHandler(mockHandler)
	record := slog.Record{Level: SpanLevel}
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

func TestCustomHandler_Enabled_MetricLevel(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("Enabled", mock.Anything, MetricLevel).Return(false)

	h := NewBaseHandler(mockHandler)
	enabled := h.Enabled(context.Background(), MetricLevel)

	assert.True(t, enabled) // MetricLevel é sempre habilitado
}

func TestCustomHandler_Enabled_SpanLevel(t *testing.T) {
	mockHandler := &MockSlogHandler{}
	mockHandler.On("Enabled", mock.Anything, SpanLevel).Return(false)

	h := NewBaseHandler(mockHandler)
	enabled := h.Enabled(context.Background(), SpanLevel)

	assert.True(t, enabled) // SpanLevel é sempre habilitado
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
