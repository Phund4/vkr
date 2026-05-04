package dto

import (
	"time"

	"data-service/internal/core/domain"
)

// GetProductResponse DTO ошибки для endpoint получения продукта.
type GetProductResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ProductCreation DTO модели создания продукта для HTTP-ответа.
type ProductCreation struct {
	ID                          string    `json:"id"`
	ProductID                   int64     `json:"product_id"`
	TraceID                     string    `json:"trace_id"`
	Timestamp                   time.Time `json:"timestamp"`
	EndTimestamp                time.Time `json:"end_timestamp"`
	Duration                    int64     `json:"duration"`
	GroupName                   string    `json:"group_name"`
	Presets                     []int64   `json:"presets"`
	BadPresets                  []int64   `json:"bad_presets"`
	PublishedProductPresets     int32     `json:"published_product_presets"`
	PublishedProductZeroPresets int32     `json:"published_product_zero_presets"`
	ForceTrigger                bool      `json:"force_trigger"`
	PresetIDs                   []int64   `json:"preset_ids"`
	PresetRequireds             []bool    `json:"preset_requireds"`
	PresetScores                []uint16  `json:"preset_scores"`
	PresetConversions           []uint16  `json:"preset_conversions"`
}

// NewProductCreation преобразует доменную модель в HTTP DTO.
func NewProductCreation(product *domain.ProductCreation) ProductCreation {
	return ProductCreation{
		ID:                          product.ID.String(),
		ProductID:                   product.ProductID,
		TraceID:                     product.TraceIDHex(),
		Timestamp:                   product.Timestamp,
		EndTimestamp:                product.EndTimestamp,
		Duration:                    product.Duration,
		GroupName:                   product.GroupName,
		Presets:                     product.Presets,
		BadPresets:                  product.BadPresets,
		PublishedProductPresets:     product.PublishedProductPresets,
		PublishedProductZeroPresets: product.PublishedProductZeroPresets,
		ForceTrigger:                product.ForceTrigger,
		PresetIDs:                   product.PresetIDs,
		PresetRequireds:             product.PresetRequireds,
		PresetScores:                product.PresetScores,
		PresetConversions:           product.PresetConversions,
	}
}
