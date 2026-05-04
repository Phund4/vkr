package domain

import (
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

// ProductCreation структура для хранения данных о создании продукта
type ProductCreation struct {
	// ID UUID записи события.
	ID uuid.UUID

	// ProductID идентификатор НМКИ.
	ProductID int64

	// TraceID UUID трейса процесса.
	TraceID uuid.UUID

	// Timestamp старт процесса создания.
	Timestamp time.Time

	// EndTimestamp конец процесса создания.
	EndTimestamp time.Time

	// Duration длительность процесса (сек).
	Duration int64

	// GroupName имя группы пресетов.
	GroupName string

	// Presets пресеты, прошедшие майнинг.
	Presets []int64

	// BadPresets пресеты, не прошедшие майнинг.
	BadPresets []int64

	// PublishedProductPresets кол-во отправленных пресетов.
	PublishedProductPresets int32

	// PublishedProductZeroPresets кол-во пресетов со score=0.
	PublishedProductZeroPresets int32

	// ForceTrigger флаг форсированного пропуша.
	ForceTrigger bool

	// PresetIDs пресеты для расчета релевантности.
	PresetIDs []int64

	// PresetRequireds признаки обязательности пресета.
	PresetRequireds []bool

	// PresetScores оценки релевантности по пресетам.
	PresetScores []uint16

	// PresetConversions конверсии по пресетам.
	PresetConversions []uint16
}

func (m ProductCreation) TraceIDHex() string {
	return hex.EncodeToString(m.TraceID[:])
}
