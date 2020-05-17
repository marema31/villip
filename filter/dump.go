package filter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

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

func (f *Filter) dumpToFile(fileType string, requestID string, url string, header http.Header, body string) string {
	fileName := filepath.Join(f.dumpFolder, fmt.Sprintf("%s.%s", requestID, fileType))

	file, err := os.Create(fileName)
	if err != nil {
		f.log.Fatalf("Failed to create %s: %v", fileName, err)
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf("URL: %s\n", url)); err != nil {
		f.log.Fatalf("Failed to write header in %s: %v", requestID, err)
	}

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

func (f *Filter) dumpToLog(fileType string, requestID string, url string, header http.Header, body string) string {
	log := f.log.WithFields(logrus.Fields{"response-type": fileType, "requestID": requestID, "url": url})

	for name, values := range header {
		for _, value := range values {
			log.WithField("part", "header").Debugf("%s: %s\n", name, value)
		}
	}

	log.WithField("part", "body").Debug(body)

	return requestID
}

func (f *Filter) dumpResponse(requestID string, url string, header http.Header, body string) string {
	fileType := "filtered"

	if requestID == "" {
		rID, err := generateID()
		if err != nil {
			f.log.Fatalf("Failed to generate requestId: %v", err)
		}

		requestID = rID
		fileType = "original"
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
