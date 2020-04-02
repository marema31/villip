package filter

import (
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

//NewFromEnv instanciate a Filter object from the environment variable configuration
func NewFromEnv(upLog *logrus.Entry) *Filter {
	var ok bool
	var from, to, restricteds string

	f := Filter{}

	f.froms = []string{}
	f.tos = []string{}
	if from, ok = os.LookupEnv("VILLIP_FROM"); ok {
		if to, ok = os.LookupEnv("VILLIP_TO"); !ok {
			upLog.Fatal("Missing VILLIP_TO environment variable")
		}
		f.froms = append(f.froms, from)
		f.tos = append(f.tos, to)
	}

	f.restricted = []*net.IPNet{}
	if restricteds, ok = os.LookupEnv("VILLIP_RESTRICTED"); ok {
		for _, ip := range strings.Split(strings.Replace(restricteds, " ", "", -1), ",") {
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				upLog.Fatalf("\"%s\" in VILLIP_RESTRICTED environment variable is not a valid CIDR", ip)
			}
			f.restricted = append(f.restricted, ipnet)
		}
	}

	i := 1
	for {
		from, ok = os.LookupEnv(fmt.Sprintf("VILLIP_FROM_%d", i))
		if !ok {
			break
		}
		to, ok = os.LookupEnv(fmt.Sprintf("VILLIP_TO_%d", i))
		if !ok {
			upLog.Fatalf("Missing VILLIP_TO_%d environment variable", i)
		}
		f.froms = append(f.froms, from)
		f.tos = append(f.tos, to)
		i++
	}

	url, ok := os.LookupEnv("VILLIP_URL")
	if !ok {
		upLog.Fatal("Missing VILLIP_URL environment variable")
	}
	f.url = url

	contenttypes, ok := os.LookupEnv("VILLIP_TYPES")
	if !ok {
		contenttypes = "text/html, text/css, application/javascript"
	}
	f.contentTypes = strings.Split(strings.Replace(contenttypes, " ", "", -1), ",")

	villipPort, ok := os.LookupEnv("VILLIP_PORT")
	if !ok {
		villipPort = "8080"
	}

	port, err := strconv.Atoi(villipPort)
	if err != nil || port > 65535 || 0 > port {
		log.Fatal(fmt.Sprintf("VILLIP_PORT environment variable (%s) is not a valid TCP port", villipPort))
	}

	f.port = villipPort

	f.log = upLog.WithField("port", f.port)
	f.startLog()

	return &f
}
