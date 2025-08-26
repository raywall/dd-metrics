package ddmetrics

import (
	"context"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// MockTracer para simular o comportamento do tracer
type MockTracer struct {
	mock.Mock
}

func (m *MockTracer) Start(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	args := m.Called(ctx, name, opts)
	return args.Get(0).(context.Context), args.Get(1).(trace.Span)
}

func (m *MockTracer) tracer() {
	m.Called()
}

func (m *MockTracer) Enabled(ctx context.Context, level slog.Level) bool {
	args := m.Called(ctx, level)
	return args.Bool(0)
}

func (m *MockTracer) Handle(ctx context.Context, r slog.Record) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockTracer) WithAttrs(attrs []slog.Attr) slog.Handler {
	args := m.Called(attrs)
	return args.Get(0).(slog.Handler)
}

func (m *MockTracer) WithGroup(name string) slog.Handler {
	args := m.Called(name)
	return args.Get(0).(slog.Handler)
}

// MockSpan para simular o comportamento do span
type MockSpan struct {
	mock.Mock
}

func (m *MockSpan) SetAttributes(attrs ...attribute.KeyValue) {
	m.Called(attrs)
}

func (m *MockSpan) End(options ...trace.SpanEndOption) {
	m.Called(options)
}

func (m *MockSpan) span() {
	m.Called()
}

// MockTracerProvider para simular o TracerProvider
type MockTracerProvider struct {
	mock.Mock
}

func (m *MockTracerProvider) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	args := m.Called(name, opts)
	return args.Get(0).(trace.Tracer)
}

// func TestTracingHandler_NewTracingHandler(t *testing.T) {
// 	mockHandler := &MockSlogHandler{}
// 	mockTracerProvider := &MockTracerProvider{}
// 	mockTracer := &MockTracer{}
// 	mockTracerProvider.On("Tracer", "ddmetrics", mock.Anything).Return(mockTracer)

// 	h := NewTracingHandler(mockHandler, mockTracerProvider)

// 	assert.NotNil(t, h)
// 	assert.NotNil(t, h.CustomHandler)
// 	assert.NotNil(t, h.tracer)
// 	mockTracerProvider.AssertCalled(t, "Tracer", "ddmetrics", mock.Anything)
// }

// func TestTracingHandler_ProcessSpans(t *testing.T) {
// 	mockTracer := &MockTracer{}
// 	mockSpan := &MockSpan{}
// 	mockSpan.On("SetAttributes", mock.Anything).Return()
// 	mockSpan.On("End", mock.Anything).Return()

// 	// Cria um contexto para o teste
// 	ctx := context.WithValue(context.Background(), "test_key", "test_value")
// 	mockTracer.On("Start", ctx, "test-span", mock.Anything).Return(ctx, mockSpan)

// 	h := &TracingHandler{
// 		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
// 		tracer:        mockTracer,
// 	}
// 	go h.processSpans()

// 	// Envia um registro com span
// 	record := slog.Record{Level: SpanLevel, Message: "test-span"}
// 	record.AddAttrs(
// 		slog.String("key", "value"),
// 		slog.Any("ctx", ctx),
// 	)
// 	h.ch <- record

// 	// Aguarda processamento
// 	time.Sleep(10 * time.Millisecond)

// 	mockTracer.AssertCalled(t, "Start", ctx, "test-span", mock.Anything)
// 	mockSpan.AssertCalled(t, "SetAttributes", []attribute.KeyValue{attribute.String("key", "value")})
// 	mockSpan.AssertCalled(t, "End", mock.Anything)
// }

// func TestTracingHandler_ProcessSpans_NoContext(t *testing.T) {
// 	mockTracer := &MockTracer{}
// 	mockSpan := &MockSpan{}
// 	mockSpan.On("SetAttributes", mock.Anything).Return()
// 	mockSpan.On("End", mock.Anything).Return()

// 	// Configura o tracer para aceitar context.Background()
// 	mockTracer.On("Start", context.Background(), "test-span", mock.Anything).Return(context.Background(), mockSpan)

// 	h := &TracingHandler{
// 		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
// 		tracer:        mockTracer,
// 	}
// 	go h.processSpans()

// 	// Envia um registro sem contexto
// 	record := slog.Record{Level: SpanLevel, Message: "test-span"}
// 	record.AddAttrs(slog.String("key", "value"))
// 	h.ch <- record

// 	// Aguarda processamento
// 	time.Sleep(10 * time.Millisecond)

// 	mockTracer.AssertCalled(t, "Start", context.Background(), "test-span", mock.Anything)
// 	mockSpan.AssertCalled(t, "SetAttributes", []attribute.KeyValue{attribute.String("key", "value")})
// 	mockSpan.AssertCalled(t, "End", mock.Anything)
// }

// func TestTracingHandler_IgnoreNonSpanLevel(t *testing.T) {
// 	mockTracer := &MockTracer{}
// 	h := &TracingHandler{
// 		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
// 		tracer:        mockTracer,
// 	}
// 	go h.processSpans()

// 	// Envia um registro com nível diferente
// 	record := slog.Record{Level: MetricLevel}
// 	h.ch <- record

// 	// Aguarda processamento
// 	time.Sleep(10 * time.Millisecond)

// 	mockTracer.AssertNotCalled(t, "Start", mock.Anything, mock.Anything, mock.Anything)
// }

// func TestTracingHandler_Shutdown(t *testing.T) {
// 	mockTracer := &MockTracer{}
// 	h := &TracingHandler{
// 		CustomHandler: NewBaseHandler(&MockSlogHandler{}),
// 		tracer:        mockTracer,
// 	}
// 	go h.processSpans()

// 	h.Shutdown()
// 	time.Sleep(10 * time.Millisecond)

// 	select {
// 	case <-h.ctx.Done():
// 		// Sucesso: contexto foi cancelado
// 	default:
// 		t.Error("Shutdown não cancelou o contexto")
// 	}
// }
