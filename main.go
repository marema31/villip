package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
)

var Froms, Tos, ContentTypes []string
var Debug bool

func UpdateResponse(r *http.Response) error {

	if Debug == true {
		// The Request in the Response is the last URL the client tried to access.
		fmt.Printf("\n%s [%d]", r.Request.URL.String(), r.StatusCode)
	}

	if r.StatusCode == 200 {
		currentType := r.Header.Get("Content-Type")
		found := false

		for _, testedType := range ContentTypes {
			if strings.Contains(currentType, testedType) {
				found = true
				break
			}
		}
		if found == false {
			if Debug == true {
				fmt.Printf("... skipping type = %s", currentType)
			}
			return nil
		}
	} else if r.StatusCode != 302 && r.StatusCode != 301 {
		if Debug == true {
			fmt.Printf("... skipping status = %d", r.StatusCode)
		}
		return nil
	}

	if Debug == true {
		fmt.Print("...filtering")
	}

	var b []byte

	var body io.ReadCloser
	switch r.Header.Get("Content-Encoding") {
	case "gzip":
		body, _ = gzip.NewReader(r.Body)
		//		defer body.Close()
	default:
		body = r.Body
	}

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}

	s := string(b)
	for i, _ := range Froms {
		s = strings.Replace(s, Froms[i], Tos[i], -1)
	}

	location := r.Header.Get("Location")
	if location != "" {
		if Debug == true {
			fmt.Printf("... will rewrite location header = %s", location)
		}
		for i, _ := range Froms {
			location = strings.Replace(location, Froms[i], Tos[i], -1)
		}
		if Debug == true {
			fmt.Printf("=> %s\n", location)
		}
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
	var from, to string

	if from, ok = os.LookupEnv("VILLIP_DEBUG"); ok == true {
		Debug = true
	}
	if from, ok = os.LookupEnv("VILLIP_FROM"); ok == false {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_FROM environment variable")
		os.Exit(1)
	}
	if to, ok = os.LookupEnv("VILLIP_TO"); ok == false {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_TO environment variable")
		os.Exit(1)
	}
	Froms = append(Froms, from)
	Tos = append(Tos, to)

	i := 1
	for {
		from, ok = os.LookupEnv(fmt.Sprintf("VILLIP_FROM_%d", i))
		if ok == false {
			break
		}
		to, ok = os.LookupEnv(fmt.Sprintf("VILLIP_TO_%d", i))
		if ok == false {
			fmt.Fprintf(os.Stderr, "Missing VILLIP_TO_%d environment variable", i)
			os.Exit(1)
		}
		Froms = append(Froms, from)
		Tos = append(Tos, to)
		i++
	}

	villip_url, ok := os.LookupEnv("VILLIP_URL")
	if ok == false {
		fmt.Fprintf(os.Stderr, "Missing VILLIP_URL environment variable")
		os.Exit(1)
	}

	villip_contenttypes, ok := os.LookupEnv("VILLIP_TYPES")
	if ok == false {
		villip_contenttypes = "text/html, text/css, application/javascript"
	}

	ContentTypes = strings.Split(strings.Replace(villip_contenttypes, " ", "", -1), ",")

	fmt.Printf("Will filter responses fron %s\n", os.Getenv("VILLIP_URL"))
	fmt.Printf("For content-type %s\nAnd replace:\n", ContentTypes)
	for i, _ := range Froms {
		fmt.Printf("   %s  by  %s\n", Froms[i], Tos[i])
	}

	u, _ := url.Parse(villip_url)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.ModifyResponse = UpdateResponse

	http.Handle("/", proxy)

	http.ListenAndServe(":8080", nil)
}
