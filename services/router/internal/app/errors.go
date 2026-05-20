package app

import "errors"

var (
	// ErrMissingAWSCredentials в окружении нет ключей для S3.
	ErrMissingAWSCredentials = errors.New("AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY must be set (e.g. for MinIO)")
	// ErrCoordinatorBaseURL не задан COORDINATOR_BASE_URL.
	ErrCoordinatorBaseURL = errors.New("COORDINATOR_BASE_URL must be set")
	// ErrCoordinatorIdentity не заданы COORDINATOR_ZONE_ID, COORDINATOR_CLUSTER_ID или COORDINATOR_INSTANCE_ID.
	ErrCoordinatorIdentity = errors.New("COORDINATOR_ZONE_ID, COORDINATOR_CLUSTER_ID and COORDINATOR_INSTANCE_ID must be set")
)
