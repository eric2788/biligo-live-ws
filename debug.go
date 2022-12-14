package main

import (
	log "github.com/sirupsen/logrus"
	"net/http"
)

func debugServe() {
	if err := http.ListenAndServe("0.0.0.0:8082", http.DefaultServeMux); err != nil {
		log.Fatal(err)
	}
}
