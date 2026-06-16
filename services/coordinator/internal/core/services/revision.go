package services

import "sync/atomic"

// assignmentsRevision увеличивается при POST /v1/assignments/reload — router перечитывает назначения.
var assignmentsRevision atomic.Uint64

func init() {
	assignmentsRevision.Store(1)
}

// CurrentAssignmentsRevision текущая версия назначений для GET /v1/assignments.
func CurrentAssignmentsRevision() uint64 {
	return assignmentsRevision.Load()
}

// BumpAssignmentsRevision сигнализирует router'ам перестроить источники.
func BumpAssignmentsRevision() uint64 {
	return assignmentsRevision.Add(1)
}
