package filter

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

// UpdateResponse will be called back when the proxyfied server respond and filter the response if necessary.
func (f *Filter) UpdateResponse(r *http.Response) error {
	var (
		contentLength int
		originalBody  string
		modifiedBody  string
		err           error
	)

	requestLog := f.log.WithFields(
		logrus.Fields{
			"url":    r.Request.URL.String(),
			"action": "response",
			"status": r.StatusCode,
			"source": r.Request.RemoteAddr,
		})
	// The Request in the Response is the last URL the client tried to access.
	requestURL := strings.TrimPrefix(r.Request.URL.String(), f.url)

	if !f.force && !f.toFilter(requestLog, r) {
		return nil
	}

	if r.Body != nil {
		requestLog.Debug("filtering")

		contentLength, r.Body, originalBody, modifiedBody, err =
			f.readAndReplaceBody(requestURL, f.response.Replace, r.Body, r.Header)

		if err != nil {
			return err
		}

		f.log.Info(r.Request.Header.Get("X-VILLIP-Request-ID"))

		requestID := ""

		if f.dumpFolder != "" || len(f.dumpURLs) != 0 {
			requestID = f.dumpHTTPMessage(
				requestID,
				r.Request.Header.Get("X-VILLIP-Request-ID"),
				requestURL,
				r.Header,
				originalBody,
			)
		}

		requestLog.WithFields(logrus.Fields{"requestID": requestID})
		f.location(requestLog, r, requestURL)
		r.Header["Content-Length"] = []string{fmt.Sprint(contentLength)}

		if requestID != "" {
			f.dumpHTTPMessage(requestID, "", requestURL, r.Header, modifiedBody)
		}
	}

	if len(f.response.Header) > 0 {
		f.headerReplace(requestLog, r.Header, f.response.Header)
	}

	return nil
}

func (f *Filter) toFilter(log logrus.FieldLogger, r *http.Response) bool {
	if r.StatusCode == http.StatusOK {
		currentType := r.Header.Get("Content-Type")

		for _, testedType := range f.contentTypes {
			if strings.Contains(currentType, testedType) {
				return true
			}
		}

		log.WithFields(logrus.Fields{"type": currentType}).Debug("... skipping type")

		return false
	}

	for _, status := range f.status {
		if r.StatusCode == status {
			return true
		}
	}

	log.Debug("... skipping status")

	return false
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

func (f *Filter) location(requestLog logrus.FieldLogger, r *http.Response, requestURL string) {
	location := r.Header.Get("Location")

	if location != "" {
		origLocation := location
		location = do(requestURL, location, f.response.Replace, false)

		requestLog.
			WithFields(logrus.Fields{"location": origLocation, "rewrited_location": location}).
			Debug("will rewrite location header")
		r.Header.Set("Location", location)
	}
}
