package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var filters = []*filter.Filter{}

func main() {
	var f *filter.Filter

	log.SetLevel(logrus.InfoLevel)

	if _, ok := os.LookupEnv("VILLIP_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
	}

	upLog := log.WithField("app", "villip")

	if _, ok := os.LookupEnv("VILLIP_URL"); ok {
		f = filter.NewFromEnv(upLog)
		filters = append(filters, f)
	}

	if folderPath, ok := os.LookupEnv("VILLIP_FOLDER"); ok {
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			log.Fatalf("Error getting list of configuration files: %v", err)
		}

		for _, file := range files {
			ext := filepath.Ext(file.Name())
			if file.Mode().IsRegular() && (ext == ".yml" || ext == ".yaml") {
				f = filter.NewFromYAML(upLog, filepath.Join(folderPath, file.Name()))
				filters = append(filters, f)
			}

			if file.Mode().IsRegular() && (ext == ".json") {
				f = filter.NewFromJSON(upLog, filepath.Join(folderPath, file.Name()))
				filters = append(filters, f)
			}
		}
	}

	if len(filters) == 0 {
		log.Fatal("No filter configuration provided")
	}

	for _, f = range filters {
		go f.Serve()
	}

	for {
		time.Sleep(time.Hour * 24) //nolint: gomnd
	}
}
