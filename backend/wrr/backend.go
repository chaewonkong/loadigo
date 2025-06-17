package wrr

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/chaewonkong/loadigo/backend"
)

type Backend struct {
	http.Handler
	name string

	weight   float64
	deadline float64
}

// NewBackend creates a new backend server that acts as a reverse proxy to the specified server URL.
func NewBackend(serverURL string, weight float64) (backend.Backend, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	if weight <= 0 {
		return nil, fmt.Errorf("deadline must be greater than 0, got %f", weight)
	}

	return &Backend{
		Handler: httputil.NewSingleHostReverseProxy(u),
		name:    serverURL,
		weight:  weight,
	}, nil
}

func (b *Backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.Handler.ServeHTTP(w, r)
}

func (b *Backend) Name() string {
	return b.name
}
