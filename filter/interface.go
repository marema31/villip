package filter

import (
	"net"
	"net/http"
)

// FilteredServer represents a reverse proxy.
type FilteredServer interface {
	IsConcerned(net.IP, http.Header) bool
	Serve(http.ResponseWriter, *http.Request)
	ServeTCP() error
	IsConditional() bool
	PrefixReplace(string) string
	Kind() Type
}
