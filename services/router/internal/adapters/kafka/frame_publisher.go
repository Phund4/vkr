package kafkapub

import (
	"context"
	"encoding/json"
	"fmt"

	"router/internal/core/domain"
)

// FramePublisher пишет FrameIngestEvent в топик its.frames.ingest.
type FramePublisher struct {
	w *Publisher
}

// NewFramePublisher создаёт publisher кадров; nil при невалидных brokers/topic.
func NewFramePublisher(brokersCSV, topic string) *FramePublisher {
	p := NewPublisher(brokersCSV, topic)
	if p == nil {
		return nil
	}
	return &FramePublisher{w: p}
}

// PublishFrame partition key — segment_id.
func (p *FramePublisher) PublishFrame(ctx context.Context, ev domain.FrameIngestEvent) error {
	if p == nil || p.w == nil {
		return nil
	}
	b, err := json.Marshal(ev)
	if err != nil {
		return fmt.Errorf("json: %w", err)
	}
	return p.w.Publish(ctx, []byte(ev.SegmentID), b)
}

// Close освобождает writer.
func (p *FramePublisher) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
