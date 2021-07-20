package filter

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/sirupsen/logrus"
)

// Make newFromConfig mockable for unit test.
var _newFromConfig = newFromConfig //nolint: gochecknoglobals

func parseReplaceConfig(log logrus.FieldLogger, rep []Creplacement) []replaceParameters {
	result := make([]replaceParameters, 0)

	for _, r := range rep {
		p := replaceParameters{from: r.From, to: r.To, urls: []*regexp.Regexp{}}

		for _, reg := range r.Urls {
			r, err := regexp.Compile(reg)
			if err != nil {
				log.Fatalf("Failed to compile '%s' regular expression: %v", reg, err)
			}

			p.urls = append(p.urls, r)
		}

		result = append(result, p)
	}

	return result
}

func parseTokenConfig(log logrus.FieldLogger, tokenConfig CtokenAction) (string, headerConditions) {
	var hc headerConditions

	if len(tokenConfig.Header) == 0 {
		log.Fatal("token header parameter cannot be empty")
	}

	hc.value = tokenConfig.Value
	hc.action = accept

	action := strings.ToLower(tokenConfig.Action)
	switch action {
	case "accept":
		hc.action = accept
	case "reject":
		hc.action = reject
	case "notempty":
		hc.action = notEmpty
	default:
		log.Fatalf("'%s' is not a valid action for token condition", action)
	}

	return tokenConfig.Header, hc
}

//nolint: funlen,gocognit
func newFromConfig(log logrus.FieldLogger, c Config) (string, uint8, *Filter) {
	f := Filter{}

	if c.URL == "" {
		log.Fatal("Missing url variable")
	}

	f.url = c.URL
	f.priority = fmt.Sprintf("%d", c.Priority)

	if c.Port == 0 {
		c.Port = 8080
	}

	if c.Port > 65535 || 0 > c.Port {
		log.Fatalf("%d is not a valid TCP port", c.Port)
	}

	switch strings.ToLower(c.Type) {
	case "http":
		f.kind = httpFilter
	case "tcp":
		f.kind = httpFilter
	case "udp":
		f.kind = httpFilter
	default:
		f.kind = httpFilter
	}

	f.port = fmt.Sprintf("%d", c.Port)

	f.log = log.WithFields(logrus.Fields{"port": f.port, "url": f.url, "priority": f.priority})

	if c.Dump.Folder != "" {
		f.dumpFolder = c.Dump.Folder
		if _, err := os.Stat(f.dumpFolder); !os.IsNotExist(err) {
			err = os.MkdirAll(f.dumpFolder, os.ModePerm)
			if err != nil {
				f.log.Fatalf("Failed to create the dump folder %s: %v", f.dumpFolder, err)
			}
		}
	}

	f.dumpURLs = make([]*regexp.Regexp, 0)

	for _, reg := range c.Dump.URLs {
		r, err := regexp.Compile(reg)
		if err != nil {
			f.log.Fatalf("Failed to compile '%s' regular expression: %v", reg, err)
		}

		f.dumpURLs = append(f.dumpURLs, r)
	}

	f.force = c.Force
	f.insecure = c.Insecure

	responseReplace := make([]Creplacement, 0)

	switch {
	case len(c.Response.Replace) > 0 && len(c.Replace) > 0:
		f.log.Fatalf("Please check your config file you cannot set a response and a replace at the same time")
	case len(c.Replace) > 0:
		responseReplace = c.Replace
	case len(c.Response.Replace) > 0:
		responseReplace = c.Response.Replace
	}

	f.response.Replace = make([]replaceParameters, 0)
	if len(responseReplace) > 0 {
		f.response.Replace = parseReplaceConfig(f.log, responseReplace)
	}

	f.request.Replace = make([]replaceParameters, 0)
	if len(c.Request.Replace) > 0 {
		f.request.Replace = parseReplaceConfig(f.log, c.Request.Replace)
	}

	f.request.Header = make([]Cheader, 0)
	if len(c.Request.Header) > 0 {
		f.request.Header = c.Request.Header
	}

	f.response.Header = make([]Cheader, 0)
	if len(c.Response.Header) > 0 {
		f.response.Header = c.Response.Header
	}

	f.restricted = []*net.IPNet{}

	f.token = make(map[string][]headerConditions)

	for _, tokenConfig := range c.Token {
		header, token := parseTokenConfig(f.log, tokenConfig)
		if _, ok := f.token[header]; !ok {
			f.token[header] = make([]headerConditions, 0)
		}

		f.token[header] = append(f.token[header], token)
	}

	for _, ip := range c.Restricted {
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			f.log.Fatal(fmt.Sprintf("\"%s\" in restricted parameter is not a valid CIDR", ip))
		}

		f.restricted = append(f.restricted, ipnet)
	}

	f.contentTypes = append(f.contentTypes, c.ContentTypes...)

	if len(f.contentTypes) == 0 {
		f.contentTypes = append(f.contentTypes, []string{"text/html", "text/css", "application/javascript"}...)
	}

	f.startLog()

	return f.port, c.Priority, &f
}
