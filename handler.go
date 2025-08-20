package ddmetrics

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

// Handle ajustado para level customizado
func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	// Se for level normal, loga JSON; se for MetricLevel, pula o logging
	if r.Level != MetricLevel {
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
	return h.delegate.Enabled(ctx, level)
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

// Versão para custom metrics com StatsD (ajustada)
type MetricsHandler struct {
	*CustomHandler
	client *statsd.Client
}

func NewMetricsHandler(delegate slog.Handler, statsdAddr string) *MetricsHandler {
	client, err := statsd.New(statsdAddr) // Ex: "127.0.0.1:8125"
	if err != nil {
		panic(err) // Trate melhor em produção
	}

	h := &MetricsHandler{
		CustomHandler: NewBaseHandler(delegate),
		client:        client,
	}
	go h.processMetrics()
	return h
}

func (h *MetricsHandler) processMetrics() {
	for {
		select {
		case r := <-h.ch:
			// Processa apenas se for para métrica (mas como enfileiramos tudo, filtramos aqui se necessário)
			// Extraia atributos para métricas
			var metricName, metricType string
			var metricValue float64
			r.Attrs(func(a slog.Attr) bool {
				switch a.Key {
				case "metric_name":
					metricName = a.Value.String()
				case "metric_type": // Ex: "gauge", "counter", "histogram"
					metricType = a.Value.String()
				case "metric_value":
					metricValue = a.Value.Float64()
				}
				return true
			})

			if metricName != "" && metricType != "" {
				tags := []string{"env:production"} // Extraia de attrs se necessário
				switch metricType {
				case "gauge":
					h.client.Gauge(metricName, metricValue, tags, 1.0)
				case "counter":
					h.client.Count(metricName, int64(metricValue), tags, 1.0)
				case "histogram":
					h.client.Histogram(metricName, metricValue, tags, 1.0)
				// Adicione mais
				}
				// Trate erros se necessário
			}
		case <-h.ctx.Done():
			h.client.Close()
			return
		}
	}
}