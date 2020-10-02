package filter

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"

	"github.com/sirupsen/logrus"
)

type replaceParameters struct {
	from string
	to   string
	urls []*regexp.Regexp
}

type response struct {
	Replace  []replaceParameters `yaml:"replace" json:"replace"`
	Header	 []header	   `yaml:"header" json:"header"`
}

type request struct {
	Replace  []replaceParameters `yaml:"replace" json:"replace"`
	Header	 []header	   `yaml:"header" json:"header"`
}

//Filter proxifies an URL and filter the response.
type Filter struct {
	force        bool
	response	 response
	request      request
	contentTypes []string
	restricted   []*net.IPNet
	url          string
	port         string
	log          *logrus.Entry
	dumpFolder   string
	dumpURLs     []*regexp.Regexp
}

func (f *Filter) startLog() {
	f.log.Info(fmt.Sprintf("Listen on port %s", f.port))
	f.log.Info(fmt.Sprintf("Will filter responses from %s", f.url))

	if len(f.restricted) != 0 {
		f.log.Info(fmt.Sprintf("Only for request from: %s ", f.restricted))
	}

	f.log.Info(fmt.Sprintf("For content-type %s", f.contentTypes))
	f.log.Info("And replace:")

	for _, r := range f.response.Replace {
		f.log.Info(fmt.Sprintf("   %s  by  %s", r.from, r.to))

		if len(r.urls) != 0 {
			var us []string

			for _, u := range r.urls {
				us = append(us, u.String())
			}

			f.log.Info(fmt.Sprintf("    for %v", us))
		}
	}
}


//Serve starts a filtering http proxy.
func (f *Filter) Serve() {
	u, _ := url.Parse(f.url)

	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = f.UpdateResponse
	proxy.Director = f.UpdateRequest

	mx := http.NewServeMux()
	mx.Handle("/", proxy)


	err := http.ListenAndServe(fmt.Sprintf(":%s", f.port), mx)
	if err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}
}
