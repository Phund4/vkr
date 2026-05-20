package coordinatorclient

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"router/internal/config"
	"router/internal/core/domain"
)

// Client HTTP-клиент к API coordinator (назначения и heartbeat).
type Client struct {
	base string
	cli  *http.Client
}

// New создаёт клиент с базовым URL и таймаутом HTTP.
func New(baseURL string, timeout time.Duration) *Client {
	return &Client{
		base: strings.TrimRight(baseURL, "/"),
		cli:  &http.Client{Timeout: timeout},
	}
}

// fetchAssignments выполняет GET /v1/assignments с фильтрами зоны/кластера/класса данных.
func (c *Client) fetchAssignments(ctx context.Context, zoneID, clusterID, instanceID, dataClass string) ([]assignmentItemJSON, error) {
	u, err := url.Parse(c.base + pathAssignments)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set(queryZoneID, zoneID)
	q.Set(queryClusterID, clusterID)
	q.Set(queryInstanceID, instanceID)
	q.Set(queryDataClass, dataClass)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, assignmentsStatusError(resp.Status)
	}
	var ar assignmentsRespJSON
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return ar.Items, nil
}

// FetchCameraAssignments возвращает камеры с data_class=road_segment_video для данного инстанса.
func (c *Client) FetchCameraAssignments(ctx context.Context, zoneID, clusterID, instanceID string) ([]domain.Camera, error) {
	items, err := c.fetchAssignments(ctx, zoneID, clusterID, instanceID, config.DataClassRoadSegmentVideo)
	if err != nil {
		return nil, err
	}
	return assignmentItemsToDomain(items), nil
}

// SendHeartbeat отправляет POST /v1/workers/heartbeat с числом активных назначений.
func (c *Client) SendHeartbeat(ctx context.Context, zoneID, clusterID, instanceID string, assignments int) error {
	body := map[string]any{
		"zone_id":     zoneID,
		"cluster_id":  clusterID,
		"instance_id": instanceID,
		"assignments": assignments,
		"observed_at": time.Now().UTC(),
		"load":        0.0,
	}
	b, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+pathHeartbeat, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		return heartbeatStatusError(resp.Status)
	}
	return nil
}
