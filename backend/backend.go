package backend

import "net/http"

// Backend server interface
type Backend interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
	Name() string
}
