package server

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
	filters []*filter.Filter
}

// New returns a new object Server.
func New(upLog logrus.FieldLogger, port string, f *filter.Filter) *Server {
	fs := make([]*filter.Filter, 0, 1)
	fs = append(fs, f)

	return &Server{
		port:    port,
		log:     upLog.WithField("port", port),
		filters: fs,
	}
}

// Insert a filter in the list.
func (s *Server) Insert(f *filter.Filter) {
	s.filters = append(s.filters, f)
}

func (s *Server) conditionalProxy(res http.ResponseWriter, req *http.Request) {
	sip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		s.log.WithFields(logrus.Fields{"userip": req.RemoteAddr}).Error("userip is not IP:port")
		http.Error(res, "Unable to parse source IP", http.StatusInternalServerError)
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

	mux.HandleFunc("/", s.conditionalProxy)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), mux)
	if err != nil {
		s.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}

	return err
}
