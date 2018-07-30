package main

import (
	"net/http"
	"time"

	"github.com/kyeett/escape-the-world/pkg/device-gateway"
)

func main() {

	srv := &http.Server{
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8080",
	}
	gateway.Start(srv)
}
