package clickhouse

const (
	tableRoadIncidents  = "road_incidents"
	tableRoadCongestion = "road_congestion"

	queryRoadIncidents = `
SELECT observed_at, segment_id, camera_id, s3_key, crash_probability, incident_label, raw_ml
FROM %s
WHERE 1=1`

	queryRoadCongestion = `
SELECT observed_at, segment_id, camera_id, s3_key, congestion_score, raw_ml
FROM %s
WHERE 1=1`
)
