package filter

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"net/http/httputil"
	"net/url"
	"github.com/sirupsen/logrus"
)

func (f *Filter) do(url string, s string) string {
	for _, r := range f.response.Replace {
		if len(r.urls) != 0 {
			found := false

			for _, reg := range r.urls {
				if reg.MatchString(url) {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		s = strings.Replace(s, r.from, r.to, -1)
	}

	return s
}

//UpdateResponse will be called back when the proxyfied server respond and filter the response if necessary.
func (f *Filter) UpdateResponse(r *http.Response) error {
	requestLog := f.log.WithFields(logrus.Fields{"url": r.Request.URL.String(), "status": r.StatusCode, "source": r.Request.RemoteAddr})
	// The Request in the Response is the last URL the client tried to access.
	requestLog.Debug("Request")

	authorized, err := f.isAuthorized(requestLog, r)
	if err != nil || !authorized {
		return err
	}

	if !f.force && !f.toFilter(requestLog, r) {
		return nil
	}

	requestLog.Debug("filtering")

	s, err := f.readBody(r)
	if err != nil {
		return err
	}

	requestURL := strings.TrimPrefix(r.Request.URL.String(), f.url)

	requestID := ""
	if f.dumpFolder != "" || len(f.dumpURLs) != 0 {
		requestID = f.dumpResponse(requestID, requestURL, r.Header, s)
	}

	s = f.do(requestURL, s)

	requestLog.WithFields(logrus.Fields{"requestID": requestID}).Debug("will rewrite content")

	f.location(requestLog, r, requestURL)

	if requestID != "" {
		f.dumpResponse(requestID, requestURL, r.Header, s)
	}

	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		w, err := f.compress(s)
		if err != nil {
			return err
		}

		r.Body = ioutil.NopCloser(w)
		r.Header["Content-Length"] = []string{fmt.Sprint(w.Len())}

	default:
		buf := bytes.NewBufferString(s)
		r.Body = ioutil.NopCloser(buf)
		r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	}

	return nil
}

func (f *Filter) UpdateRequest(r *http.Request) {
	u, _ := url.Parse(f.url)
	r.URL.Host = u.Host
	r.URL.Scheme = "http"
	data, err := httputil.DumpRequest(r, false)
	if err != nil {
		f.log.Error(fmt.Printf("Error"))
	}
	f.log.Info(fmt.Sprintf("%s", string(data)))
	//r.Header.Set("X-OVH-Gateway-Source", "titi")
}

//nolint: nestif
func (f *Filter) isAuthorized(log *logrus.Entry, r *http.Response) (bool, error) {
	if len(f.restricted) != 0 {
		sip, _, err := net.SplitHostPort(r.Request.RemoteAddr)
		if err != nil {
			log.WithFields(logrus.Fields{"userip": r.Request.RemoteAddr}).Error("userip is not IP:port")
			return true, err
		}

		ip := net.ParseIP(sip)
		if !ip.IsLoopback() {
			seen := false

			for _, ipnet := range f.restricted {
				if ipnet.Contains(ip) {
					seen = true
					break
				}
			}

			if !seen {
				log.WithFields(logrus.Fields{"source": ip}).Debug("forbidden from this IP")

				buf := bytes.NewBufferString("Access forbidden from this IP")
				r.Body = ioutil.NopCloser(buf)
				r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
				r.StatusCode = http.StatusForbidden

				return false, nil
			}
		}
	}

	return true, nil
}

func (f *Filter) toFilter(log *logrus.Entry, r *http.Response) bool {
	if r.StatusCode == http.StatusOK {
		currentType := r.Header.Get("Content-Type")

		for _, testedType := range f.contentTypes {
			if strings.Contains(currentType, testedType) {
				return true
			}
		}

		log.WithFields(logrus.Fields{"type": currentType}).Debug("... skipping type")

		return false
	} else if r.StatusCode != http.StatusFound && r.StatusCode != http.StatusMovedPermanently {
		log.Debug("... skipping status")
		return false
	}

	return true
}

func (f *Filter) readBody(r *http.Response) (string, error) {
	var body io.ReadCloser

	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(r.Body)
		//		defer body.Close()
	default:
		body = r.Body
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(b), err
}

func (f *Filter) compress(s string) (*bytes.Buffer, error) {
	var w bytes.Buffer

	compressed := gzip.NewWriter(&w)

	_, err := compressed.Write([]byte(s))
	if err != nil {
		return nil, err
	}

	err = compressed.Flush()
	if err != nil {
		return nil, err
	}

	err = compressed.Close()
	if err != nil {
		return nil, err
	}

	return &w, nil
}

func (f *Filter) location(requestLog *logrus.Entry, r *http.Response, requestURL string) {
	location := r.Header.Get("Location")
	if location != "" {
		origLocation := location
		location = f.do(requestURL, location)

		requestLog.WithFields(logrus.Fields{"location": origLocation, "rewrited_location": location}).Debug("will rewrite location header")
		r.Header.Set("Location", location)
	}
}
