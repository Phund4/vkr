package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"coordinator/internal/core/domain"
)

type Store struct {
	db *sql.DB
}

func New(ctx context.Context, dsn string) (*Store, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) Sources(ctx context.Context, zoneID string) ([]domain.Source, error) {
	base := querySources
	args := []any{}
	if zoneID != "" {
		base += querySourcesZoneFilter
		args = append(args, zoneID)
	}
	rows, err := s.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Source{}
	for rows.Next() {
		var srow domain.Source
		if err := rows.Scan(&srow.SourceID, &srow.DataClass, &srow.ZoneID, &srow.SegmentID, &srow.CameraID, &srow.RTSPURL, &srow.Enabled); err != nil {
			return nil, err
		}
		out = append(out, srow)
	}
	return out, rows.Err()
}

func (s *Store) ZoneWorkers(ctx context.Context, zoneID string) (map[string][]domain.Replica, error) {
	base := queryZoneWorkers
	args := []any{}
	if zoneID != "" {
		base += queryZoneWorkersZoneFilter
		args = append(args, zoneID)
	}
	rows, err := s.db.QueryContext(ctx, base, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string][]domain.Replica{}
	for rows.Next() {
		var zid string
		var r domain.Replica
		if err := rows.Scan(&zid, &r.ClusterID, &r.InstanceID, &r.URL); err != nil {
			return nil, err
		}
		r.ClusterID = strings.TrimSpace(r.ClusterID)
		r.InstanceID = strings.TrimSpace(r.InstanceID)
		out[zid] = append(out[zid], r)
	}
	return out, rows.Err()
}

func (s *Store) UpsertHeartbeat(ctx context.Context, hb domain.WorkerHeartbeat) error {
	_, err := s.db.ExecContext(ctx, queryUpsertWorkerHeartbeat, hb.ZoneID, hb.ClusterID, hb.InstanceID, hb.Load, hb.Assignments, hb.ObservedAt)
	return err
}

func (s *Store) Heartbeats(ctx context.Context) ([]domain.WorkerHeartbeat, error) {
	rows, err := s.db.QueryContext(ctx, queryWorkerHeartbeats)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WorkerHeartbeat{}
	for rows.Next() {
		var hb domain.WorkerHeartbeat
		if err := rows.Scan(&hb.ZoneID, &hb.ClusterID, &hb.InstanceID, &hb.Load, &hb.Assignments, &hb.ObservedAt); err != nil {
			return nil, err
		}
		out = append(out, hb)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Store) String() string {
	return fmt.Sprintf("postgres-store{%p}", s.db)
}
