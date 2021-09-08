package server

import "github.com/marema31/villip/filter"

// Server interface allowing different protocols.
type Server interface {
	Serve() error
	Insert(f filter.FilteredServer)
}
