package filter

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"regexp"
	"strings"
	"testing"

	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestFilter_Serve(t *testing.T) {
	type fields struct {
		insecure   bool
		request    request
		response   response
		dumpFolder string
		dumpURLs   []*regexp.Regexp
	}
	type args struct {
		reqBody   string
		reqHeader http.Header
		resBody   string
		resHeader http.Header
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  args
	}{
		{
			"no filtering",
			fields{
				false,
				request{},
				response{},
				"",
				[]*regexp.Regexp{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host": []string{"example.com"},
				},
				"walk outside,\n play boardgames",
				http.Header{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Accept-Encoding":  []string{"gzip"},
					"Content-Length":   []string{"40"},
					"X-Forwarded-Host": []string{"example.com"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
			},
		},
		{
			"insecure",
			fields{
				true,
				request{},
				response{},
				"",
				[]*regexp.Regexp{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host": []string{"example.com"},
				},
				"walk outside,\n play boardgames",
				http.Header{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Accept-Encoding":  []string{"gzip"},
					"Content-Length":   []string{"40"},
					"X-Forwarded-Host": []string{"example.com"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
			},
		},
		{
			"only request",
			fields{
				true,
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
				response{},
				"",
				[]*regexp.Regexp{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host":  []string{"example.com"},
					"X-ENV": []string{"dev"},
				},
				"walk outside,\n play boardgames",
				http.Header{},
			},
			args{
				"take your smartphone,\ntry to dance\n sing often",
				http.Header{
					"Accept-Encoding":  []string{"gzip"},
					"Content-Length":   []string{"46"},
					"X-Forwarded-Host": []string{"example.com"},
					"X-Env":            []string{"prod"},
					"User-Agent":       []string{"Go-http-client/1.1"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
			},
		},
		{
			"only response",
			fields{
				true,
				request{},
				response{
					[]replaceParameters{
						{
							from: "boardgames",
							to:   "videogames",
						},
					},
					[]Cheader{
						{
							Name:  "X-ENV",
							Value: "prod",
						},
					},
				},
				"",
				[]*regexp.Regexp{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host": []string{"example.com"},
				},
				"walk outside,\n play boardgames",
				http.Header{},
			},
			args{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Accept-Encoding":  []string{"gzip"},
					"Content-Length":   []string{"40"},
					"X-Forwarded-Host": []string{"example.com"},
				},
				"walk outside,\n play videogames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
					"X-Env":          []string{"prod"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Mock the response
				w.Write([]byte(tt.wants.resBody))
				h := w.Header()
				for name, value := range tt.args.resHeader {
					h[name] = value
				}

				//Tests
				b, _ := ioutil.ReadAll(r.Body)
				if string(b) != tt.wants.reqBody {
					t.Errorf("Request body: got = %s, want %s", string(b), tt.wants.reqBody)
				}
				if !reflect.DeepEqual(r.Header, tt.wants.reqHeader) {
					t.Errorf("Request header \ngot  = %#v\nwant = %#v", r.Header, tt.wants.reqHeader)
				}

			}))
			defer backend.Close()
			backendURL, _ := url.Parse(backend.URL)

			// logrus mock
			log, _ := logrustest.NewNullLogger()

			f := &Filter{
				contentTypes: []string{"text/plain"},
				insecure:     tt.fields.insecure,
				response:     tt.fields.response,
				request:      tt.fields.request,
				url:          backendURL.String(),
				log:          log,
				dumpFolder:   tt.fields.dumpFolder,
				dumpURLs:     tt.fields.dumpURLs,
			}

			req, _ := http.NewRequest("GET", "/", strings.NewReader(tt.args.reqBody))
			for name, value := range tt.args.reqHeader {
				req.Header[name] = value
			}

			res := httptest.NewRecorder()

			f.Serve(res, req)

			//Tests
			b, _ := ioutil.ReadAll(res.Body)
			if string(b) != tt.wants.resBody {
				t.Errorf("Response body: got = %s, want %s", string(b), tt.wants.resBody)
			}

			h := res.Header()
			delete(h, "Date")
			if !reflect.DeepEqual(h, tt.wants.resHeader) {
				t.Errorf("Response header \ngot  = %#v\nwant = %#v", res.Header(), tt.wants.resHeader)
			}
		})
	}
}
