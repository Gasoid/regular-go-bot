package main

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/Gasoid/regular-go-bot/v3/metrics"
)

//nolint:errcheck
func httpEndpoint() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "healthy")
	})

	http.Handle("/metrics", metrics.Handler())

	err := http.ListenAndServe(":8080", nil)
	slog.Error("health endpoint exit with err", "err", err)
}
