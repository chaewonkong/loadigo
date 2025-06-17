package main

import (
	"log"
	"net/http"
	"time"

	"github.com/chaewonkong/loadigo/backend/p2c"
)

var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
	"http://localhost:8084",
	"http://localhost:8085",
}

func main() {
	ticker := time.NewTicker(5 * time.Second)
	balancer := p2c.NewLoadBalancer(ticker)
	for _, u := range backends {
		b, err := p2c.NewBackend(u)
		if err != nil {
			log.Fatalf("Failed to create backend for %s: %v", u, err)
		}
		err = balancer.AddServer(u, b)
		if err != nil {
			log.Fatalf("Failed to add backend %s: %v", u, err)
		}
	}

	go balancer.CheckServerStatus()

	if err := http.ListenAndServe(":8080", balancer); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
