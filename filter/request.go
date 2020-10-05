package filter

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

//UpdateRequest will be called back when the request is received by the proxy.
func (f *Filter) UpdateRequest(r *http.Request) {
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

	if r.Body != nil && len(f.request.Replace) > 0 {
		s, err := f.readBody(r.Body, r.Header)

		if err != nil {
			f.log.Fatal(err)
		}

		f.log.Debug(fmt.Sprintf("Body of the before replacement : %s", s))
		requestURL := strings.TrimPrefix(r.URL.String(), f.url)
		s = do(requestURL, s, &f.request.Replace)

		switch r.Header.Get("Content-Encoding") {
		case "gzip":
			w, _ := f.compress(s)

			r.Body = ioutil.NopCloser(w)
			r.ContentLength = int64(w.Len())

		default:
			buf := bytes.NewBufferString(s)
			r.Body = ioutil.NopCloser(buf)
			r.ContentLength = int64(buf.Len())
		}

		f.log.Debug(fmt.Sprintf("Body of the request after replacement : %s", s))
	}

	if len(f.request.Header) > 0 {
		f.headerReplace(requestLog, &r.Header, "request")
	}
}
