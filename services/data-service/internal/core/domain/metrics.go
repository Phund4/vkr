package domain

// MetricsPort порт для создания или получения метрик разных типов
type MetricsPort interface {
	Counter(name, help string, labelNames ...string) CounterMetric
	Gauge(name, help string, labelNames ...string) GaugeMetric
	Histogram(name, help string, buckets []float64, labelNames ...string) HistogramMetric
	Summary(name, help string, objectives map[float64]float64, labelNames ...string) SummaryMetric
}

// CounterMetric представляет счетчик (counter), который может только увеличиваться
type CounterMetric interface {
	// Inc увеличивает counter на 1
	Inc(labelValues ...string)

	// Add увеличивает counter на передаваемое значение (value)
	Add(value float64, labelValues ...string)
}

// GaugeMetric представляет метрику, которая может увеличиваться и уменьшаться
type GaugeMetric interface {
	// Set устанавливает значение gauge
	Set(value float64, labelValues ...string)

	// Inc увеличивает gauge на 1
	Inc(labelValues ...string)

	// Dec уменьшает gauge на 1
	Dec(labelValues ...string)

	// Add увеличивает gauge на указанное значение
	Add(value float64, labelValues ...string)

	// Sub уменьшает gauge на указанное значение
	Sub(value float64, labelValues ...string)
}

// HistogramMetric представляет гистограмму
type HistogramMetric interface {
	// Observe добавляет значение в гистограмму
	Observe(value float64, labelValues ...string)
}

// SummaryMetric представляет summary
type SummaryMetric interface {
	// Observe добавляет значение в summary
	Observe(value float64, labelValues ...string)
}
