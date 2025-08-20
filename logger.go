package ddmetrics

// MetricLevel: Level customizado para "somente métrica"
const MetricLevel = slog.Level(12)

// Wrapper para facilitar o uso do level customizado
type MetricLogger struct {
	*slog.Logger
}

func (l *MetricLogger) Metric(msg string, args ...any) {
	l.Log(context.Background(), MetricLevel, msg, args...)
}
