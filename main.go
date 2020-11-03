package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/marema31/villip/filter"
	"github.com/marema31/villip/server"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()
var servers = make(map[string]*server.Server)

func main() {
	var (
		f    *filter.Filter
		port string
	)

	log.SetLevel(logrus.InfoLevel)

	if _, ok := os.LookupEnv("VILLIP_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
	}

	upLog := log.WithField("app", "villip")

	if _, ok := os.LookupEnv("VILLIP_URL"); ok {
		port, f = filter.NewFromEnv(upLog)
		servers[port] = server.New(upLog, port, f)
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
				port, f = filter.NewFromYAML(upLog, filepath.Join(folderPath, file.Name()))
			case file.Mode().IsRegular() && (ext == ".json"):
				port, f = filter.NewFromJSON(upLog, filepath.Join(folderPath, file.Name()))
			default:
				continue
			}

			if _, ok := servers[port]; ok {
				servers[port].Insert(f)
			} else {
				servers[port] = server.New(upLog, port, f)
			}
		}
	}

	if len(servers) == 0 {
		log.Fatal("No filter configuration provided")
	}

	for _, s := range servers {
		go s.Serve()
	}

	for {
		time.Sleep(time.Hour * 24) //nolint: gomnd
	}
}
