package domain

import "time"

// TraceWithErrors структура для хранения данных о несобранных трейсах
type TraceWithErrors struct {
	// TraceID идентификатор трейса.
	TraceID string

	// SpanID идентификатор текущего спана.
	SpanID string

	// ParentSpanID идентификатор родительского спана.
	ParentSpanID *string

	// Timestamp время попадания спана в систему.
	Timestamp time.Time

	// EndTimestamp время завершения формирования спана.
	EndTimestamp time.Time

	// Duration длительность формирования (ns).
	Duration int64

	// Attributes JSON с атрибутами шага.
	Attributes string
}
