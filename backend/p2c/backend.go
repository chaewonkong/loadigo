package p2c

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync/atomic"

	"github.com/chaewonkong/loadigo/backend"
)

type Backend struct {
	http.Handler
	name string

	// inflight tracks the number of inflight requests to this backend.
	inflight atomic.Int64
}

// NewBackend creates a new backend server that acts as a reverse proxy to the specified server URL.
func NewBackend(serverURL string) (backend.Backend, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &Backend{
		Handler: httputil.NewSingleHostReverseProxy(u),
		name:    serverURL,
	}, nil
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Increment the inflight counter when a request is received.
	// Decrement it when the request is done.
	b.inflight.Add(1)
	defer b.inflight.Add(-1)

	b.Handler.ServeHTTP(w, r)
}

func (b *Backend) Name() string {
	return b.name
}

func (b *Backend) Inflight() int64 {
	return b.inflight.Load()
}
