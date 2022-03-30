package main

import (
	"fmt"
	"log"
	"net/http"
)

func runEndpoint() {

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "healthy")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
