package ddmetrics

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// TracingHandler para spans
type TracingHandler struct {
	*CustomHandler
	tracer trace.Tracer
}

func NewTracingHandler(delegate slog.Handler, tracerProvider trace.TracerProvider) *TracingHandler {
	h := &TracingHandler{
		CustomHandler: NewBaseHandler(delegate),
		tracer:        tracerProvider.Tracer("ddmetrics"),
	}
	go h.processSpans()
	return h
}

func (h *TracingHandler) processSpans() {
	for {
		select {
		case r := <-h.ch:
			// Processa apenas se for para span
			if r.Level != SpanLevel {
				continue
			}
			// Cria um span com base no registro
			ctx := r.Context
			spanName := r.Message
			_, span := h.tracer.Start(ctx, spanName)

			// Adiciona atributos do registro como atributos do span
			var attrs []attribute.KeyValue
			r.Attrs(func(a slog.Attr) bool {
				attrs = append(attrs, attribute.String(a.Key, a.Value.String()))
				return true
			})
			span.SetAttributes(attrs...)

			// Finaliza o span imediatamente (pode ser ajustado para spans mais complexos)
			span.End()
		case <-h.ctx.Done():
			return
		}
	}
}