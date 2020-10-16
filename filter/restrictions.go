package filter

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"
)

//IsConcerned determine if the request fulfil the filter condition.
func (f *Filter) IsConcerned(ip net.IP, parsedHeader http.Header) bool { //TODO: Add header parsing
	return f.isAuthorized(ip) && f.isAccepted(parsedHeader)
}

func (f *Filter) isAuthorized(ip net.IP) bool {
	if ip.IsLoopback() {
		return true
	}

	if len(f.restricted) != 0 {
		seen := false

		for _, ipnet := range f.restricted {
			if ipnet.Contains(ip) {
				seen = true
				break
			}
		}

		if !seen {
			f.log.WithFields(logrus.Fields{"source": ip}).Debug("filter forbidden for this IP")
			return false
		}
	}

	return true
}

func (f *Filter) isAccepted(parsedHeader http.Header) bool {
	for key, conditions := range f.token {
		value := parsedHeader.Get(key)
		if value == "" {
			f.log.WithFields(logrus.Fields{"header": key}).Debug("missing header for this filter")
			return false
		}

		f.log.WithFields(logrus.Fields{"header": key, "value": value}).Debug("lookup for condition")

		accepted := false
		rejected := false

		if len(conditions) == 1 && conditions[0].action == notEmpty {
			accepted = true
		}

		for _, cond := range conditions {
			switch cond.action {
			case notEmpty:
				accepted = true
			case accept:
				if value == cond.value {
					accepted = true
				}
			case reject:
				if value == cond.value {
					rejected = true
				}
			}
		}

		if rejected || !accepted {
			f.log.WithField("header", key).Debug("Refused")
			return false
		}
	}

	f.log.Debug("Accepted")

	return true
}

//IsConditional returns true if the filter has conditions.
func (f *Filter) IsConditional() bool {
	return len(f.restricted) != 0 || len(f.token) != 0
}
