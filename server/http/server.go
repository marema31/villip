package http

import (
	"fmt"
	"net"
	"net/http"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
)

// Server will manage proxing for one port using one or more filter.
type Server struct {
	port    string
	log     logrus.FieldLogger
	filters []filter.FilteredServer
}

// New returns a new object Server.
func New(upLog logrus.FieldLogger, port string, f filter.FilteredServer) *Server {
	fs := make([]filter.FilteredServer, 0, 1)
	fs = append(fs, f)

	return &Server{
		port:    port,
		log:     upLog.WithField("port", port),
		filters: fs,
	}
}

// Insert a filter in the list.
func (s *Server) Insert(f filter.FilteredServer) {
	s.filters = append(s.filters, f)
}

// ConditionalProxy will call the corresponding filter proxy handler.
func (s *Server) ConditionalProxy(res http.ResponseWriter, req *http.Request) {
	sip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		s.log.WithFields(logrus.Fields{"userip": req.RemoteAddr}).Error("userip is not IP:port")
		http.Error(res, "Unable to parse source IP", http.StatusInternalServerError)

		return
	}

	ip := net.ParseIP(sip)

	for _, f := range s.filters {
		if f.IsConcerned(ip, req.Header) {
			f.Serve(res, req)

			return
		}
	}

	http.Error(res, "No filter correspond to this requests", http.StatusNotFound)
}

// Serve listens to the port and call the correct filter.
func (s *Server) Serve() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", s.ConditionalProxy)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), mux)
	if err != nil {
		s.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}

	return err
}
