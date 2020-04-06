package filter

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func generateID() (string, error) {
	r := make([]byte, 12)
	_, err := rand.Read(r)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(r), nil
}

func (f *Filter) dumpResponse(requestID string, header http.Header, body string) string {
	fileName := fmt.Sprintf("%s.filtered", requestID)
	if requestID == "" {
		rID, err := generateID()
		if err != nil {
			f.log.Fatalf("Failed to generate requestId: %v", err)
		}
		requestID = rID
		fileName = fmt.Sprintf("%s.original", requestID)
	}

	fileName = filepath.Join(f.dumpFolder, fileName)
	file, err := os.Create(fileName)
	if err != nil {
		f.log.Fatalf("Failed to create %s: %v", fileName, err)
	}
	defer file.Close()

	for name, values := range header {
		for _, value := range values {
			if _, err := file.WriteString(fmt.Sprintf("%s: %s\n", name, value)); err != nil {
				f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
			}
		}
	}

	if _, err := file.WriteString("\n"); err != nil {
		f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
	}
	if _, err := file.WriteString(body); err != nil {
		f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
	}

	return requestID
}

//UpdateResponse will be called back when the proxyfied server respond and filter the response if necessary
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

	var b []byte

	var body io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(r.Body)
		//		defer body.Close()
	default:
		body = r.Body
	}

	b, err = ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	s := string(b)

	requestID := ""
	if f.dumpFolder != "" {
		requestID = f.dumpResponse(requestID, r.Header, s)
	}

	for i := range f.froms {
		s = strings.Replace(s, f.froms[i], f.tos[i], -1)
	}

	requestLog.WithFields(logrus.Fields{"requestID": requestID}).Debug("will rewrite content")

	location := r.Header.Get("Location")
	if location != "" {
		origLocation := location
		for i := range f.froms {
			location = strings.Replace(location, f.froms[i], f.tos[i], -1)
		}
		requestLog.WithFields(logrus.Fields{"location": origLocation, "rewrited_location": location}).Debug("will rewrite location header")
		r.Header.Set("Location", location)
	}

	if requestID != "" {
		f.dumpResponse(requestID, r.Header, s)
	}

	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		var w bytes.Buffer

		compressed := gzip.NewWriter(&w)

		_, err := compressed.Write([]byte(s))
		if err != nil {
			return err
		}

		err = compressed.Flush()
		if err != nil {
			return err
		}
		err = compressed.Close()
		if err != nil {
			return err
		}

		r.Body = ioutil.NopCloser(&w)

		r.Header["Content-Length"] = []string{fmt.Sprint(w.Len())}

	default:
		buf := bytes.NewBufferString(s)
		r.Body = ioutil.NopCloser(buf)
		r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	}

	return nil
}

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
				buf := bytes.NewBufferString("Access forbiden from this IP")
				r.Body = ioutil.NopCloser(buf)
				r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
				r.StatusCode = 403
				return false, nil
			}
		}
	}
	return true, nil
}

func (f *Filter) toFilter(log *logrus.Entry, r *http.Response) bool {
	if r.StatusCode == 200 {
		currentType := r.Header.Get("Content-Type")
		for _, testedType := range f.contentTypes {
			if strings.Contains(currentType, testedType) {
				return true
			}
		}
		log.WithFields(logrus.Fields{"type": currentType}).Debug("... skipping type")
		return false

	} else if r.StatusCode != 302 && r.StatusCode != 301 {
		log.Debug("... skipping status")
		return false
	}
	return true
}
