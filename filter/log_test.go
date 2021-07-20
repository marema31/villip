package filter

import (
	"net"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestFilter_startLog(t *testing.T) {
	type fields struct {
		request      request
		response     response
		contentTypes []string
		restricted   []*net.IPNet
	}
	tests := []struct {
		name    string
		fields  fields
		wantLog []string
	}{
		{
			"minimal response",
			fields{
				request{},
				response{},
				[]string{"text/html", "text/css", "application/javascript"},
				[]*net.IPNet{},
			},
			[]string{"All requests", "For content-type [text/html text/css application/javascript]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{
				log:          log,
				response:     tt.fields.response,
				request:      tt.fields.request,
				contentTypes: tt.fields.contentTypes,
				restricted:   tt.fields.restricted,
			}
			f.startLog()

			verifyLogged("Filter.startLog", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_printBodyReplaceInLog(t *testing.T) {
	type fields struct {
		request      request
		response     response
		contentTypes []string
	}
	type args struct {
		action string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLog []string
	}{
		{
			"minimal response",
			fields{
				request{},
				response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"response"},
			[]string{},
		},
		{
			"minimal request",
			fields{
				request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				response{},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"request"},
			[]string{},
		},
		{
			"response",
			fields{
				request{},
				response{
					Replace: []replaceParameters{
						{
							from: "videogame",
							to:   "boardgame",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/boomer"),
								regexp.MustCompile("/grandparent"),
							},
						},
					},
					Header: []Cheader{},
				},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"response"},
			[]string{"And replace in response body:", "   videogame  by  boardgame", "    for [/boomer /grandparent]"},
		},
		{
			"request",
			fields{
				request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				response{},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"request"},
			[]string{"And replace in request body:", "   book  by  smartphone", "    for [/youngster /children]"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{
				log:      log,
				response: tt.fields.response,
				request:  tt.fields.request,
			}
			f.printBodyReplaceInLog(tt.args.action)

			verifyLogged("Filter.printBodyReplaceInLog", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_printHeaderReplaceInLog(t *testing.T) {
	type fields struct {
		request      request
		response     response
		contentTypes []string
	}
	type args struct {
		action string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantLog []string
	}{
		{
			"minimal response",
			fields{
				request{},
				response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"response"},
			[]string{},
		},
		{
			"minimal request",
			fields{
				request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				response{},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"request"},
			[]string{},
		},
		{
			"response",
			fields{
				request{},
				response{
					Replace: []replaceParameters{},
					Header: []Cheader{
						{
							Name:  "X-TEST",
							Value: "valid",
							Force: false,
						},
						{
							Name:  "X-Author",
							Value: "bob",
							Force: true,
						},
					},
				},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"response"},
			[]string{
				"And set/replace in response Header:",
				"    for header X-TEST set/replace value by : valid (force = false -> only if value is empty or header undefined)",
				"    for header X-Author set/replace value by : bob (force = true -> in all the cases)",
			},
		},
		{
			"request",
			fields{
				request{
					Replace: []replaceParameters{},
					Header: []Cheader{
						{
							Name:  "X-ENV",
							Value: "dev",
							Force: false,
						},
						{
							Name:  "X-Version",
							Value: "1.2",
							Force: true,
						},
					},
				},
				response{},
				[]string{"text/html", "text/css", "application/javascript"},
			},
			args{"request"},
			[]string{
				"And set/replace in request Header:",
				"    for header X-ENV set/replace value by : dev (force = false -> only if value is empty or header undefined)",
				"    for header X-Version set/replace value by : 1.2 (force = true -> in all the cases)",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			f := &Filter{
				log:          log,
				response:     tt.fields.response,
				request:      tt.fields.request,
				contentTypes: tt.fields.contentTypes,
			}

			f.printHeaderReplaceInLog(tt.args.action)

			verifyLogged("Filter.printHeaderReplaceInLog", tt.wantLog, hook, t)
		})
	}
}
