package coordinatorclient

const (
	// pathAssignments HTTP-путь списка назначений.
	pathAssignments = "/v1/assignments"
	// pathHeartbeat HTTP-путь heartbeat воркера.
	pathHeartbeat = "/v1/workers/heartbeat"

	// queryZoneID имя query-параметра зоны.
	queryZoneID = "zone_id"
	// queryClusterID имя query-параметра кластера.
	queryClusterID = "cluster_id"
	// queryInstanceID имя query-параметра инстанса.
	queryInstanceID = "instance_id"
	// queryDataClass имя query-параметра класса данных.
	queryDataClass = "data_class"
)
