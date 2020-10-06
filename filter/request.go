package filter

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

//UpdateRequest will be called back when the request is received by the proxy.
func (f *Filter) UpdateRequest(r *http.Request) {
	var contentLength int

	requestLog := f.log.WithFields(logrus.Fields{"url": r.URL.String(), "action": "request", "source": r.RemoteAddr})
	requestLog.Debug("Request")

	u, _ := url.Parse(f.url)
	r.URL.Host = u.Host
	r.Host = u.Host
	r.URL.Scheme = u.Scheme

	data, err := httputil.DumpRequest(r, false)
	if err != nil {
		f.log.Error(fmt.Printf("Error"))
	}

	f.log.Debug(fmt.Sprintf("Request received\n %s", string(data)))
	requestURL := strings.TrimPrefix(r.URL.String(), f.url)

	s, err := f.readBody(r.Body, r.Header)

	if err != nil {
		f.log.Fatal(err)
	}

	if r.Body != nil && len(f.request.Replace) > 0 {
		contentLength, r.Body, err = f.replaceBody(requestURL, f.request.Replace, r.Body, s, r.Header)

		if err != nil {
			f.log.Fatal(err)
		}

		f.log.Info(contentLength)
		r.ContentLength = int64(contentLength)
	}

	if len(f.request.Header) > 0 {
		f.headerReplace(requestLog, r.Header, f.request.Header)
	}
}
