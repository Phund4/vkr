package clickhouse

import (
	"time"

	"github.com/google/uuid"

	"data-service/internal/core/domain"
)

// productCreationEventModel внутренняя модель строки из product_creation_events.
type productCreationEventModel struct {
	ID                          uuid.UUID
	ProductID                   int64
	TraceID                     uuid.UUID
	Timestamp                   time.Time
	Duration                    int64
	GroupName                   string
	Presets                     []int64
	BadPresets                  []int64
	PublishTimestamp            time.Time
	IsPublishedProduct          bool
	PublishedProductPresets     int32
	PublishedProductZeroPresets int32
	ForceTrigger                bool
	PresetIDs                   []int64
	PresetRequireds             []bool
	PresetScores                []uint16
	PresetConversions           []uint16
}

// toDomain конвертирует ClickHouse модель в доменную сущность.
func (m productCreationEventModel) toDomain() domain.ProductCreation {
	return domain.ProductCreation{
		ID:                          m.ID,
		ProductID:                   m.ProductID,
		TraceID:                     m.TraceID,
		Timestamp:                   m.Timestamp,
		EndTimestamp:                m.PublishTimestamp,
		Duration:                    m.Duration,
		GroupName:                   m.GroupName,
		Presets:                     m.Presets,
		BadPresets:                  m.BadPresets,
		PublishedProductPresets:     m.PublishedProductPresets,
		PublishedProductZeroPresets: m.PublishedProductZeroPresets,
		ForceTrigger:                m.ForceTrigger,
		PresetIDs:                   m.PresetIDs,
		PresetRequireds:             m.PresetRequireds,
		PresetScores:                m.PresetScores,
		PresetConversions:           m.PresetConversions,
	}
}
