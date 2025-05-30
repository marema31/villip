package filter

import (
	"fmt"
)

func (f *Filter) startLog() {
	if len(f.restricted) != 0 {
		f.log.Info(fmt.Sprintf("Only for request from: %s ", f.restricted))
	} else {
		f.log.Info("All requests")
	}

	if f.kind != HTTP {
		return
	}

	f.log.Info(fmt.Sprintf("For content-type %s", f.contentTypes))

	f.printBodyReplaceInLog("request")
	f.printHeaderReplaceInLog("request")
	f.printBodyReplaceInLog("response")
	f.printHeaderReplaceInLog("response")
}

func (f *Filter) printBodyReplaceInLog(action string) {
	rep := []replaceParameters{}

	switch action {
	case "request":
		rep = f.request.Replace
	case "response":
		rep = f.response.Replace
	}

	if len(rep) > 0 {
		f.log.Info(fmt.Sprintf("And replace in %s body:", action))

		for _, r := range rep {
			f.log.Info(fmt.Sprintf("   %s  by  %s", r.from, r.to))

			if len(r.urls) != 0 {
				var us []string

				for _, u := range r.urls {
					us = append(us, u.String())
				}

				f.log.Info(fmt.Sprintf("    for %v", us))
			}
		}
	}
}

func (f *Filter) printHeaderReplaceInLog(action string) {
	head := []Cheader{}

	switch action {
	case "request":
		head = f.request.Header
	case "response":
		head = f.response.Header
	}

	if len(head) > 0 {
		f.log.Info(fmt.Sprintf("And set/replace in %s Header:", action))

		for _, h := range head {
			m := fmt.Sprintf("    for header %s set/replace value by : %s", h.Name, h.Value)
			if h.Force {
				m += " (force = true -> in all the cases)"
			} else {
				m += " (force = false -> only if value is empty or header undefined)"
			}

			f.log.Info(m)
		}
	}
}
