package filter

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

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

func (f *Filter) headerReplace(log *logrus.Entry, parsedHeader http.Header, headerConfig []header) {
	log.Debug("Checking if need to replace header")

	for _, h := range headerConfig {
		if parsedHeader.Get(h.Name) == "" || h.Force {
			parsedHeader.Set(h.Name, h.Value)
			log.Debug(fmt.Sprintf("Set header %s with value :  %s", h.Name, h.Value))
		}
	}
}

func (f *Filter) readAndReplaceBody(requestURL string, rep []replaceParameters, bod io.ReadCloser, parsedHeader http.Header) (int, io.ReadCloser, string, string, error) {
	var originalBody string

	var modifiedBody string

	var contentLength int

	var body io.ReadCloser

	switch parsedHeader.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(bod)
		//		defer body.Close()
	default:
		body = bod
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return 0, nil, "", "", err
	}

	originalBody = string(b)

	f.log.Debug(fmt.Sprintf("Body before the replacement : %s", originalBody))

	modifiedBody = do(requestURL, originalBody, rep)

	f.log.Debug(fmt.Sprintf("Body after the replacement : %s", modifiedBody))

	switch parsedHeader.Get("Content-Encoding") {
	case "gzip":
		w, err := f.compress(modifiedBody)
		if err != nil {
			return 0, nil, "", "", err
		}

		body = ioutil.NopCloser(w)
		contentLength = w.Len()

	default:
		buf := bytes.NewBufferString(modifiedBody)
		body = ioutil.NopCloser(buf)
		contentLength = buf.Len()
	}

	return contentLength, body, originalBody, modifiedBody, nil
}
