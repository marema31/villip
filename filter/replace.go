package filter

import (
	"fmt"
	"bytes"
	"net/http"
	"strings"
	"io"
	"io/ioutil"
	"compress/gzip"

	"github.com/sirupsen/logrus"
)

func do(url string, s string, rep []replaceParameters) string {
	for _, r := range rep {
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

func (f *Filter) headerReplace(log *logrus.Entry, parsedHeader http.Header, headerConfig []header ) {
	log.Debug("Checking if need to replace header")

	for _, h := range headerConfig {
		if parsedHeader[h.Name] == nil || parsedHeader.Get(h.Name) == "" || h.Force {
			parsedHeader.Set(h.Name, h.Value)
			log.Debug(fmt.Sprintf("Set header %s with value :  %s", h.Name, h.Value))
		}
	}
}

func (f *Filter) replaceBody(requestURL string, rep []replaceParameters, body io.ReadCloser, bodyString string, parsedHeader http.Header) (int, io.ReadCloser, error) {
	var contentLength int

	
	f.log.Debug(fmt.Sprintf("Body before the replacement : %s", bodyString))

	bodyString = do(requestURL, bodyString, f.response.Replace)

	f.log.Debug(fmt.Sprintf("Body after the replacement : %s", bodyString))

	switch parsedHeader.Get("Content-Encoding") {
	case "gzip":
		w, err := f.compress(bodyString)
		if err != nil {
			return contentLength, body ,err
		}

		body = ioutil.NopCloser(w)
		contentLength = w.Len()

	default:
		buf := bytes.NewBufferString(bodyString)
		body = ioutil.NopCloser(buf)
		contentLength = buf.Len()
	}

	return contentLength, body, nil 
}


func (f *Filter) readBody(bod io.ReadCloser, h http.Header) (string, error) {
	var body io.ReadCloser

	switch h.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(bod)
		//		defer body.Close()
	default:
		body = bod
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	return string(b), err
}