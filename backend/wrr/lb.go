package wrr

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chaewonkong/loadigo/backend"
)

// LoadBalancer defines the interface for a load balancer that can distribute requests
type LoadBalancer interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	AddServer(name string, svr backend.Backend) error
	CheckServerStatus()
}

type loadBalancer struct {
	servers []backend.Backend
	status  map[string]struct{}
	current uint64
	ticker  *time.Ticker
	mu      sync.RWMutex
}

// NewLoadBalancer creates a new LoadBalancer instance.
func NewLoadBalancer(ticker *time.Ticker) LoadBalancer {
	return &loadBalancer{
		servers: make([]backend.Backend, 0),
		current: 0,
		status:  make(map[string]struct{}),
		ticker:  ticker,
		mu:      sync.RWMutex{},
	}
}

func (lb *loadBalancer) nextServer() http.Handler {
	if len(lb.servers) == 0 {
		return nil
	}

	for i := 0; i < len(lb.servers); i++ {
		idx := atomic.AddUint64(&lb.current, 1)
		svr := lb.servers[int(idx)%len(lb.servers)]
		lb.mu.RLock()
		_, ok := lb.status[svr.Name()]
		lb.mu.RUnlock()
		if ok {
			return svr
		}
	}

	return nil
}

func (lb *loadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	svr := lb.nextServer()
	if svr == nil {
		http.Error(w, "No backend available", http.StatusServiceUnavailable)
		return
	}

	svr.ServeHTTP(w, r)
}

func (lb *loadBalancer) AddServer(name string, svr backend.Backend) error {
	if svr == nil {
		return fmt.Errorf("server cannot be nil")
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.status[name] = struct{}{}

	lb.servers = append(lb.servers, svr)
	return nil
}

func (lb *loadBalancer) CheckServerStatus() {
	for range lb.ticker.C {
		for name := range lb.status {
			alive := lb.checkServerStatus(name)
			if alive {
				lb.mu.Lock()
				lb.status[name] = struct{}{}
				lb.mu.Unlock()
			} else {
				lb.mu.Lock()
				delete(lb.status, name)
				lb.mu.Unlock()
				log.Printf("Server %s is down, removing from status", name)
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
