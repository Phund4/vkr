// Package metrics — Prometheus для router.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var OperationErrors = promauto.NewCounterVec(
	prometheus.CounterOpts{
		Name: "router_operation_errors_total",
		Help: "Errors during video ingest pipeline by stage.",
	},
	[]string{"stage"},
)
