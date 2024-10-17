package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	commandExecutions *prometheus.CounterVec
	voiceExecutions   *prometheus.CounterVec
)

func Handler() http.Handler {
	return promhttp.Handler()
}

func CommandInc(command string) {
	commandExecutions.With(prometheus.Labels{"command": command}).Inc()
}

func VoiceInc(parser string) {
	voiceExecutions.With(prometheus.Labels{"parser": parser}).Inc()
}

func init() {
	commandExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_executions_total",
			Help: "Number of command executions",
		},
		[]string{"command"},
	)

	voiceExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "voice_transcriptions_total",
			Help: "Number of voice transcriptions",
		},
		[]string{"parser"},
	)
	prometheus.MustRegister(commandExecutions, voiceExecutions)
}
