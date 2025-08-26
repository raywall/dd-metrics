// logger.go
package ddmetrics

import (
	"context"
	"log/slog"
)

// MetricLevel: Level customizado para "somente métrica"
const MetricLevel = slog.Level(12)

// SpanLevel: Level customizado para spans
const SpanLevel = slog.Level(16)

// Wrapper para facilitar o uso do level customizado
type MetricLogger struct {
	*slog.Logger
}

// Metric: Registra uma métrica com o nível MetricLevel
func (l *MetricLogger) Metric(msg string, args ...any) {
	l.Log(context.Background(), MetricLevel, msg, args...)
}

// Span: Cria um span com o nível SpanLevel
func (l *MetricLogger) Span(ctx context.Context, name string, args ...any) context.Context {
	// Cria um contexto com o span, se o tracing estiver habilitado
	l.Log(ctx, SpanLevel, name, args...)
	return ctx // Retorna o contexto, que pode conter o span
}
