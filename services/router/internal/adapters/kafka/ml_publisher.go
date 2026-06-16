package kafkapub

import (
	"context"
	"encoding/base64"
	"encoding/json"

	"golang.org/x/sync/errgroup"

	"router/internal/core/domain"
)

// MLPublisher публикует кадры в топики its.ml.accident.in и its.ml.congestion.in.
type MLPublisher struct {
	accident   *Publisher
	congestion *Publisher
}

// NewMLPublisher создаёт пару writer'ов; пустой brokers или topic — nil.
func NewMLPublisher(brokersCSV, accidentTopic, congestionTopic string) *MLPublisher {
	acc := NewPublisher(brokersCSV, accidentTopic)
	cong := NewPublisher(brokersCSV, congestionTopic)
	if acc == nil && cong == nil {
		return nil
	}
	return &MLPublisher{accident: acc, congestion: cong}
}

// PublishBoth кодирует JPEG и пишет в оба входных топика ML параллельно.
func (p *MLPublisher) PublishBoth(ctx context.Context, jpeg []byte, meta domain.ProcessMeta) error {
	if p == nil {
		return nil
	}
	msg := domain.MLFrameMessage{
		SegmentID:         meta.SegmentID,
		CameraID:          meta.CameraID,
		ObservedAt:        meta.ObservedAt,
		PipelineStartedAt: meta.PipelineStartedAt,
		S3Key:             meta.S3Key,
		JPEGBase64:        base64.StdEncoding.EncodeToString(jpeg),
	}
	body, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	key := []byte(meta.SegmentID)
	var g errgroup.Group
	if p.accident != nil {
		g.Go(func() error {
			return p.accident.Publish(ctx, key, body)
		})
	}
	if p.congestion != nil {
		g.Go(func() error {
			return p.congestion.Publish(ctx, key, body)
		})
	}
	return g.Wait()
}

// Close закрывает оба writer.
func (p *MLPublisher) Close() error {
	if p == nil {
		return nil
	}
	var first error
	if p.accident != nil {
		if err := p.accident.Close(); err != nil && first == nil {
			first = err
		}
	}
	if p.congestion != nil {
		if err := p.congestion.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}
