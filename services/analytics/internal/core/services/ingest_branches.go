package services

import "context"

// ProcessAccidentResult обрабатывает its.ml.accident.out без ожидания congestion.
func (s *IngestService) ProcessAccidentResult(ctx context.Context, body []byte) error {
	ev, seg, cam, atStr, s3k, pipeAt, err := parseRoadEvent(body)
	if err != nil {
		return err
	}
	ml, raw, hasML := parseMLParsed(ev)
	if !hasML {
		return ErrIngestValidation
	}
	alert := s.incidentAlert(ml, hasML)
	return s.publishPersist(ctx, seg, cam, atStr, s3k, pipeAt, ml, raw, hasML,
		alert, false, true, false)
}

// ProcessCongestionResult обрабатывает its.ml.congestion.out отдельным событием.
func (s *IngestService) ProcessCongestionResult(ctx context.Context, body []byte) error {
	ev, seg, cam, atStr, s3k, pipeAt, err := parseRoadEvent(body)
	if err != nil {
		return err
	}
	ml, raw, hasML := parseMLParsed(ev)
	if !hasML {
		return ErrIngestValidation
	}
	return s.publishPersist(ctx, seg, cam, atStr, s3k, pipeAt, ml, raw, hasML,
		false, true, false, true)
}

// ProcessVideoMeta принимает its.video.ingest (метаданные кадра); persist не выполняет.
func (s *IngestService) ProcessVideoMeta(ctx context.Context, body []byte) error {
	_, _, _, _, _, _, err := parseRoadEvent(body)
	return err
}

// ProcessIngest совместимость HTTP: маршрутизация по содержимому ml без склейки веток.
func (s *IngestService) ProcessIngest(ctx context.Context, body []byte) error {
	ev, seg, cam, atStr, s3k, pipeAt, err := parseRoadEvent(body)
	if err != nil {
		return err
	}
	hasI, hasC, _, _ := extractMLHalves(ev)
	if hasI && !hasC {
		return s.ProcessAccidentResult(ctx, body)
	}
	if hasC && !hasI {
		return s.ProcessCongestionResult(ctx, body)
	}
	ml, raw, hasML := parseMLParsed(ev)
	if !hasML {
		return ErrIngestValidation
	}
	alert := s.incidentAlert(ml, hasML)
	persistCong := s.shouldPersistCongestion(seg, cam)
	return s.publishPersist(ctx, seg, cam, atStr, s3k, pipeAt, ml, raw, hasML,
		alert, persistCong, true, true)
}
