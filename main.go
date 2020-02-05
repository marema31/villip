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
)

var (
	//Froms strings to be replaced
	Froms []string
	//Tos replacement string
	Tos []string
	//ContentTypes list of content types that will be filtered
	ContentTypes []string
	//Restricted list of net ranges allowed to connect to villip
	Restricted = []*net.IPNet{}
	//Debug flag to log more informations to screen
	Debug bool
)

func printDebug(format string, args ...interface{}) {
	if Debug {
		fmt.Printf(format, args...)
	}
}

func isAuthorized(r *http.Response) (bool, error) {
	if len(Restricted) != 0 {
		sip, _, err := net.SplitHostPort(r.Request.RemoteAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "userip: %q is not IP:port", r.Request.RemoteAddr)
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
				printDebug("... forbidden from this IP \n")
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

func toFilter(r *http.Response) bool {
	if r.StatusCode == 200 {
		currentType := r.Header.Get("Content-Type")
		for _, testedType := range ContentTypes {
			if strings.Contains(currentType, testedType) {
				return true
			}
		}
		printDebug("... skipping type = %s", currentType)
		return false

	} else if r.StatusCode != 302 && r.StatusCode != 301 {
		printDebug("... skipping status = %d", r.StatusCode)
		return false
	}
	return true
}

//UpdateResponse will be called back when the proxyfied server respond and filter the response if necessary
func UpdateResponse(r *http.Response) error {

	// The Request in the Response is the last URL the client tried to access.
	printDebug("\n%s [%d] from %s", r.Request.URL.String(), r.StatusCode, r.Request.RemoteAddr)

	authorized, err := isAuthorized(r)
	if err != nil || !authorized {
		return err
	}

	if !toFilter(r) {
		return nil
	}
	printDebug("...filtering")

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
		printDebug("... will rewrite location header = %s", location)
		for i := range Froms {
			location = strings.Replace(location, Froms[i], Tos[i], -1)
		}
		printDebug("=> %s\n", location)
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

	if _, ok = os.LookupEnv("VILLIP_DEBUG"); ok {
		Debug = true
	}
	if from, ok = os.LookupEnv("VILLIP_FROM"); !ok {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_FROM environment variable")
		os.Exit(1)
	}
	if to, ok = os.LookupEnv("VILLIP_TO"); !ok {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_TO environment variable")
		os.Exit(1)
	}

	if restricteds, ok = os.LookupEnv("VILLIP_RESTRICTED"); ok {
		for _, ip := range strings.Split(strings.Replace(restricteds, " ", "", -1), ",") {
			_, ipnet, err := net.ParseCIDR(ip)
			if err != nil {
				fmt.Fprintf(os.Stderr, "\"%s\" in VILLIP_RESTRICTED environment variable is not a valid CIDR", ip)
				os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Missing VILLIP_TO_%d environment variable", i)
			os.Exit(1)
		}
		Froms = append(Froms, from)
		Tos = append(Tos, to)
		i++
	}

	villipURL, ok := os.LookupEnv("VILLIP_URL")
	if !ok {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_URL environment variable")
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "VILLIP_PORT environment variable (%s) is not a valid TCP port", villipPort)
		os.Exit(1)
	}

	ContentTypes = strings.Split(strings.Replace(villipContenttypes, " ", "", -1), ",")

	fmt.Printf("Listen on port %d\n", port)
	fmt.Printf("Will filter responses fron %s\n", os.Getenv("VILLIP_URL"))
	if len(Restricted) != 0 {
		fmt.Printf("Only for request from: %s \n", restricteds)
	}
	fmt.Printf("For content-type %s\nAnd replace:\n", ContentTypes)
	for i := range Froms {
		fmt.Printf("   %s  by  %s\n", Froms[i], Tos[i])
	}

	u, _ := url.Parse(villipURL)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = UpdateResponse

	http.Handle("/", proxy)

	err = http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "villip close on error %v", err)
		os.Exit(1)
	}
}
