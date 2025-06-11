package lb

import (
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	AddServer(name string, svr http.Handler) error
	CheckServerStatus()
}

type loadBalancer struct {
	servers []http.Handler
	status  map[string]struct{}
	current uint64
}

// New creates a new LoadBalancer instance.
func New(ticker *time.Ticker) LoadBalancer {
	return &loadBalancer{
		servers: make([]http.Handler, 0),
		current: 0,
		status:  make(map[string]struct{}),
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

func (lb *loadBalancer) AddServer(name string, svr http.Handler) error {
	if svr == nil {
		return fmt.Errorf("server cannot be nil")
	}
	lb.servers = append(lb.servers, svr)
	return nil
}

func (lb *loadBalancer) CheckServerStatus() {
	t := time.NewTicker(2 * time.Minute)

	for range t.C {
		for name := range lb.status {
			alive := lb.checkServerStatus(name)
			if alive {
				lb.status[name] = struct{}{}
			} else {
				delete(lb.status, name)
			}
		}
	}
}

func (lb *loadBalancer) checkServerStatus(url string) bool {
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
