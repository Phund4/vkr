package metrics

// Значения лейбла stage для pusher_kafka_consume_errors_total.
const (
	KafkaConsumeStageRead    = "read"
	KafkaConsumeStageProcess = "process"
)

// Значения лейбла op для pusher_clickhouse_errors_total.
const (
	ClickHouseOpPrepareIncident   = "prepare_incident"
	ClickHouseOpAppendIncident    = "append_incident"
	ClickHouseOpSendIncident      = "send_incident"
	ClickHouseOpPrepareCongestion = "prepare_congestion"
	ClickHouseOpAppendCongestion  = "append_congestion"
	ClickHouseOpSendCongestion    = "send_congestion"
)

// Значения лейбла stage для pusher_process_errors_total.
const (
	ProcessStageJSON     = "json"
	ProcessStageValidate = "validate"
	ProcessStageFileMeta = "file_meta"
	ProcessStageFileB64  = "file_b64"
)
