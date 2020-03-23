package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

var (
	//Froms strings to be replaced
	Froms []string
	//Tos replacement string
	Tos []string
	//ContentTypes list of content types that will be filtered
	ContentTypes []string
	//Restricted list of net ranges allowed to connect to villip
	Restricted = []*net.IPNet{}
)

func isAuthorized(r *http.Response) (bool, error) {
	if len(Restricted) != 0 {
		sip, _, err := net.SplitHostPort(r.Request.RemoteAddr)
		if err != nil {
			log.WithFields(logrus.Fields{"userip": r.Request.RemoteAddr}).Error("userip is not IP:port")
			return true, err
		}

		ip := net.ParseIP(sip)
		if !ip.IsLoopback() {
			seen := false
			for _, ipnet := range Restricted {
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

func toFilter(log *logrus.Entry, r *http.Response) bool {
	if r.StatusCode == 200 {
		currentType := r.Header.Get("Content-Type")
		for _, testedType := range ContentTypes {
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

//UpdateResponse will be called back when the proxyfied server respond and filter the response if necessary
func UpdateResponse(r *http.Response) error {

	requestLog := log.WithFields(logrus.Fields{"url": r.Request.URL.String(), "status": r.StatusCode, "source": r.Request.RemoteAddr})
	// The Request in the Response is the last URL the client tried to access.
	requestLog.Debug("Request")

	authorized, err := isAuthorized(r)
	if err != nil || !authorized {
		return err
	}

	if !toFilter(requestLog, r) {
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
	for i := range Froms {
		s = strings.Replace(s, Froms[i], Tos[i], -1)
	}

	location := r.Header.Get("Location")
	if location != "" {
		origLocation := location
		for i := range Froms {
			location = strings.Replace(location, Froms[i], Tos[i], -1)
		}
		requestLog.WithFields(logrus.Fields{"location": origLocation, "rewrited_location": location}).Debug("will rewrite location header")
		r.Header.Set("Location", location)
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

func main() {
	var ok bool
	var from, to, restricteds string

	log.SetLevel(logrus.InfoLevel)
	if _, ok = os.LookupEnv("VILLIP_DEBUG"); ok {
		log.SetLevel(logrus.DebugLevel)
	}
	if from, ok = os.LookupEnv("VILLIP_FROM"); !ok {
		log.Fatal("Missing VILLIP_FROM environment variable")
	}
	if to, ok = os.LookupEnv("VILLIP_TO"); !ok {
		log.Fatal("Missing VILLIP_TO environment variable")
	}

	if restricteds, ok = os.LookupEnv("VILLIP_RESTRICTED"); ok {
		for _, ip := range strings.Split(strings.Replace(restricteds, " ", "", -1), ",") {
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				log.Fatal(fmt.Sprintf("\"%s\" in VILLIP_RESTRICTED environment variable is not a valid CIDR", ip))
			}
			Restricted = append(Restricted, ipnet)
		}
	}

	Froms = append(Froms, from)
	Tos = append(Tos, to)

	i := 1
	for {
		from, ok = os.LookupEnv(fmt.Sprintf("VILLIP_FROM_%d", i))
		if !ok {
			break
		}
		to, ok = os.LookupEnv(fmt.Sprintf("VILLIP_TO_%d", i))
		if !ok {
			log.Fatal(fmt.Sprintf("Missing VILLIP_TO_%d environment variable", i))
		}
		Froms = append(Froms, from)
		Tos = append(Tos, to)
		i++
	}

	villipURL, ok := os.LookupEnv("VILLIP_URL")
	if !ok {
		log.Fatal("Missing VILLIP_URL environment variable")
	}

	villipContenttypes, ok := os.LookupEnv("VILLIP_TYPES")
	if !ok {
		villipContenttypes = "text/html, text/css, application/javascript"
	}

	villipPort, ok := os.LookupEnv("VILLIP_PORT")
	if !ok {
		villipPort = "8080"
	}
	port, err := strconv.Atoi(villipPort)
	if err != nil || port > 65535 || 0 > port {
		log.Fatal(fmt.Sprintf("VILLIP_PORT environment variable (%s) is not a valid TCP port", villipPort))
	}

	ContentTypes = strings.Split(strings.Replace(villipContenttypes, " ", "", -1), ",")

	log.Info(fmt.Sprintf("Listen on port %d\n", port))
	log.Info(fmt.Sprintf("Will filter responses from %s\n", villipURL))
	if len(Restricted) != 0 {
		log.Info(fmt.Sprintf("Only for request from: %s \n", restricteds))
	}
	log.Info(fmt.Sprintf("For content-type %s", ContentTypes))
	log.Info("And replace:")
	for i := range Froms {
		log.Info(fmt.Sprintf("   %s  by  %s\n", Froms[i], Tos[i]))
	}

	u, _ := url.Parse(villipURL)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = UpdateResponse

	http.Handle("/", proxy)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.WithFields(logrus.Fields{"error": err}).Fatal("villip close on error")
	}
}
