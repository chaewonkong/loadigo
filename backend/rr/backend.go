package rr

import (
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/chaewonkong/loadigo/backend"
)

type Backend struct {
	http.Handler
	name string
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
	b.Handler.ServeHTTP(w, r)
}

func (b *Backend) Name() string {
	return b.name
}
