package postgres

const (
	querySources           = `select source_id,data_class,zone_id,segment_id,camera_id,rtsp_url,enabled from sources where enabled=true`
	querySourcesZoneFilter = ` and zone_id=$1`

	queryZoneWorkers           = `select zone_id,cluster_id,instance_id,url from ingestion_instances where enabled=true`
	queryZoneWorkersZoneFilter = ` and zone_id=$1`

	queryWorkerHeartbeats      = `select zone_id,cluster_id,instance_id,load,assignments,observed_at from worker_heartbeats`
	queryUpsertWorkerHeartbeat = `insert into worker_heartbeats(zone_id,cluster_id,instance_id,load,assignments,observed_at) values ($1,$2,$3,$4,$5,$6) on conflict (zone_id,cluster_id,instance_id) do update set load=excluded.load, assignments=excluded.assignments, observed_at=excluded.observed_at`
)
