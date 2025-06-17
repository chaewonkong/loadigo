package wrr

import (
	"container/heap"
	"fmt"
	"log"
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
	servers []*Backend
	status  map[string]struct{}
	current uint64
	ticker  *time.Ticker
	mu      sync.RWMutex

	// curDeadline is used to track the current deadline for server checks.
	curDeadline float64
}

// NewLoadBalancer creates a new LoadBalancer instance.
func NewLoadBalancer(ticker *time.Ticker) LoadBalancer {
	return &loadBalancer{
		servers: make([]*Backend, 0),
		current: 0,
		status:  make(map[string]struct{}),
		ticker:  ticker,
		mu:      sync.RWMutex{},
	}
}

// heap.Interface implementation for loadBalancer
func (lb *loadBalancer) Len() int {
	return len(lb.servers)
}

func (lb *loadBalancer) Less(i, j int) bool {
	return lb.servers[i].deadline < lb.servers[j].deadline
}

func (lb *loadBalancer) Swap(i, j int) {
	lb.servers[i], lb.servers[j] = lb.servers[j], lb.servers[i]
}

func (lb *loadBalancer) Push(x interface{}) {
	b, ok := x.(*Backend)
	if !ok {
		return
	}

	lb.servers = append(lb.servers, b)
}

func (lb *loadBalancer) Pop() interface{} {
	if len(lb.servers) == 0 {
		return nil
	}

	b := lb.servers[len(lb.servers)-1]
	lb.servers = lb.servers[:len(lb.servers)-1]

	return b
}

func (lb *loadBalancer) nextServer() http.Handler {
	if len(lb.servers) == 0 {
		return nil
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	var b *Backend
	for {
		b = heap.Pop(lb).(*Backend)

		lb.curDeadline = b.deadline
		b.deadline += 1 / (b.weight)
		heap.Push(lb, b)

		if _, ok := lb.status[b.Name()]; ok {
			break
		}
	}

	return b
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

	b, ok := svr.(*Backend)
	if !ok {
		return fmt.Errorf("server must be of type *Backend")
	}

	lb.mu.Lock()
	defer lb.mu.Unlock()

	// set initial deadline for the server
	b.deadline = lb.curDeadline + 1/(b.weight)
	heap.Push(lb, b)

	lb.status[name] = struct{}{}

	lb.servers = append(lb.servers, b)
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
