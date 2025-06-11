package rr

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// Backend server interface
type Backend interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Name() string
}

type backend struct {
	reverseProxy *httputil.ReverseProxy
	name         string
}

// New creates a new backend server that acts as a reverse proxy to the specified server URL.
func New(serverURL string) (Backend, error) {
	u, err := url.Parse(serverURL)
	if err != nil {
		return nil, err
	}

	return &backend{
		reverseProxy: httputil.NewSingleHostReverseProxy(u),
		name:         serverURL,
	}, nil
}

func (b *backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.reverseProxy.ServeHTTP(w, r)
}

func (b *backend) Name() string {
	return b.name
}
