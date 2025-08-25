package ddmetrics

// Versão para custom metrics com StatsD (como antes)
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
			// Processa apenas se for para métrica
			if r.Level != MetricLevel {
				continue
			}
			var metricName, metricType string
			var metricValue float64
			r.Attrs(func(a slog.Attr) bool {
				switch a.Key {
				case "metric_name":
					metricName = a.Value.String()
				case "metric_type":
					metricType = a.Value.String()
				case "metric_value":
					metricValue = a.Value.Float64()
				}
				return true
			})

			if metricName != "" && metricType != "" {
				tags := []string{"env:production"}
				switch metricType {
				case "gauge":
					h.client.Gauge(metricName, metricValue, tags, 1.0)
				case "counter":
					h.client.Count(metricName, int64(metricValue), tags, 1.0)
				case "histogram":
					h.client.Histogram(metricName, metricValue, tags, 1.0)
				}
			}
		case <-h.ctx.Done():
			h.client.Close()
			return
		}
	}
}