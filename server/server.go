package server

import (
	"fmt"
	"net"
	"net/http"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
)

//Server will manage proxing for one port using one or more filter.
type Server struct {
	port    string
	log     *logrus.Entry
	filters []*filter.Filter
}

//New returns a new object Server.
func New(upLog *logrus.Entry, port string, f *filter.Filter) *Server {
	fs := make([]*filter.Filter, 0, 1)
	fs = append(fs, f)

	return &Server{
		port:    port,
		log:     upLog.WithField("port", port),
		filters: fs,
	}
}

//Insert a filter in the list, the filter without condition will be the last one.
func (s *Server) Insert(f *filter.Filter) {
	if !f.IsConditional() {
		s.filters = append(s.filters, f)
	} else {
		// Prepending filter to the list using golang tricks
		s.filters = append(s.filters, &filter.Filter{})
		copy(s.filters[1:], s.filters)
		s.filters[0] = f
	}
}

func (s *Server) conditionalProxy(res http.ResponseWriter, req *http.Request) {
	sip, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		s.log.WithFields(logrus.Fields{"userip": req.RemoteAddr}).Error("userip is not IP:port")
		http.Error(res, "Unable to parse source IP", 500)
	}

	ip := net.ParseIP(sip)

	for _, f := range s.filters {
		if f.IsConcerned(ip, req.Header) {
			f.Serve(res, req)
			return
		}
	}

	http.Error(res, "No filter correspond to this requests", 404)
}

//Serve listens to the port and call the correct filter.
func (s *Server) Serve() {
	http.HandleFunc("/", s.conditionalProxy)

	err := http.ListenAndServe(fmt.Sprintf(":%s", s.port), nil)
	if err != nil {
		s.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}
}
