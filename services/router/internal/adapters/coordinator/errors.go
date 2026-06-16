package coordinatorclient

import (
	"errors"
	"fmt"
)

var (
	// ErrAssignmentsStatus не-200 ответ при получении назначений.
	ErrAssignmentsStatus = errors.New("coordinator assignments unexpected status")
	// ErrHeartbeatStatus не-204 ответ при heartbeat.
	ErrHeartbeatStatus = errors.New("coordinator heartbeat unexpected status")
)

// assignmentsStatusError оборачивает ErrAssignmentsStatus с текстом статуса.
func assignmentsStatusError(status string) error {
	return fmt.Errorf("%w: %s", ErrAssignmentsStatus, status)
}

// heartbeatStatusError оборачивает ErrHeartbeatStatus с текстом статуса.
func heartbeatStatusError(status string) error {
	return fmt.Errorf("%w: %s", ErrHeartbeatStatus, status)
}
