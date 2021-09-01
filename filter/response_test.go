package filter

import (
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestFilter_UpdateResponse(t *testing.T) {
	type fields struct {
		force    bool
		response response
		dumpURLS []*regexp.Regexp
	}
	type args struct {
		status int
		url    string
		body   string
		header http.Header
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantLog     []string
		wantBody    string
		wantHeaders http.Header
	}{
		{
			"empty body",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"",
				http.Header{
					"Content-Type": []string{"text/html"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : ",
				"Body after the replacement : ",
				"",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"",
			http.Header{
				"Content-Type":   []string{"text/html"},
				"Content-Length": []string{"0"},
				"X-ENV":          []string{"prod"},
			},
		},
		{
			"replace error",
			fields{
				false,
				response{},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"hello world",
				http.Header{
					"Content-Type":     []string{"text/html"},
					"Content-Encoding": []string{"gzip"},
				},
			},
			true,
			[]string{
				"filtering",
				"Body before the replacement : ",
				"Body after the replacement : ",
				"",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"",
			http.Header{
				"Content-Type":   []string{"text/html"},
				"Content-Length": []string{"0"},
				"X-ENV":          []string{"prod"},
			},
		},
		{
			"not filtered",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":        []string{"dev"},
					"X-Authors":    []string{"alice", "bob"},
					"Content-Type": []string{"text/xml"},
				},
			},
			false,
			[]string{
				"... skipping type",
			},
			"take your book,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":        []string{"dev"},
				"X-Authors":    []string{"alice", "bob"},
				"Content-Type": []string{"text/xml"},
			},
		},
		{
			"forced",
			fields{
				true,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":        []string{"dev"},
					"X-Authors":    []string{"alice", "bob"},
					"Content-Type": []string{"text/xml"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":          []string{"prod"},
				"X-Authors":      []string{"alice", "bob"},
				"Content-Length": []string{"46"},
				"Content-Type":   []string{"text/xml"},
			},
		},
		{
			"replace only body",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":        []string{"dev"},
					"X-Authors":    []string{"alice", "bob"},
					"Content-Type": []string{"text/html"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":          []string{"prod"},
				"X-Authors":      []string{"alice", "bob"},
				"Content-Length": []string{"46"},
				"Content-Type":   []string{"text/html"},
			},
		},
		{
			"dump",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{regexp.MustCompile("/youngster")},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":        []string{"dev"},
					"X-Authors":    []string{"alice", "bob"},
					"Content-Type": []string{"text/html"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"",
				"Content-Type: text/html",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"take your book,\ntry to dance\n sing often",
				"Content-Length: 46",
				"Content-Type: text/html",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"take your smartphone,\ntry to dance\n sing often",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod"},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":          []string{"prod"},
				"X-Authors":      []string{"alice", "bob"},
				"Content-Length": []string{"46"},
				"Content-Type":   []string{"text/html"},
			},
		},
		{
			"requestID",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{regexp.MustCompile("/youngster")},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":               []string{"dev"},
					"X-VILLIP-Request-ID": []string{"ABCDEFGHIJ"},
					"X-Authors":           []string{"alice", "bob"},
					"Content-Type":        []string{"text/html"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"",
				"Content-Type: text/html",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"X-VILLIP-Request-ID: ABCDEFGHIJ",
				"take your book,\ntry to dance\n sing often",
				"Content-Length: 46",
				"Content-Type: text/html",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"X-VILLIP-Request-ID: ABCDEFGHIJ",
				"take your smartphone,\ntry to dance\n sing often",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod"},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":               []string{"prod"},
				"X-Authors":           []string{"alice", "bob"},
				"Content-Length":      []string{"46"},
				"Content-Type":        []string{"text/html"},
				"X-VILLIP-Request-ID": []string{"ABCDEFGHIJ"},
			},
		},
		{
			"replace header",
			fields{
				false,
				response{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				[]*regexp.Regexp{},
			},
			args{
				http.StatusOK,
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":        []string{"dev"},
					"X-Authors":    []string{"alice", "bob"},
					"Content-Type": []string{"text/html"},
				},
			},
			false,
			[]string{
				"filtering",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":          []string{"prod"},
				"X-Authors":      []string{"alice", "bob"},
				"Content-Length": []string{"46"},
				"Content-Type":   []string{"text/html"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			req, _ := http.NewRequest("GET", tt.args.url, strings.NewReader(tt.args.body))
			if tt.args.body == "" {
				req.Body = nil
			}

			r := http.Response{
				Header:     tt.args.header,
				StatusCode: tt.args.status,
				Request:    req,
				Body:       ioutil.NopCloser(strings.NewReader(tt.args.body)),
			}

			f := &Filter{
				force:        tt.fields.force,
				dumpURLs:     tt.fields.dumpURLS,
				response:     tt.fields.response,
				contentTypes: []string{"text/html", "application/json"},
				log:          log,
			}
			err := f.UpdateResponse(&r)

			if (err != nil) != tt.wantErr {
				t.Errorf("Filter.UpdateResponse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil {
				return
			}

			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != tt.wantBody {
				t.Errorf("Filter.UpdateResponse() got = %#v, \nwant %#v", string(body), tt.wantBody)
			}
			if !reflect.DeepEqual(r.Header, tt.wantHeaders) {
				t.Errorf("Filter.UpdateResponse() got1 = %#v, want %#v", r.Header, tt.wantHeaders)
			}

			verifyLogged("Filter.UpdateResponse", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_toFilter(t *testing.T) {
	type args struct {
		contentTypes string
		status       int
		header       http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantLog []string
	}{
		{
			"Ok filtered",
			args{
				"text/html",
				http.StatusOK,
				http.Header{
					"Content-Type": []string{"text/html"},
				},
			},
			true,
			[]string{},
		},
		{
			"Ok not filtered",
			args{
				"text/html",
				http.StatusOK,
				http.Header{
					"Content-Type": []string{"text/xml"},
				},
			},
			false,
			[]string{"... skipping type"},
		},
		{
			"Found filtered",
			args{
				"text/html",
				http.StatusFound,
				http.Header{
					"Content-Type": []string{"text/html"},
				},
			},
			true,
			[]string{},
		},
		{
			"Moved Permanently filtered",
			args{
				"text/html",
				http.StatusMovedPermanently,
				http.Header{
					"Content-Type": []string{"text/html"},
				},
			},
			true,
			[]string{},
		},
		{
			"Default not filtered",
			args{
				"text/html",
				http.StatusAccepted,
				http.Header{
					"Content-Type": []string{"text/html"},
				},
			},
			false,
			[]string{"... skipping status"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			r := http.Response{
				Header:     tt.args.header,
				StatusCode: tt.args.status,
			}

			f := &Filter{
				contentTypes: []string{"text/html", "application/json"},
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
			}

			if got := f.toFilter(log, &r); got != tt.want {
				t.Errorf("Filter.toFilter() = %v, want %v", got, tt.want)
			}

			verifyLogged("Filter.toFilter", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_location(t *testing.T) {
	type fields struct {
		response response
	}
	type args struct {
		requestURL string
		header     http.Header
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantLog     []string
		wantHeaders http.Header
	}{
		{
			"minimal",
			fields{
				response{},
			},
			args{
				"http://localhost:8081/youngster/1",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			[]string{},
			http.Header{
				"X-ENV":     []string{"dev"},
				"X-Authors": []string{"alice", "bob"},
			},
		},
		{
			"unchange location",
			fields{
				response{},
			},
			args{
				"http://localhost:8081/youngster/1",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
					"Location":  []string{"http://www.example.com"},
				},
			},
			[]string{"will rewrite location header"},
			http.Header{
				"X-ENV":     []string{"dev"},
				"X-Authors": []string{"alice", "bob"},
				"Location":  []string{"http://www.example.com"},
			},
		},
		{
			"change location",
			fields{
				response{
					[]replaceParameters{
						{
							from: "example",
							to:   "test",
							urls: []*regexp.Regexp{},
						},
					},
					[]Cheader{},
				},
			},
			args{
				"http://localhost:8081/youngster/1",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
					"Location":  []string{"http://www.example.com"},
				},
			},
			[]string{"will rewrite location header"},
			http.Header{
				"X-ENV":     []string{"dev"},
				"X-Authors": []string{"alice", "bob"},
				"Location":  []string{"http://www.test.com"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{
				response: tt.fields.response,
			}

			r := http.Response{
				Header: tt.args.header,
			}

			f.location(log, &r, tt.args.requestURL)

			if !reflect.DeepEqual(r.Header, tt.wantHeaders) {
				t.Errorf("Filter.location() got1 = %#v, want %#v", r.Header, tt.wantHeaders)
			}

			verifyLogged("Filter.location", tt.wantLog, hook, t)
		})
	}
}
