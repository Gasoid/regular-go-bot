package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	commandExecutions *prometheus.CounterVec
)

// instead of init i would like to run func explicitly
func initMetrics() {
	commandExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_executions_total",
			Help: "Number of command executions",
		},
		[]string{"command"},
	)
	prometheus.MustRegister(commandExecutions)
}
