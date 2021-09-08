package http_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/marema31/villip/filter"
	server "github.com/marema31/villip/server/http"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestServer_ConditionalProxy(t *testing.T) {
	type fields struct {
		filter  *filter.Mock
		filters []*filter.Mock
	}
	type args struct {
		remoteAddr string
	}
	type wants struct {
		reqBody   string
		reqHeader http.Header
		resBody   string
		resHeader http.Header
		status    int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		wants  wants
	}{
		{
			"wrong ip",
			fields{
				filter.NewMock(
					filter.HTTP,
					0,
					true,
					true,
					"take your book,\ntry to dance\n sing often",
					http.Header{},
					"walk outside,\n play boardgames",
					http.Header{},
					&testing.T{},
				),
				[]*filter.Mock{},
			},
			args{"192.168.1"},
			wants{
				"take your book,\ntry to dance\n sing often",
				http.Header{},
				"Unable to parse source IP\n",
				http.Header{
					"Content-Type":           []string{"text/plain; charset=utf-8"},
					"X-Content-Type-Options": []string{"nosniff"},
				},
				http.StatusInternalServerError,
			},
		},
		{
			"One filter",
			fields{
				filter.NewMock(
					filter.HTTP,
					0,
					true,
					true,
					"take your book,\ntry to dance\n sing often",
					http.Header{
						"Host":  []string{"example.com"},
						"X-Env": []string{"prod"},
					},
					"walk outside,\n play boardgames",
					http.Header{
						"Content-Length": []string{"30"},
					},
					&testing.T{},
				),
				[]*filter.Mock{},
			},
			args{"192.168.1.2:65432"},
			wants{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host":  []string{"example.com"},
					"X-Env": []string{"prod"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				http.StatusOK,
			},
		},
		{
			"Not Concerned",
			fields{
				filter.NewMock(
					filter.HTTP,
					0,
					false,
					false,
					"take your book,\ntry to dance\n sing often",
					http.Header{
						"Host":  []string{"example.com"},
						"X-Env": []string{"prod"},
					},
					"walk outside,\n play boardgames",
					http.Header{
						"Content-Length": []string{"30"},
					},
					&testing.T{},
				),
				[]*filter.Mock{},
			},
			args{"192.168.1.2:65432"},
			wants{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host":  []string{"example.com"},
					"X-Env": []string{"prod"},
				},
				"No filter correspond to this requests\n",
				http.Header{
					"Content-Type":           []string{"text/plain; charset=utf-8"},
					"X-Content-Type-Options": []string{"nosniff"},
				},
				http.StatusNotFound,
			},
		}, {
			"One filter",
			fields{
				filter.NewMock(
					filter.HTTP,
					0,
					true,
					true,
					"take your book,\ntry to dance\n sing often",
					http.Header{
						"Host":  []string{"example.com"},
						"X-Env": []string{"prod"},
					},
					"walk outside,\n play boardgames",
					http.Header{
						"Content-Length": []string{"30"},
					},
					&testing.T{},
				),
				[]*filter.Mock{},
			},
			args{"192.168.1.2:65432"},
			wants{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host":  []string{"example.com"},
					"X-Env": []string{"prod"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				http.StatusOK,
			},
		},
		{
			"list of filters",
			fields{
				filter.NewMock(
					filter.HTTP,
					0,
					false,
					true,
					"hello world",
					http.Header{
						"Host":  []string{"example.com"},
						"X-Env": []string{"prod"},
					},
					"world hello",
					http.Header{
						"Content-Length": []string{"30"},
					},
					&testing.T{},
				),
				[]*filter.Mock{
					filter.NewMock(
						filter.HTTP,
						0,
						false,
						true,
						"hello world 2",
						http.Header{
							"Host":  []string{"example.com"},
							"X-Env": []string{"prod"},
						},
						"world hello 2",
						http.Header{
							"Content-Length": []string{"30"},
						},
						&testing.T{},
					),
					filter.NewMock(
						filter.HTTP,
						0,
						true,
						true,
						"take your book,\ntry to dance\n sing often",
						http.Header{
							"Host":  []string{"example.com"},
							"X-Env": []string{"prod"},
						},
						"walk outside,\n play boardgames",
						http.Header{
							"Content-Length": []string{"30"},
						},
						&testing.T{},
					),
					filter.NewMock(
						filter.HTTP,
						0,
						true,
						true,
						"hello world 3",
						http.Header{
							"Host":  []string{"example.com"},
							"X-Env": []string{"prod"},
						},
						"world hello 3",
						http.Header{
							"Content-Length": []string{"30"},
						},
						&testing.T{},
					)},
			},
			args{"192.168.1.2:65432"},
			wants{
				"take your book,\ntry to dance\n sing often",
				http.Header{
					"Host":  []string{"example.com"},
					"X-Env": []string{"prod"},
				},
				"walk outside,\n play boardgames",
				http.Header{
					"Content-Length": []string{"30"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
				},
				http.StatusOK,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// logrus mock
			log, _ := logrustest.NewNullLogger()

			tt.fields.filter.T = t
			s := server.New(log, "64535", tt.fields.filter)

			for _, f := range tt.fields.filters {
				f := f // Warning of using variables in loop
				f.T = t
				s.Insert(f)
			}

			req, _ := http.NewRequest("GET", "/", strings.NewReader(tt.wants.reqBody))
			for name, value := range tt.wants.reqHeader {
				req.Header[name] = value
			}
			req.RemoteAddr = tt.args.remoteAddr

			res := httptest.NewRecorder()

			s.ConditionalProxy(res, req)

			//Tests
			if res.Result().StatusCode != tt.wants.status {
				t.Errorf("Wrong response status got = %d , want = %d", res.Result().StatusCode, tt.wants.status)
			}
			b, _ := ioutil.ReadAll(res.Body)
			if string(b) != tt.wants.resBody {
				t.Errorf("Response body: got = %s, want = %s", string(b), tt.wants.resBody)
			}

			h := res.Header()
			delete(h, "Date")
			if !reflect.DeepEqual(h, tt.wants.resHeader) {
				t.Errorf("Response header \ngot  = %#v\nwant = %#v", res.Header(), tt.wants.resHeader)
			}
		})
	}
}
