package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"
)

var backends = []string{
	"http://localhost:8081",
	"http://localhost:8082",
	"http://localhost:8083",
}

var current uint64 // atomic counter

func getNextBackend() string {
	idx := atomic.AddUint64(&current, 1)
	return backends[int(idx)%len(backends)]
}

func handleProxy(w http.ResponseWriter, r *http.Request) {
	targetURL := getNextBackend()
	backendURL, err := url.Parse(targetURL)
	if err != nil {
		http.Error(w, "Bad backend URL", http.StatusInternalServerError)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(backendURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, e error) {
		http.Error(w, "Backend unavailable", http.StatusBadGateway)
	}
	proxy.ServeHTTP(w, r)
}

func main() {
	log.Println("Load balancer started on :8080")
	http.HandleFunc("/", handleProxy)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
