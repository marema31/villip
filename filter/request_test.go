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

func TestFilter_UpdateRequest(t *testing.T) {
	type fields struct {
		request  request
		dumpURLs []*regexp.Regexp
	}
	type args struct {
		url    string
		body   string
		header http.Header
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantLog     []string
		wantBody    string
		wantHeaders http.Header
	}{
		{
			"minimal",
			fields{
				request{},
				[]*regexp.Regexp{},
			},
			args{
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			[]string{
				"Request received\nGET /youngster/1 HTTP/1.1\nHost: localhost:8080\nX-Authors: alice\nX-Authors: bob\nX-ENV: dev\n",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your book,\ntry to dance\n sing often",
			},
			"take your book,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":          []string{"dev"},
				"X-Authors":      []string{"alice", "bob"},
				"Content-Length": []string{"40"},
			},
		},
		{
			"requestId",
			fields{
				request{},
				[]*regexp.Regexp{regexp.MustCompile("/youngster")},
			},
			args{
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			[]string{
				"Request received\nGET /youngster/1 HTTP/1.1\nHost: localhost:8080\nX-Authors: alice\nX-Authors: bob\nX-ENV: dev\n",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your book,\ntry to dance\n sing often",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"take your book,\ntry to dance\n sing often",
				"Content-Length: 40",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"X-Villip-Request-Id: 123456789012",
				"take your book,\ntry to dance\n sing often",
			},
			"take your book,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":               []string{"dev"},
				"X-Authors":           []string{"alice", "bob"},
				"Content-Length":      []string{"40"},
				"X-Villip-Request-Id": []string{"123456789012"},
			},
		},
		{
			"replace content",
			fields{
				request{
					[]replaceParameters{
						{
							from: "book",
							to:   "smartphone",
						},
					},
					[]Cheader{},
				},
				[]*regexp.Regexp{regexp.MustCompile("/youngster")},
			},
			args{
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			[]string{
				"Request received\nGET /youngster/1 HTTP/1.1\nHost: localhost:8080\nX-Authors: alice\nX-Authors: bob\nX-ENV: dev\n",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"take your book,\ntry to dance\n sing often",
				"Content-Length: 46",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"X-Villip-Request-Id: 123456789012",
				"take your smartphone,\ntry to dance\n sing often",
			},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":               []string{"dev"},
				"X-Authors":           []string{"alice", "bob"},
				"Content-Length":      []string{"46"},
				"X-Villip-Request-Id": []string{"123456789012"},
			},
		},
		{
			"replace header",
			fields{
				request{
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
				"http://localhost:8081/youngster/1",
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			[]string{
				"Request received\nGET /youngster/1 HTTP/1.1\nHost: localhost:8080\nX-Authors: alice\nX-Authors: bob\nX-ENV: dev\n",
				"Body before the replacement : take your book,\ntry to dance\n sing often",
				"Body after the replacement : take your smartphone,\ntry to dance\n sing often",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"take your book,\ntry to dance\n sing often",
				"Content-Length: 46",
				"X-Authors: alice",
				"X-Authors: bob",
				"X-ENV: dev",
				"X-Villip-Request-Id: 123456789012",
				"take your smartphone,\ntry to dance\n sing often",
				"Checking if need to replace header",
				"Set header X-ENV with value :  prod",
			},
			"take your smartphone,\ntry to dance\n sing often",
			http.Header{
				"X-ENV":               []string{"prod"},
				"X-Authors":           []string{"alice", "bob"},
				"Content-Length":      []string{"46"},
				"X-Villip-Request-Id": []string{"123456789012"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			f := &Filter{
				request:  tt.fields.request,
				url:      "http://localhost:8080",
				log:      log,
				dumpURLs: tt.fields.dumpURLs,
			}

			r, _ := http.NewRequest("GET", tt.args.url, strings.NewReader(tt.args.body))
			r.Header = tt.args.header

			f.UpdateRequest(r)

			body, _ := ioutil.ReadAll(r.Body)
			if string(body) != tt.wantBody {
				t.Errorf("Filter.UpdateRequest() got = %v, want %v", string(body), tt.wantBody)
			}
			if !reflect.DeepEqual(r.Header, tt.wantHeaders) {
				t.Errorf("Filter.UpdateRequest() got1 = %#v, want %#v", r.Header, tt.wantHeaders)
			}

			verifyLogged("Filter.UpdateRequest", tt.wantLog, hook, t)
		})
	}
}
