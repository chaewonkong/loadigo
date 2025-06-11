package rr

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type backend struct {
	reverseProxy *httputil.ReverseProxy
}

// New creates a new backend server that acts as a reverse proxy to the specified server URL.
func New(serverUrl string) (http.Handler, error) {
	u, err := url.Parse(serverUrl)
	if err != nil {
		return nil, err
	}

	return &backend{
		reverseProxy: httputil.NewSingleHostReverseProxy(u),
	}, nil
}

func (b *backend) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	b.reverseProxy.ServeHTTP(w, r)
}
