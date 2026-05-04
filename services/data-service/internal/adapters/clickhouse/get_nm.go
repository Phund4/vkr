package clickhouse

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"data-service/internal/core/domain"
)

var ErrProductCreationNotFound = errors.New("product creation not found")

const GetProductCreationQuery = `
SELECT
	id,
	product_id,
	trace_id,
	timestamp,
	duration,
	group_name,
	presets,
	bad_presets,
	publish_timestamp,
	is_published_product,
	published_product_presets,
	published_product_zero_presets,
	force_trigger,
	preset_ids,
	preset_requireds,
	preset_scores,
	preset_conversions
FROM product_creation_events
WHERE product_id = ?
ORDER BY timestamp DESC
LIMIT 1
`

// GetProductCreation получает запись о создании продукта по product_id
func (r *Repository) GetProductCreation(ctx context.Context, productID int64) (*domain.ProductCreation, error) {
	var model productCreationEventModel

	err := r.client.QueryRowContext(ctx, GetProductCreationQuery, productID).Scan(
		&model.ID,
		&model.ProductID,
		&model.TraceID,
		&model.Timestamp,
		&model.Duration,
		&model.GroupName,
		&model.Presets,
		&model.BadPresets,
		&model.PublishTimestamp,
		&model.IsPublishedProduct,
		&model.PublishedProductPresets,
		&model.PublishedProductZeroPresets,
		&model.ForceTrigger,
		&model.PresetIDs,
		&model.PresetRequireds,
		&model.PresetScores,
		&model.PresetConversions,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return new(domain.ProductCreation), nil
		}

		return nil, fmt.Errorf("query product_creation_events by product_id=%d: %w", productID, err)
	}

	result := model.toDomain()
	return &result, nil
}
