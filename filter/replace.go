package filter

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/sirupsen/logrus"
)

func do(url string, s string, rep *[]replaceParameters) string {
	for _, r := range *rep {
		if len(r.urls) != 0 {
			found := false

			for _, reg := range r.urls {
				if reg.MatchString(url) {
					found = true
					break
				}
			}

			if !found {
				continue
			}
		}

		s = strings.Replace(s, r.from, r.to, -1)
	}

	return s
}

func (f *Filter) headerReplace(log *logrus.Entry, h *http.Header, a string) {
	log.Debug("Checking if need to replace header")

	var header []header

	if a == "request" {
		header = f.request.Header
	} else if a == "response" {
		header = f.response.Header
	}

	for _, head := range header {
		if (*h)[head.Name] == nil || (*h)[head.Name][0] == "" || head.Force {
			h.Set(head.Name, head.Value)
			log.Debug(fmt.Sprintf("Set header %s with value :  %s", head.Name, head.Value))
		}
	}
}
