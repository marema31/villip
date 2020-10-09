package filter

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

//NewFromEnv instantiate a Filter object from the environment variable configuration.
func NewFromEnv(upLog *logrus.Entry) *Filter {
	var ok bool

	var c config

	var from, to, restricteds string

	urls := []string{}
	villipPort, _ := os.LookupEnv("VILLIP_PORT")

	port, err := strconv.Atoi(villipPort)
	if err != nil {
		log.Fatalf("%s is not a valid TCP port", villipPort)
	}

	c.Port = port

	c.Force = false
	if _, ok := os.LookupEnv("VILLIP_FORCE"); ok {
		c.Force = true
	}

	if dumpFolder, ok := os.LookupEnv("VILLIP_DUMPFOLDER"); ok {
		c.Dump.Folder = dumpFolder
	}

	c.Response.Replace = []replacement{}

	if from, ok = os.LookupEnv("VILLIP_FROM"); ok {
		if to, ok = os.LookupEnv("VILLIP_TO"); !ok {
			log.Fatal("Missing VILLIP_TO environment variable")
		}

		if urlList, ok := os.LookupEnv("VILLIP_FOR"); ok {
			urls = strings.Split(strings.Replace(urlList, " ", "", -1), ",")
		}

		c.Response.Replace = append(c.Response.Replace, replacement{From: from, To: to, Urls: urls})
	}

	if restricteds, ok = os.LookupEnv("VILLIP_RESTRICTED"); ok {
		c.Restricted = strings.Split(strings.Replace(restricteds, " ", "", -1), ",")
	}

	i := 1

	for {
		from, ok = os.LookupEnv(fmt.Sprintf("VILLIP_FROM_%d", i))
		if !ok {
			break
		}

		to, ok = os.LookupEnv(fmt.Sprintf("VILLIP_TO_%d", i))
		if !ok {
			log.Fatalf("Missing VILLIP_TO_%d environment variable", i)
		}

		urls = []string{}
		if urlList, ok := os.LookupEnv(fmt.Sprintf("VILLIP_FOR_%d", i)); ok {
			urls = strings.Split(strings.Replace(urlList, " ", "", -1), ",")
		}

		c.Response.Replace = append(c.Response.Replace, replacement{From: from, To: to, Urls: urls})
	}

	url, ok := os.LookupEnv("VILLIP_URL")
	if !ok {
		log.Fatal("Missing VILLIP_URL environment variable")
	}

	c.URL = url

	if contenttypes, ok := os.LookupEnv("VILLIP_TYPES"); ok {
		c.ContentTypes = strings.Split(strings.Replace(contenttypes, " ", "", -1), ",")
	}

	if dumpURLs, ok := os.LookupEnv("VILLIP_DUMPURLS"); ok {
		c.Dump.URLs = strings.Split(strings.Replace(dumpURLs, " ", "", -1), ",")
	}

	return newFromConfig(upLog, c)
}
