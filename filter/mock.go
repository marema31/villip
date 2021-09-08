package filter

import (
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"
)

// Mock provides a way to mimic a Filter object.
type Mock struct {
	Position    int
	Concerned   bool
	Conditional bool
	reqBody     string
	reqHeader   http.Header
	resBody     string
	resHeader   http.Header
	kind        Type
	T           *testing.T
}

// NewMock returns a new mock.
func NewMock(
	kind Type,
	position int,
	concerned bool,
	conditional bool,
	reqBody string,
	reqHeader http.Header,
	resBody string,
	resHeader http.Header,
	t *testing.T,
) *Mock {
	return &Mock{
		kind:        kind,
		Position:    position,
		Concerned:   concerned,
		Conditional: conditional,
		reqHeader:   reqHeader,
		reqBody:     reqBody,
		resHeader:   resHeader,
		resBody:     resBody,
		T:           t,
	}
}

// IsConcerned mimics the IsConcerned from Filter.
func (m *Mock) IsConcerned(ip net.IP, h http.Header) bool {
	return m.Concerned
}

// IsConditional mimics the IsConditional from Filter.
func (m *Mock) IsConditional() bool {
	return m.Conditional
}

// Kind mimics the Kind from Filter.
func (m *Mock) Kind() Type {
	return m.kind
}

// Serve mimics the Serve from Filter.
func (m *Mock) Serve(res http.ResponseWriter, req *http.Request) {
	_, _ = res.Write([]byte(m.resBody))
	h := res.Header()

	for name, value := range m.resHeader {
		h[name] = value
	}

	// Tests
	b, _ := ioutil.ReadAll(req.Body)
	if string(b) != m.reqBody {
		m.T.Errorf("Request body: got = %s, want %s", string(b), m.reqBody)
	}

	if !reflect.DeepEqual(req.Header, m.reqHeader) {
		m.T.Errorf("Request header \ngot  = %#v\nwant = %#v", req.Header, m.reqHeader)
	}
}

// PrefixReplace mimics the PrefixReplace from Filter.
func (m *Mock) PrefixReplace(URL string) string {
	return URL
}

// ServeTCP mimics the ServeTCP from Filter.
func (m *Mock) ServeTCP() error {
	return nil
}
