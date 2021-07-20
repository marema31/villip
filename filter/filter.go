package filter

import (
	"net"
	"regexp"

	"github.com/sirupsen/logrus"
)

type replaceParameters struct {
	from string
	to   string
	urls []*regexp.Regexp
}

type headerAction int

const (
	accept   headerAction = iota
	reject   headerAction = iota
	notEmpty headerAction = iota
)

type filterType int

const (
	httpFilter filterType = iota

//	tcpFilter
//	udpFilter
)

type headerConditions struct {
	value  string
	action headerAction
}

type response struct {
	Replace []replaceParameters `yaml:"replace" json:"replace"`
	Header  []Cheader           `yaml:"header" json:"header"`
}

type request struct {
	Replace []replaceParameters `yaml:"replace" json:"replace"`
	Header  []Cheader           `yaml:"header" json:"header"`
}

// Filter proxifies an URL and filter the response.
type Filter struct {
	insecure     bool
	force        bool
	response     response
	request      request
	contentTypes []string
	restricted   []*net.IPNet
	token        map[string][]headerConditions
	url          string
	port         string
	priority     string
	log          logrus.FieldLogger // Interface for Logger and Entry
	dumpFolder   string
	dumpURLs     []*regexp.Regexp
	kind         filterType
}
