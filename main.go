package main

import (
	"log"
	"net/http"

	"github.com/chaewonkong/loadigo/backend"
	"github.com/chaewonkong/loadigo/lb"
)

var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

func main() {
	balancer := lb.New()
	for _, u := range backends {
		b, err := backend.New(u)
		if err != nil {
			log.Fatalf("Failed to create backend for %s: %v", u, err)
		}
		err = balancer.AddServer(b)
		if err != nil {
			log.Fatalf("Failed to add backend %s: %v", u, err)
		}
	}

	if err := http.ListenAndServe(":8080", balancer); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
