// Package routernotify — HTTP push reload на router (POST /v1/reload).
package routernotify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const (
	pathReload       = "/v1/reload"
	defaultNotifyTO  = 15 * time.Second
)

// Result исход уведомления одного router-инстанса.
type Result struct {
	ZoneID     string `json:"zone_id"`
	ClusterID  string `json:"cluster_id"`
	InstanceID string `json:"instance_id"`
	URL        string `json:"url"`
	OK         bool   `json:"ok"`
	Error      string `json:"error,omitempty"`
}

// ReloadURL строит POST /v1/reload из URL инстанса (часто .../metrics в ingestion_instances).
func ReloadURL(instanceURL string) string {
	u := strings.TrimRight(strings.TrimSpace(instanceURL), "/")
	if strings.HasSuffix(u, "/metrics") {
		u = strings.TrimSuffix(u, "/metrics")
	}
	return u + pathReload
}

// PostReload вызывает router POST /v1/reload.
func PostReload(ctx context.Context, instanceURL string, revision uint64, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = defaultNotifyTO
	}
	body, err := json.Marshal(map[string]uint64{"revision": revision})
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, ReloadURL(instanceURL), bytes.NewReader(body),
	)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	cli := &http.Client{Timeout: timeout}
	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("router reload %s: %s", ReloadURL(instanceURL), resp.Status)
	}
	return nil
}
