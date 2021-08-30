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

// Mockable function.
var _do = do //nolint: gochecknoglobals

func do(url string, s string, rep []replaceParameters, prefix bool) string {
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

		if prefix {
			if len(s) > len(r.from) {
				s = r.to + s[len(r.from):]
			}
		} else {
			s = strings.Replace(s, r.from, r.to, -1)
		}
	}

	return s
}

func (f *Filter) headerReplace(log logrus.FieldLogger, parsedHeader http.Header, headerConfig []Cheader) {
	log.Debug("Checking if need to replace header")

	for _, h := range headerConfig {
		if h.Add {
			if parsedHeader.Get(h.Name) == "" {
				parsedHeader[h.Name] = []string{h.Value}
			} else {
				parsedHeader[h.Name] = append(parsedHeader[h.Name], h.Value)
			}

			log.Debug(fmt.Sprintf("Adding to header %s with value :  %s", h.Name, h.Value))

			continue
		}

		if parsedHeader.Get(h.Name) == "" || h.Force {
			// parsedHeader.Set(h.Name, h.Value) // Use CanonicalMIMEHeaderKey that modify the key
			parsedHeader[h.Name] = []string{h.Value}
			log.Debug(fmt.Sprintf("Set header %s with value :  %s", h.Name, h.Value))
		}
	}
}

func (f *Filter) readAndReplaceBody(
	requestURL string,
	rep []replaceParameters,
	bod io.ReadCloser,
	parsedHeader http.Header,
) (int, io.ReadCloser, string, string, error) {
	var (
		originalBody  string
		modifiedBody  string
		contentLength int
		body          io.ReadCloser
		err           error
	)

	switch parsedHeader.Get("Content-Encoding") {
	case "gzip":
		body, err = gzip.NewReader(bod)
		if err != nil {
			f.log.Errorf("Impossible to decompress: %v", err)

			return 0, nil, "", "", err
		}
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

	modifiedBody = _do(requestURL, originalBody, rep, false)

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

func (f *Filter) PrefixReplace(URL string) string {
	return do(URL, URL, f.prefix, true)
}
