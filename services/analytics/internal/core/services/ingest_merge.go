package services

import (
	"context"
	"encoding/json"
	"time"

	zlog "github.com/rs/zerolog/log"

	"traffic-analytics/internal/core/domain"
)

var (
	defaultIncidentJSON   = json.RawMessage(`{"crash_probability":0,"label":"","has_incident":false}`)
	defaultCongestionJSON = json.RawMessage(`{"congestion_score":0}`)
	// combinedMLFallbackJSON запасной JSON при ошибке json.Marshal в buildCombinedML.
	combinedMLFallbackJSON = json.RawMessage(`{"incident":{"crash_probability":0,"label":"","has_incident":false},"congestion":{"congestion_score":0}}`)
)

// mergeEntry буфер для склейки частичных ответов ML (incident и congestion) по одному кадру.
type mergeEntry struct {
	segmentID         string
	cameraID          string
	observedAt        string
	s3Key             string
	pipelineStartedAt string

	incident   json.RawMessage
	congestion json.RawMessage
	hasI       bool
	hasC       bool

	timer   *time.Timer
	flushed bool
}

// mergeKey строит ключ корреляции: при наличии s3_key — префикс mergeKeyPrefixS3, иначе segment+camera+time.
func mergeKey(seg, cam, s3k, at string) string {
	if s3k != "" {
		return mergeKeyPrefixS3 + s3k
	}
	return seg + mergeKeyFieldSep + cam + mergeKeyFieldSep + at
}

// extractMLHalves определяет, какие поддеревья incident/congestion присутствуют в ev.ML.
func extractMLHalves(ev domain.RoadEvent) (hasI, hasC bool, incident, congestion json.RawMessage) {
	if len(ev.ML) == 0 || string(ev.ML) == "null" {
		return false, false, nil, nil
	}
	var probe struct {
		Incident   json.RawMessage `json:"incident"`
		Congestion json.RawMessage `json:"congestion"`
	}
	if err := json.Unmarshal(ev.ML, &probe); err != nil {
		return false, false, nil, nil
	}
	hasI = len(probe.Incident) > 0
	hasC = len(probe.Congestion) > 0
	return hasI, hasC, probe.Incident, probe.Congestion
}

// applyBase дополняет mergeEntry идентификаторами из последнего входящего события (не затирает пустыми строками).
func (e *mergeEntry) applyBase(seg, cam, atStr, s3k, pipeAt string) {
	if seg != "" {
		e.segmentID = seg
	}
	if cam != "" {
		e.cameraID = cam
	}
	if atStr != "" {
		e.observedAt = atStr
	}
	if s3k != "" {
		e.s3Key = s3k
	}
	if pipeAt != "" {
		e.pipelineStartedAt = pipeAt
	}
}

// toRoadEvent собирает domain.RoadEvent для processIngestDirect с уже склеенным ml.
func (e *mergeEntry) toRoadEvent(ml json.RawMessage) domain.RoadEvent {
	return domain.RoadEvent{
		SegmentID:         e.segmentID,
		CameraID:          e.cameraID,
		ObservedAt:        e.observedAt,
		S3Key:             e.s3Key,
		PipelineStartedAt: e.pipelineStartedAt,
		ML:                ml,
	}
}

// buildCombinedML мержит fragmentы incident/congestion; отсутствующие части заменяются дефолтным JSON.
func (e *mergeEntry) buildCombinedML() json.RawMessage {
	inc := e.incident
	cong := e.congestion
	if !e.hasI {
		inc = defaultIncidentJSON
	}
	if !e.hasC {
		cong = defaultCongestionJSON
	}
	wrap := struct {
		Incident   json.RawMessage `json:"incident"`
		Congestion json.RawMessage `json:"congestion"`
	}{Incident: inc, Congestion: cong}
	b, err := json.Marshal(wrap)
	if err != nil {
		return combinedMLFallbackJSON
	}
	return b
}

// mergeIngest обновляет буфер склейки; при полной паре incident+congestion сразу вызывает processIngestDirect.
func (s *IngestService) mergeIngest(ctx context.Context, seg, cam, atStr, s3k, pipeAt string, hasI, hasC bool, incident, congestion json.RawMessage) error {
	key := mergeKey(seg, cam, s3k, atStr)

	s.mergeMu.Lock()
	e := s.mergeBuf[key]
	if e == nil {
		e = &mergeEntry{}
		s.mergeBuf[key] = e
	}
	e.applyBase(seg, cam, atStr, s3k, pipeAt)
	if hasI {
		e.incident = incident
		e.hasI = true
	}
	if hasC {
		e.congestion = congestion
		e.hasC = true
	}

	if e.hasI && e.hasC {
		ml := e.buildCombinedML()
		evOut := e.toRoadEvent(ml)
		if e.timer != nil {
			e.timer.Stop()
		}
		delete(s.mergeBuf, key)
		s.mergeMu.Unlock()
		return s.processIngestDirect(ctx, evOut)
	}

	s.scheduleMergeFlushLocked(key, e)
	s.mergeMu.Unlock()
	return nil
}

// scheduleMergeFlushLocked перезапускает таймер отложенной публикации при неполной паре ML.
func (s *IngestService) scheduleMergeFlushLocked(key string, e *mergeEntry) {
	if e.flushed {
		return
	}
	if e.timer != nil {
		e.timer.Stop()
	}
	to := s.cfg.IngestMergeTimeout
	if to <= 0 {
		to = mergeScheduleDefaultSeconds * time.Second
	}
	e.timer = time.AfterFunc(to, func() { s.mergeTimeoutFire(key) })
}

// mergeTimeoutFire публикует склеенное событие по таймауту или удаляет пустую запись без ML.
func (s *IngestService) mergeTimeoutFire(key string) {
	s.mergeMu.Lock()
	e, ok := s.mergeBuf[key]
	if !ok || e == nil || e.flushed {
		s.mergeMu.Unlock()
		return
	}
	if !e.hasI && !e.hasC {
		delete(s.mergeBuf, key)
		s.mergeMu.Unlock()
		return
	}
	e.flushed = true
	if e.timer != nil {
		e.timer.Stop()
	}
	delete(s.mergeBuf, key)
	ml := e.buildCombinedML()
	evOut := e.toRoadEvent(ml)
	s.mergeMu.Unlock()

	ctx, cancel := context.WithTimeout(s.appCtx, s.publishTO)
	defer cancel()
	if err := s.processIngestDirect(ctx, evOut); err != nil {
		zlog.Warn().Err(err).Msg("ingest merge timeout flush")
	}
}
