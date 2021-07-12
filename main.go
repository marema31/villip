package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"golang.org/x/sync/errgroup"

	"github.com/sirupsen/logrus"

	"github.com/marema31/villip/filter"
)

func main() {
	log := logrus.New()
	filters := make(filtersList)
	var (
		f        *filter.Filter
		port     string
		priority uint8
	)

	log.SetLevel(logrus.InfoLevel)

	if _, ok := os.LookupEnv("VILLIP_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
	}

	upLog := log.WithField("app", "villip")

	if _, ok := os.LookupEnv("VILLIP_URL"); ok {
		port, priority, f = filter.NewFromEnv(upLog)
		insertInFilters(filters, port, priority, f)
	}

	if folderPath, ok := os.LookupEnv("VILLIP_FOLDER"); ok {
		files, err := ioutil.ReadDir(folderPath)
		if err != nil {
			log.Fatalf("Error getting list of configuration files: %v", err)
		}

		for _, file := range files {
			ext := filepath.Ext(file.Name())

			switch {
			case file.Mode().IsRegular() && (ext == ".yml" || ext == ".yaml"):
				port, priority, f = filter.NewFromYAML(upLog, filepath.Join(folderPath, file.Name()))
			case file.Mode().IsRegular() && (ext == ".json"):
				port, priority, f = filter.NewFromJSON(upLog, filepath.Join(folderPath, file.Name()))
			default:
				continue
			}
			insertInFilters(filters, port, priority, f)

		}
	}

	servers := createServers(filters, upLog)
	if len(servers) == 0 {
		log.Fatal("No filter configuration provided")
	}

	g := new(errgroup.Group)

	for _, s := range servers {
		g.Go(s.Serve)
	}

	if err := g.Wait(); err != nil {
		log.Fatalf("One server exiting in error: %v", err)
	}
}
