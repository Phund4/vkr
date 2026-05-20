package services

import (
	"context"

	"data-service/internal/core/domain"
)

// RoadDataService единый сервис чтения дорожных событий и кадров.
type RoadDataService struct {
	repo   RoadDataRepository
	frames FrameURLSigner
}

// NewRoadDataService создаёт сервис; frames может быть nil — без presigned URL.
func NewRoadDataService(repo RoadDataRepository, frames FrameURLSigner) *RoadDataService {
	return &RoadDataService{repo: repo, frames: frames}
}

// Close закрывает внешние ресурсы.
func (s *RoadDataService) Close(ctx context.Context) error {
	return s.repo.Close(ctx)
}

// ListRoadIncidents список инцидентов.
func (s *RoadDataService) ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error) {
	return s.repo.ListRoadIncidents(ctx, p)
}

// ListRoadCongestion список загруженности.
func (s *RoadDataService) ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error) {
	return s.repo.ListRoadCongestion(ctx, p)
}

// ListIncidentsByCamera инциденты одной камеры за интервал.
func (s *RoadDataService) ListIncidentsByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.RoadIncident, error) {
	return s.repo.ListIncidentsByCamera(ctx, p)
}

// AvgCongestionBySegment средняя загруженность контура (segment) за интервал.
func (s *RoadDataService) AvgCongestionBySegment(ctx context.Context, p domain.RoadTimeRangeParams) (domain.CongestionAverageResult, error) {
	return s.repo.AvgCongestionBySegment(ctx, p)
}

// ListFramesByCamera кадры камеры за интервал; при наличии FrameURLSigner заполняет DownloadURL.
func (s *RoadDataService) ListFramesByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.FrameWithURL, error) {
	refs, err := s.repo.ListFramesByCamera(ctx, p)
	if err != nil {
		return nil, err
	}
	out := make([]domain.FrameWithURL, 0, len(refs))
	for _, fr := range refs {
		item := domain.FrameWithURL{FrameRef: fr}
		if s.frames != nil && fr.S3Key != "" {
			u, uerr := s.frames.PresignGetURL(ctx, fr.S3Key)
			if uerr == nil {
				item.DownloadURL = u
			}
		}
		out = append(out, item)
	}
	return out, nil
}
