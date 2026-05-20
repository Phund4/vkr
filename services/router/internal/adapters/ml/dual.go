package mlclient

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"

	"router/internal/core/domain"
)

// Dual два HTTP-клиента к разным путям одного ML-сервиса (инцидент и загруженность).
type Dual struct {
	Accident   *Client
	Congestion *Client
}

// NewDual создаёт пару клиентов; пустые пути заменяются на defaultAccidentPath / defaultCongestionPath.
func NewDual(baseURL, accidentPath, congestionPath string, timeout time.Duration) *Dual {
	if accidentPath == "" {
		accidentPath = defaultAccidentPath
	}
	if congestionPath == "" {
		congestionPath = defaultCongestionPath
	}
	return &Dual{
		Accident:   New(baseURL, accidentPath, timeout),
		Congestion: New(baseURL, congestionPath, timeout),
	}
}

// PostBoth параллельно вызывает оба эндпоинта; ошибка одного не отменяет второй (обычный errgroup.Group).
func (d *Dual) PostBoth(ctx context.Context, jpeg []byte, filename string, meta domain.ProcessMeta) error {
	if d == nil {
		return nil
	}
	var g errgroup.Group
	g.Go(func() error {
		return d.Accident.PostProcess(ctx, jpeg, filename, meta)
	})
	g.Go(func() error {
		return d.Congestion.PostProcess(ctx, jpeg, filename, meta)
	})
	return g.Wait()
}
