package filter

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/sirupsen/logrus"
)

//Filter proxifies an URL and filter the response
type Filter struct {
	force bool
	//Froms strings to be replaced
	froms []string
	//Tos replacement string
	tos []string
	//ContentTypes list of content types that will be filtered
	contentTypes []string
	//Restricted list of net ranges allowed to connect to villip
	restricted []*net.IPNet
	url        string
	port       string
	log        *logrus.Entry
}

func (f *Filter) startLog() {
	f.log.Info(fmt.Sprintf("Listen on port %s", f.port))
	f.log.Info(fmt.Sprintf("Will filter responses from %s", f.url))
	if len(f.restricted) != 0 {
		f.log.Info(fmt.Sprintf("Only for request from: %s ", f.restricted))
	}
	f.log.Info(fmt.Sprintf("For content-type %s", f.contentTypes))
	f.log.Info("And replace:")
	for i := range f.froms {
		f.log.Info(fmt.Sprintf("   %s  by  %s", f.froms[i], f.tos[i]))
	}

}

//Serve starts a filtering http proxy
func (f *Filter) Serve() {
	u, _ := url.Parse(f.url)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = f.UpdateResponse

	mx := http.NewServeMux()
	mx.Handle("/", proxy)

	err := http.ListenAndServe(fmt.Sprintf(":%s", f.port), mx)
	if err != nil {
		f.log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}
}
