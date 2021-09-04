package server

import "github.com/marema31/villip/filter"

type Server interface {
	Serve() error
	Insert(f filter.FilteredServer)
}
