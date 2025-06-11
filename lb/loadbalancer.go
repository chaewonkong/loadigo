package lb

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	AddServer(svr http.Handler) error
}

type loadBalancer struct {
	servers []http.Handler
	current uint64
}

// New creates a new LoadBalancer instance.
func New() LoadBalancer {
	return &loadBalancer{
		servers: make([]http.Handler, 0),
		current: 0,
	}
}

func (lb *loadBalancer) nextServer() http.Handler {
	if len(lb.servers) == 0 {
		return nil
	}
	idx := atomic.AddUint64(&lb.current, 1)
	return lb.servers[int(idx)%len(lb.servers)]
}

func (lb *loadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svr := lb.nextServer()
	if svr == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	svr.ServeHTTP(w, r)
}

func (lb *loadBalancer) AddServer(svr http.Handler) error {
	if svr == nil {
		return fmt.Errorf("server cannot be nil")
	}
	lb.servers = append(lb.servers, svr)
	return nil
}
