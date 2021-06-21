package filter

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"time"

	"github.com/sirupsen/logrus"
)

type replaceParameters struct {
	from string
	to   string
	urls []*regexp.Regexp
}

type headerAction int

const (
	accept   headerAction = iota
	reject   headerAction = iota
	notEmpty headerAction = iota
)

type headerConditions struct {
	value  string
	action headerAction
}

type response struct {
	Replace []replaceParameters `yaml:"replace" json:"replace"`
	Header  []header            `yaml:"header" json:"header"`
}

type request struct {
	Replace []replaceParameters `yaml:"replace" json:"replace"`
	Header  []header            `yaml:"header" json:"header"`
}

//Filter proxifies an URL and filter the response.
type Filter struct {
	insecure     bool
	force        bool
	response     response
	request      request
	contentTypes []string
	restricted   []*net.IPNet
	token        map[string][]headerConditions
	url          string
	port         string
	log          *logrus.Entry
	dumpFolder   string
	dumpURLs     []*regexp.Regexp
}

//Serve starts a filtering http proxy.
func (f *Filter) Serve(res http.ResponseWriter, req *http.Request) {
	u, _ := url.Parse(f.url)

	proxy := httputil.NewSingleHostReverseProxy(u)
	if len(f.response.Replace) > 0 || len(f.response.Header) > 0 || f.dumpFolder != "" || len(f.dumpURLs) != 0 {
		proxy.ModifyResponse = f.UpdateResponse
	}

	if len(f.request.Replace) > 0 || len(f.request.Header) > 0 || f.dumpFolder != "" || len(f.dumpURLs) != 0 {
		proxy.Director = f.UpdateRequest
	}

	// Update the headers to allow for SSL redirection
	req.URL.Host = u.Host
	req.URL.Scheme = u.Scheme
	req.Header.Set("X-Forwarded-Host", req.Header.Get("Host"))
	req.Host = u.Host

	transport := http.Transport{
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second, //nolint: gomnd
	}

	if f.insecure {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint: gosec

		f.log.Debug("Not checking SSL certificates")
	}

	proxy.Transport = &transport

	f.log.Debug("proxying")
	proxy.ServeHTTP(res, req)
}
