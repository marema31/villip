package tcp

import (
	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
)

// Server will manage proxing for one port using one or more filter.
type Server struct {
	port   string
	log    logrus.FieldLogger
	filter filter.FilteredServer
}

// New returns a new object Server.
func New(upLog logrus.FieldLogger, port string, f filter.FilteredServer) *Server {
	return &Server{
		port:   port,
		log:    upLog.WithField("port", port),
		filter: f,
	}
}

// Insert a filter in the list.
func (s *Server) Insert(f filter.FilteredServer) {
	s.log.Fatal("Cannot have several filters to the same port for raw proxy")
}

// Serve listens to the port and call the correct filter.
func (s *Server) Serve() error {
	err := s.filter.ServeTCP()
	if err != nil {
		s.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}

	return err
}
