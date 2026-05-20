package metrics

import (
	"net/http"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"data-service/internal/core/domain"
)

// PrometheusAdapter адаптер для работы с метриками Prometheus.
type PrometheusAdapter struct {
	registry   *prometheus.Registry
	namespace  string
	subsystem  string
	counters   map[string]*prometheus.CounterVec
	gauges     map[string]*prometheus.GaugeVec
	histograms map[string]*prometheus.HistogramVec
	summaries  map[string]*prometheus.SummaryVec
	mu         sync.RWMutex
}

// NewPrometheusAdapter создает новый адаптер для Prometheus
func NewPrometheusAdapter(namespace, subsystem string) *PrometheusAdapter {
	return &PrometheusAdapter{
		registry:   prometheus.NewRegistry(),
		namespace:  namespace,
		subsystem:  subsystem,
		counters:   make(map[string]*prometheus.CounterVec),
		gauges:     make(map[string]*prometheus.GaugeVec),
		histograms: make(map[string]*prometheus.HistogramVec),
		summaries:  make(map[string]*prometheus.SummaryVec),
	}
}

// Counter реализует метод Counter из интерфейса domain.MetricsPort
func (p *PrometheusAdapter) Counter(name, help string, labelNames ...string) domain.CounterMetric {
	p.mu.RLock()
	counter, ok := p.counters[name]
	p.mu.RUnlock()

	if ok {
		return &prometheusCounter{counter: counter}
	}

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
		},
		labelNames,
	)

	p.mu.Lock()
	p.counters[name] = counter
	p.registry.MustRegister(counter)
	p.mu.Unlock()

	return &prometheusCounter{counter: counter}
}

// Gauge реализует метод Gauge из интерфейса domain.MetricsPort
func (p *PrometheusAdapter) Gauge(name, help string, labelNames ...string) domain.GaugeMetric {
	p.mu.RLock()
	gauge, ok := p.gauges[name]
	p.mu.RUnlock()

	if ok {
		return &prometheusGauge{gauge: gauge}
	}

	gauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
		},
		labelNames,
	)

	p.mu.Lock()
	p.gauges[name] = gauge
	p.registry.MustRegister(gauge)
	p.mu.Unlock()

	return &prometheusGauge{gauge: gauge}
}

// Histogram реализует метод Histogram из интерфейса domain.MetricsPort
func (p *PrometheusAdapter) Histogram(name, help string, buckets []float64, labelNames ...string) domain.HistogramMetric {
	p.mu.RLock()
	histogram, ok := p.histograms[name]
	p.mu.RUnlock()

	if ok {
		return &prometheusHistogram{histogram: histogram}
	}

	histogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: p.namespace,
			Subsystem: p.subsystem,
			Name:      name,
			Help:      help,
			Buckets:   buckets,
		},
		labelNames,
	)

	p.mu.Lock()
	p.histograms[name] = histogram
	p.registry.MustRegister(histogram)
	p.mu.Unlock()

	return &prometheusHistogram{histogram: histogram}
}

// Summary реализует метод Summary из интерфейса domain.MetricsPort
func (p *PrometheusAdapter) Summary(name, help string, objectives map[float64]float64, labelNames ...string) domain.SummaryMetric {
	p.mu.RLock()
	summary, ok := p.summaries[name]
	p.mu.RUnlock()

	if ok {
		return &prometheusSummary{summary: summary}
	}

	summary = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  p.namespace,
			Subsystem:  p.subsystem,
			Name:       name,
			Help:       help,
			Objectives: objectives,
		},
		labelNames,
	)

	p.mu.Lock()
	p.summaries[name] = summary
	p.registry.MustRegister(summary)
	p.mu.Unlock()

	return &prometheusSummary{summary: summary}
}

// ServeHTTP реализует http.Handler для экспорта метрик
func (p *PrometheusAdapter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{}).ServeHTTP(w, r)
}

// PromHandler возвращает HTTP-обработчик для экспорта метрик Prometheus
func (p *PrometheusAdapter) PromHandler() http.Handler {
	return promhttp.HandlerFor(p.registry, promhttp.HandlerOpts{})
}
