package p2c

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
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
	servers   []*Backend
	status    map[string]struct{}
	current   uint64
	ticker    *time.Ticker
	backendMu sync.RWMutex
	randMu    sync.Mutex

	randInt func(n int) int
}

// NewLoadBalancer creates a new LoadBalancer instance.
func NewLoadBalancer(ticker *time.Ticker) LoadBalancer {
	return &loadBalancer{
		servers:   make([]*Backend, 0),
		current:   0,
		status:    make(map[string]struct{}),
		ticker:    ticker,
		backendMu: sync.RWMutex{},
		randMu:    sync.Mutex{},
		randInt:   rand.New(rand.NewSource(time.Now().UnixNano())).Intn,
	}
}

func (lb *loadBalancer) nextServer() http.Handler {
	if len(lb.servers) == 0 {
		return nil
	}

	healthy := []*Backend{}
	lb.backendMu.RLock()
	for _, b := range lb.servers {
		if _, ok := lb.status[b.Name()]; ok {
			healthy = append(healthy, b)
		}
	}
	lb.backendMu.RUnlock()
	if len(healthy) == 0 {
		return nil
	}

	if len(healthy) == 1 {
		return healthy[0]
	}

	// p2c
	lb.randMu.Lock()
	// randInt returns a random integer between 0 and n-1
	i, j := lb.randInt(len(healthy)), lb.randInt(len(healthy))
	lb.randMu.Unlock()

	// Ensure b1 and b2 are different
	if i == j {
		i = (j + 1) % len(healthy)
	}

	b1, b2 := healthy[i], healthy[j]

	// check load (inflight requests) and return backend with fewer inflight requests
	if b1.Inflight() < b2.Inflight() {
		return b1
	}

	return b2
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

	backend, ok := svr.(*Backend)
	if !ok {
		return fmt.Errorf("server must be of type *p2c.Backend")
	}

	lb.backendMu.Lock()
	defer lb.backendMu.Unlock()
	lb.status[name] = struct{}{}

	lb.servers = append(lb.servers, backend)
	return nil
}

func (lb *loadBalancer) CheckServerStatus() {
	for range lb.ticker.C {
		for name := range lb.status {
			alive := lb.checkServerStatus(name)
			if alive {
				lb.backendMu.Lock()
				lb.status[name] = struct{}{}
				lb.backendMu.Unlock()
			} else {
				lb.backendMu.Lock()
				delete(lb.status, name)
				lb.backendMu.Unlock()
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
