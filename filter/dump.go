package filter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
)

// Mockable generateID for unit test.
var _generateID = generateID //nolint: gochecknoglobals

func generateID() (string, error) {
	r := make([]byte, 12)

	if _, err := rand.Read(r); err != nil {
		return "", err
	}

	return hex.EncodeToString(r), nil
}

func sortHeader(header http.Header) []string {
	sortedKeys := make([]string, 0, len(header))
	for name := range header {
		sortedKeys = append(sortedKeys, name)
	}

	sort.Strings(sortedKeys)

	sortedHeaders := make([]string, 0, len(header))

	for _, name := range sortedKeys {
		for _, value := range header[name] {
			sortedHeaders = append(sortedHeaders, fmt.Sprintf("%s: %s\n", name, value))
		}
	}

	return sortedHeaders
}

func (f *Filter) dumpToFile(fileType string, requestID string, url string, header http.Header, body string) string {
	fileName := filepath.Join(f.dumpFolder, fmt.Sprintf("%s.%s", requestID, fileType))

	file, err := os.Create(fileName)
	if err != nil {
		f.log.Fatalf("Failed to create %s: %v", fileName, err)
	}
	defer file.Close()

	if _, err := fmt.Fprintf(file, "URL: %s\n", url); err != nil {
		f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
	}

	for _, value := range sortHeader(header) {
		if _, err := file.WriteString(value); err != nil {
			f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
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

func (f *Filter) dumpToLog(fileType string, requestID string, url string, header http.Header, body string) string {
	log := f.log.WithFields(logrus.Fields{"response-type": fileType, "requestID": requestID, "url": url})

	for _, value := range sortHeader(header) {
		log.WithField("part", "header").Debug(value)
	}

	log.WithField("part", "body").Debug(body)

	return requestID
}

func (f *Filter) dumpHTTPMessage(
	requestID string,
	requestIDFromRequest string,
	url string,
	header http.Header,
	body string,
) string {
	var httpMessageType string

	if header.Get("Server") != "" {
		httpMessageType = "Response"
	} else {
		httpMessageType = "Request"
	}

	fileType := "filtered" + httpMessageType

	if requestID == "" {
		rID, err := _generateID()
		if err != nil {
			f.log.Fatalf("Failed to generate requestId: %v", err)
		}

		if requestIDFromRequest == "" {
			requestID = rID
		} else {
			requestID = requestIDFromRequest
		}

		fileType = "original" + httpMessageType
	}

	if len(f.dumpURLs) != 0 {
		found := false

		for _, reg := range f.dumpURLs {
			if reg.MatchString(url) {
				found = true

				break
			}
		}

		if !found {
			return requestID
		}
	}

	if f.dumpFolder != "" {
		return f.dumpToFile(fileType, requestID, url, header, body)
	}

	return f.dumpToLog(fileType, requestID, url, header, body)
}
