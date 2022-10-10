package filter

import (
	"fmt"
	"strconv"
	"strings"
)

// NewFromEnv instantiate a Filter object from the environment variable configuration.
// nolint: funlen,gocognit
func (f *Factory) NewFromEnv() (string, uint8, FilteredServer) {
	var ok bool

	var c Config

	var from, to, restricteds string

	urls := []string{}

	url, ok := f.lookupEnv("VILLIP_URL")
	if !ok {
		f.log.Fatal("Missing VILLIP_URL environment variable")
	}

	c.URL = url

	if villipPriority, ok := f.lookupEnv("VILLIP_PRIORITY"); ok {
		priority, err := strconv.Atoi(villipPriority)
		if err != nil {
			f.log.Fatalf("%s is not a valid priority", villipPriority)
		}

		if priority < 0 || priority > 255 {
			f.log.Fatalf("%s is not a valid priority", villipPriority)
		}

		c.Priority = uint8(priority)
	}

	villipPort, _ := f.lookupEnv("VILLIP_PORT")

	if villipPort == "" {
		villipPort = "8080"
	}

	port, err := strconv.Atoi(villipPort)
	if err != nil {
		f.log.Fatalf("%s is not a valid TCP port", villipPort)
	}

	c.Port = port

	c.Force = false
	if _, ok := f.lookupEnv("VILLIP_FORCE"); ok {
		c.Force = true
	}

	if _, ok := f.lookupEnv("VILLIP_INSECURE"); ok {
		c.Insecure = true
	}

	if dumpFolder, ok := f.lookupEnv("VILLIP_DUMPFOLDER"); ok {
		c.Dump.Folder = dumpFolder
	}

	c.Replace = make([]Creplacement, 0)
	c.Request.Header = make([]Cheader, 0)
	c.Response.Header = make([]Cheader, 0)
	c.Request.Replace = make([]Creplacement, 0)
	c.Response.Replace = make([]Creplacement, 0)
	c.Prefix = make([]Creplacement, 0)

	if from, ok = f.lookupEnv("VILLIP_FROM"); ok {
		if to, ok = f.lookupEnv("VILLIP_TO"); !ok {
			f.log.Fatal("Missing VILLIP_TO environment variable")
		}

		if urlList, ok := f.lookupEnv("VILLIP_FOR"); ok {
			urls = strings.Split(strings.Replace(urlList, " ", "", -1), ",")
		}

		c.Response.Replace = append(c.Response.Replace, Creplacement{From: from, To: to, Urls: urls})
	}

	if restricteds, ok = f.lookupEnv("VILLIP_RESTRICTED"); ok {
		c.Restricted = strings.Split(strings.Replace(restricteds, " ", "", -1), ",")
	}

	i := 1

	for {
		from, ok = f.lookupEnv(fmt.Sprintf("VILLIP_FROM_%d", i))
		if !ok {
			break
		}

		to, ok = f.lookupEnv(fmt.Sprintf("VILLIP_TO_%d", i))
		if !ok {
			f.log.Fatalf("Missing VILLIP_TO_%d environment variable", i)
		}

		urls = []string{}
		if urlList, ok := f.lookupEnv(fmt.Sprintf("VILLIP_FOR_%d", i)); ok {
			urls = strings.Split(strings.Replace(urlList, " ", "", -1), ",")
		}

		c.Response.Replace = append(c.Response.Replace, Creplacement{From: from, To: to, Urls: urls})
		i++
	}

	if status, ok := f.lookupEnv("VILLIP_STATUS"); ok {
		c.Status = strings.Split(strings.Replace(status, " ", "", -1), ",")
	}

	if contenttypes, ok := f.lookupEnv("VILLIP_TYPES"); ok {
		c.ContentTypes = strings.Split(strings.Replace(contenttypes, " ", "", -1), ",")
	}

	if dumpURLs, ok := f.lookupEnv("VILLIP_DUMPURLS"); ok {
		c.Dump.URLs = strings.Split(strings.Replace(dumpURLs, " ", "", -1), ",")
	}

	from, ok = f.lookupEnv("VILLIP_PREFIX_FROM")
	if ok {
		to, ok = f.lookupEnv("VILLIP_PREFIX_TO")
		if !ok {
			f.log.Fatalf("Missing VILLIP_PREFIX_TO environment variable", i)
		}

		c.Prefix = []Creplacement{{From: from, To: to, Urls: []string{}}}
	}

	return f.newFromConfig(f.log, c)
}
