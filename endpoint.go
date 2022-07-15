package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Gasoid/regular-go-bot/metrics"
)

func httpEndpoint() {
	go func() {
		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprint(w, "healthy")
		})

		http.Handle("/metrics", metrics.Handler())

		log.Fatal(http.ListenAndServe(":8080", nil))
	}()
}
