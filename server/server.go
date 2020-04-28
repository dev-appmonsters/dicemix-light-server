package server

import "net/http"

// Server - The main interface to enable connection.
type Server interface {
	Register(http.ResponseWriter, *http.Request)
}
