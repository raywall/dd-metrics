package ddmetrics

import (
	"context"
	"log/slog"
)

// CustomHandler base (como antes)
type CustomHandler struct {
	delegate slog.Handler       // Handler original (JSON)
	ch       chan slog.Record   // Channel para processamento assíncrono
	ctx      context.Context    // Para shutdown graceful
	cancel   context.CancelFunc // Para parar a goroutine
}

// NewBaseHandler (como antes)
func NewBaseHandler(delegate slog.Handler) *CustomHandler {
	ctx, cancel := context.WithCancel(context.Background())
	h := &CustomHandler{
		delegate: delegate,
		ch:       make(chan slog.Record, 100), // Buffer ajustável
		ctx:      ctx,
		cancel:   cancel,
	}
	return h
}

// Handle ajustado para levels customizados
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// Se for level normal, loga JSON; se for MetricLevel ou SpanLevel, pula o logging
	if r.Level != MetricLevel && r.Level != SpanLevel {
		err := h.delegate.Handle(ctx, r)
		if err != nil {
			return err
		}
	}

	// Enfileira para processamento assíncrono (não bloqueia)
	select {
	case h.ch <- r:
	default:
		// Channel cheio: descarte para evitar bloqueio/timeout
	}
	return nil
}

// Implementa slog.Handler
func (h *CustomHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.delegate.Enabled(ctx, level) || level == MetricLevel || level == SpanLevel
}

func (h *CustomHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &CustomHandler{
		delegate: h.delegate.WithAttrs(attrs),
		ch:       h.ch,
		ctx:      h.ctx,
		cancel:   h.cancel,
	}
}

func (h *CustomHandler) WithGroup(name string) slog.Handler {
	return &CustomHandler{
		delegate: h.delegate.WithGroup(name),
		ch:       h.ch,
		ctx:      h.ctx,
		cancel:   h.cancel,
	}
}

func (h *CustomHandler) Shutdown() {
	h.cancel()
}
