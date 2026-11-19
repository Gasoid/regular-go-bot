package main

import (
	"fmt"
	"net/http"
)

var (
	healthSrv = &http.Server{
		Addr: ":8080",
	}
)

//nolint:errcheck
func httpEndpoint(errChan chan<- error) {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "healthy")
	})

	err := healthSrv.ListenAndServe()
	if err != nil {
		errChan <- err
	}
}
