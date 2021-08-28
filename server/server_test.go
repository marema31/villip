package server_test

import (
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/marema31/villip/server"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

type mockFilter struct {
	concerned bool
	reqBody   string
	reqHeader http.Header
	resBody   string
	resHeader http.Header
	t         *testing.T
}

func (m *mockFilter) IsConcerned(ip net.IP, h http.Header) bool {
	return m.concerned
}

func (m *mockFilter) IsConditional() bool {
	return true
}

func (m *mockFilter) Serve(res http.ResponseWriter, req *http.Request) {
	res.Write([]byte(m.resBody))
	h := res.Header()
	for name, value := range m.resHeader {
		h[name] = value
	}

	//Tests
	b, _ := ioutil.ReadAll(req.Body)
	if string(b) != m.reqBody {
		m.t.Errorf("Request body: got = %s, want %s", string(b), m.reqBody)
	}
	if !reflect.DeepEqual(req.Header, m.reqHeader) {
		m.t.Errorf("Request header \ngot  = %#v\nwant = %#v", req.Header, m.reqHeader)
	}
}

func TestServer_ConditionalProxy(t *testing.T) {
	type fields struct {
		filter  mockFilter
		filters []mockFilter
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
				mockFilter{
					true,
					"take your book,\ntry to dance\n sing often",
					http.Header{},
					"walk outside,\n play boardgames",
					http.Header{},
					&testing.T{},
				},
				[]mockFilter{},
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
				mockFilter{
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
				},
				[]mockFilter{},
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
				mockFilter{
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
				},
				[]mockFilter{},
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
				mockFilter{
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
				},
				[]mockFilter{},
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
				mockFilter{
					false,
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
				},
				[]mockFilter{
					{
						false,
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
					},
					{
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
					},
					{
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
					}},
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

			tt.fields.filter.t = t
			s := server.New(log, "64535", &tt.fields.filter)

			for _, f := range tt.fields.filters {
				f := f // Warning of using variables in loop
				f.t = t
				s.Insert(&f)
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
