// tracing_handler.go
package ddmetrics

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

// type TracerHandler interface {
// 	Start(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, trace.Span)
// 	Enabled(context.Context, slog.Level) bool
// 	Handle(context.Context, slog.Record) error
// 	WithAttrs(attrs []slog.Attr) slog.Handler
// 	WithGroup(name string) slog.Handler
// }

// type TracerProvider interface {
// 	Tracer(name string, opts ...trace.TracerOption) trace.Tracer
// }

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
			// Extrai o contexto dos atributos
			var ctx context.Context = context.Background() // Fallback para contexto vazio
			var attrs []attribute.KeyValue
			r.Attrs(func(a slog.Attr) bool {
				if a.Key == "ctx" {
					if val, ok := a.Value.Any().(context.Context); ok {
						ctx = val
					}
				} else {
					attrs = append(attrs, attribute.String(a.Key, a.Value.String()))
				}
				return true
			})

			// Cria um span com base no registro
			spanName := r.Message
			_, span := h.tracer.Start(ctx, spanName)
			span.SetAttributes(attrs...)

			// Finaliza o span imediatamente (pode ser ajustado para spans mais complexos)
			span.End()
		case <-h.ctx.Done():
			return
		}
	}
}
