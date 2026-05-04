package domain

// NmHolderPush структура для хранения данных о событии пропуша в каталог
type NmHolderPush struct {
	// SpanID идентификатор шага обработки.
	SpanID string

	// TraceID идентификатор трейса процесса.
	TraceID string

	// NotProcessed признак, что спан не обработан.
	NotProcessed bool

	// Name название спана.
	Name string

	// StartTimeUnixNano время создания спана (ns).
	StartTimeUnixNano uint64

	// EndTimeUnixNano время завершения спана (ns).
	EndTimeUnixNano uint64

	// EventConversions значения конверсии по пресетам.
	EventConversions []uint16

	// EventPresets идентификаторы пресетов.
	EventPresets []uint32

	// EventScores оценки релевантности пресетов.
	EventScores []uint16
}
