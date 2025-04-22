package filter

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/sirupsen/logrus"
)

// UpdateRequest will be called back when the request is received by the proxy.
func (f *Filter) UpdateRequest(r *http.Request) {
	var contentLength int

	var originalBody string

	var modifiedBody string

	requestLog := f.log.WithFields(logrus.Fields{"url": r.URL.String(), "action": "request", "source": r.RemoteAddr})

	requestURL := strings.TrimPrefix(r.URL.String(), f.url)

	u, _ := url.Parse(f.url)
	r.URL.Host = u.Host
	r.Host = u.Host
	r.URL.Scheme = u.Scheme

	data, err := httputil.DumpRequest(r, false)
	if err != nil {
		requestLog.Error("Error")
	}

	requestLog.Debug(fmt.Sprintf("Request received\n%s", string(bytes.ReplaceAll(data, []byte{13, 10}, []byte{10}))))

	// in request sometimes there is no body
	if r.Body != nil {
		contentLength, r.Body, originalBody, modifiedBody, err =
			f.readAndReplaceBody(requestURL, f.request.Replace, r.Body, r.Header)

		if err != nil {
			requestLog.Fatal(err)
		}

		requestID := ""
		if f.dumpFolder != "" || len(f.dumpURLs) != 0 {
			requestID = f.dumpHTTPMessage(requestID, "", requestURL, r.Header, originalBody)
			r.Header.Set("X-VILLIP-Request-ID", requestID)
		}

		requestLog.WithFields(logrus.Fields{"requestID": requestID})

		r.Header["Content-Length"] = []string{fmt.Sprint(contentLength)}

		r.ContentLength = int64(contentLength)

		if requestID != "" {
			f.dumpHTTPMessage(requestID, "", requestURL, r.Header, modifiedBody)
		}
	}

	if len(f.request.Header) > 0 {
		f.headerReplace(requestLog, r.Header, f.request.Header)
	}
}
