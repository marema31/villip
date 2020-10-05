package filter


import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
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

	if r.Body != nil {
		s, err := f.readBody(r.Body, r.Header)
		if err != nil {
			f.log.Fatal(err)
		}

		switch r.Header.Get("Content-Encoding") {
		case "gzip":
			w, _ := f.compress(s)

			r.Body = ioutil.NopCloser(w)
			r.Header["Content-Length"] = []string{fmt.Sprint(w.Len())}

		default:
			buf := bytes.NewBufferString(s)
			r.Body = ioutil.NopCloser(buf)
			r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
		}
	}

	if len(f.request.Header) > 0 {
		f.headerReplace(requestLog, &r.Header, "request")
	}
}