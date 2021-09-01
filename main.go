package main

import (
	"os"

	"golang.org/x/sync/errgroup"

	"github.com/sirupsen/logrus"

	"github.com/marema31/villip/filterlist"
)

func main() {
	log := logrus.New()
	filters := filterlist.New()

	log.SetLevel(logrus.InfoLevel)

	if _, ok := os.LookupEnv("VILLIP_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
	}

	upLog := log.WithField("app", "villip")
	filters.ReadConfig(upLog)

	servers := filters.CreateServers(upLog)
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
