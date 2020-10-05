package filter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"regexp"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type replacement struct {
	From string   `yaml:"from" json:"from"`
	To   string   `yaml:"to" json:"to"`
	Urls []string `yaml:"urls" json:"urls"`
}

type dump struct {
	Folder string   `yaml:"folder" json:"folder"`
	URLs   []string `yaml:"urls" json:"urls"`
}

type header struct {
	Name  string  `yaml:"name" json:"name"`
	Value string  `yaml:"value" json:"value"`
	Force bool	  `yaml:"force" json:"force"`
}

type action struct {
	Replace  []replacement `yaml:"replace" json:"replace"`
	Header	 []header	   `yaml:"header" json:"header"`
}

type config struct {
	ContentTypes []string      `yaml:"content-types" json:"content-types"`
	Dump         dump          `yaml:"dump" json:"dump"`
	Force        bool          `yaml:"force" json:"force"`
	Port         int           `yaml:"port" json:"port"`
	Replace      []replacement `yaml:"replace" json:"replace"`
	Request	     action	   	    `yaml:"request" json:"request"`
	Response	 action   	   `yaml:"response" json:"response"`
	Restricted   []string      `yaml:"restricted" json:"restricted"`
	URL          string        `yaml:"url" json:"url"`
}



//NewFromYAML instantiate a Filter object from the configuration file.
func NewFromYAML(upLog *logrus.Entry, filePath string) *Filter {
	log := upLog.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c config

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode YAML: %v", err)
	}
	return newFromConfig(upLog, c)
}

//NewFromJSON instantiate a Filter object from the configuration file.
func NewFromJSON(upLog *logrus.Entry, filePath string) *Filter {
	log := upLog.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c config

	err = json.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode JSON: %v", err)
	}

	return newFromConfig(upLog, c)
}

func replaceToReplacement(log *logrus.Entry, rep []replacement) []replaceParameters {
	var result = []replaceParameters{}
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

func newFromConfig(log *logrus.Entry, c config) *Filter {
	f := Filter{}

	if c.URL == "" {
		log.Fatal("Missing url variable")
	}
	f.url = c.URL

	if c.Port == 0 {
		c.Port = 8080
	}

	if c.Port > 65535 || 0 > c.Port {
		log.Fatalf("%d is not a valid TCP port", c.Port)
	}

	f.port = fmt.Sprintf("%d", c.Port)

	f.log = log.WithField("port", f.port)

	if c.Dump.Folder != "" {
		f.dumpFolder = c.Dump.Folder
		if _, err := os.Stat(f.dumpFolder); !os.IsNotExist(err) {
			err = os.MkdirAll(f.dumpFolder, os.ModePerm)
			if err != nil {
				f.log.Fatalf("Failed to create the dump folder %s: %v", f.dumpFolder, err)
			}
		}
	}

	f.dumpURLs = []*regexp.Regexp{}

	for _, reg := range c.Dump.URLs {
		r, err := regexp.Compile(reg)
		if err != nil {
			f.log.Fatalf("Failed to compile '%s' regular expression: %v", reg, err)
		}

		f.dumpURLs = append(f.dumpURLs, r)
	}

	f.force = c.Force

	var responseReplace = []replacement{}
	if len(c.Response.Replace) > 0 && len(c.Replace) > 0 {
		f.log.Fatalf("Please check your config file you cannot set a reponse and a replace at the same time")
	} else if (len(c.Replace)) > 0 {
		responseReplace = c.Replace
	} else if (len(c.Response.Replace)) > 0 {
		responseReplace = c.Response.Replace
	}
	if len(responseReplace) > 0 {
		f.response.Replace = replaceToReplacement(f.log, responseReplace)
	}
	
	f.restricted = []*net.IPNet{}

	for _, ip := range c.Restricted {
		_, ipnet, err := net.ParseCIDR(ip)
		if err != nil {
			log.Fatal(fmt.Sprintf("\"%s\" in restricted parameter is not a valid CIDR", ip))
		}

		f.restricted = append(f.restricted, ipnet)
	}

	f.contentTypes = append(f.contentTypes, c.ContentTypes...)

	if len(f.contentTypes) == 0 {
		f.contentTypes = append(f.contentTypes, []string{"text/html", "text/css", "application/javascript"}...)
	}

	f.startLog()

	return &f
}
