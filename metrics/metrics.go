package metrics

import (
	"net/http"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	commandExecutions *prometheus.CounterVec
	parserExecutions  *prometheus.CounterVec
	receivedMessages  *prometheus.CounterVec
)

func Handler() http.Handler {
	return promhttp.Handler()
}

func CommandInc(command string, err error) {
	commandExecutions.With(
		prometheus.Labels{
			"command":   command,
			"is_failed": strconv.FormatBool(err != nil),
		}).Inc()
}

func ParserInc(parser string, err error) {
	parserExecutions.With(
		prometheus.Labels{
			"parser":    parser,
			"is_failed": strconv.FormatBool(err != nil),
		}).Inc()
}

func MessagesInc(isPrivate bool) {
	receivedMessages.With(prometheus.Labels{"is_private": strconv.FormatBool(isPrivate)}).Inc()
}

func init() {
	commandExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "command_executions_total",
			Help: "Number of command executions",
		},
		[]string{"command", "is_failed"},
	)

	parserExecutions = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "parser_executions_total",
			Help: "Number of parser executions",
		},
		[]string{"parser", "is_failed"},
	)

	receivedMessages = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "received_messages_total",
			Help: "Number of received messages",
		},
		[]string{"is_private"},
	)

	prometheus.MustRegister(
		commandExecutions,
		parserExecutions,
		receivedMessages,
	)
}
