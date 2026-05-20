package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"

	"pusher/internal/adapters/metrics"
)

// Writer batch insert в road_incidents / road_congestion.
type Writer struct {
	conn            driver.Conn
	database        string
	incidentsTable  string
	congestionTable string
}

// New открывает native-соединение (таблицы создаются clickhouse-init).
func New(ctx context.Context, addr, database, user, password, incidentsTable, congestionTable string) (*Writer, error) {
	if incidentsTable == "" {
		incidentsTable = DefaultIncidentsTable
	}
	if congestionTable == "" {
		congestionTable = DefaultCongestionTable
	}
	addr = NormalizeAddr(addr)
	opts := &chgo.Options{
		Addr: []string{addr},
		Auth: chgo.Auth{
			Database: database,
			Username: user,
			Password: password,
		},
		DialTimeout: defaultDialTimeout,
	}
	conn, err := chgo.Open(opts)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Writer{
		conn:            conn,
		database:        database,
		incidentsTable:  incidentsTable,
		congestionTable: congestionTable,
	}, nil
}

// Close закрывает соединение.
func (w *Writer) Close() error {
	return w.conn.Close()
}

// InsertIncident одна строка инцидента.
func (w *Writer) InsertIncident(ctx context.Context, observedAt time.Time, segmentID, cameraID, s3Key string, crashProb float64, label, rawML string) error {
	q := fmt.Sprintf(insertIncidentSQLTemplate, w.database, w.incidentsTable)
	batch, err := w.conn.PrepareBatch(ctx, q)
	if err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpPrepareIncident).Inc()
		return err
	}
	if err := batch.Append(observedAt, segmentID, cameraID, s3Key, crashProb, label, rawML); err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpAppendIncident).Inc()
		return err
	}
	if err := batch.Send(); err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpSendIncident).Inc()
		return err
	}
	return nil
}

// InsertCongestion одна строка загруженности.
func (w *Writer) InsertCongestion(ctx context.Context, observedAt time.Time, segmentID, cameraID, s3Key string, congestionScore float64, rawML string) error {
	q := fmt.Sprintf(insertCongestionSQLTemplate, w.database, w.congestionTable)
	batch, err := w.conn.PrepareBatch(ctx, q)
	if err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpPrepareCongestion).Inc()
		return err
	}
	if err := batch.Append(observedAt, segmentID, cameraID, s3Key, congestionScore, rawML); err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpAppendCongestion).Inc()
		return err
	}
	if err := batch.Send(); err != nil {
		metrics.ClickHouseErrors.WithLabelValues(metrics.ClickHouseOpSendCongestion).Inc()
		return err
	}
	return nil
}

// NormalizeAddr host:port для native-протокола.
func NormalizeAddr(raw string) string {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "http://")
	raw = strings.TrimPrefix(raw, "https://")
	raw = strings.TrimPrefix(raw, "tcp://")
	return raw
}
