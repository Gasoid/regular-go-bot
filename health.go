package main

import (
	"fmt"
	"net/http"

	"github.com/Gasoid/regular-go-bot/v3/metrics"
)

//nolint:errcheck
func httpEndpoint(errChan chan<- error) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "healthy")
	})

	http.Handle("/metrics", metrics.Handler())

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		errChan <- err
	}
}
