package metrics

import "github.com/prometheus/client_golang/prometheus"

// prometheusCounter реализует интерфейс domain.CounterMetric
type prometheusCounter struct {
	counter *prometheus.CounterVec
}

func (c *prometheusCounter) Inc(labelValues ...string) {
	c.counter.WithLabelValues(labelValues...).Inc()
}

func (c *prometheusCounter) Add(value float64, labelValues ...string) {
	c.counter.WithLabelValues(labelValues...).Add(value)
}

// prometheusGauge реализует интерфейс domain.GaugeMetric
type prometheusGauge struct {
	gauge *prometheus.GaugeVec
}

func (g *prometheusGauge) Set(value float64, labelValues ...string) {
	g.gauge.WithLabelValues(labelValues...).Set(value)
}

func (g *prometheusGauge) Inc(labelValues ...string) {
	g.gauge.WithLabelValues(labelValues...).Inc()
}

func (g *prometheusGauge) Dec(labelValues ...string) {
	g.gauge.WithLabelValues(labelValues...).Dec()
}

func (g *prometheusGauge) Add(value float64, labelValues ...string) {
	g.gauge.WithLabelValues(labelValues...).Add(value)
}

func (g *prometheusGauge) Sub(value float64, labelValues ...string) {
	g.gauge.WithLabelValues(labelValues...).Sub(value)
}

// prometheusHistogram реализует интерфейс domain.HistogramMetric
type prometheusHistogram struct {
	histogram *prometheus.HistogramVec
}

func (h *prometheusHistogram) Observe(value float64, labelValues ...string) {
	h.histogram.WithLabelValues(labelValues...).Observe(value)
}

// prometheusSummary реализует интерфейс domain.SummaryMetric
type prometheusSummary struct {
	summary *prometheus.SummaryVec
}

func (s *prometheusSummary) Observe(value float64, labelValues ...string) {
	s.summary.WithLabelValues(labelValues...).Observe(value)
}
