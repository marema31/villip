package filter

import (
	"net"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestFilter_IsConcerned(t *testing.T) {
	type fields struct {
		restricted []*net.IPNet
		token      map[string][]headerConditions
	}
	type args struct {
		ip           net.IP
		parsedHeader http.Header
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			"accepted",
			fields{},
			args{
				[]byte{127, 0, 0, 1},
				http.Header{},
			},
			true,
		},
		{
			"token refused",
			fields{
				[]*net.IPNet{
					{IP: []byte{10, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}},
					{IP: []byte{172, 69, 0, 0}, Mask: []byte{255, 255, 0, 0}},
					{IP: []byte{192, 168, 0, 0}, Mask: []byte{255, 255, 255, 0}},
				},
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "dev",
							action: accept,
						},
					},
				},
			},
			args{
				[]byte{127, 0, 0, 1},
				http.Header{},
			},
			false,
		},
		{
			"IP refused",
			fields{
				[]*net.IPNet{
					{IP: []byte{10, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}},
					{IP: []byte{172, 69, 0, 0}, Mask: []byte{255, 255, 0, 0}},
					{IP: []byte{192, 168, 0, 0}, Mask: []byte{255, 255, 255, 0}},
				},
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "dev",
							action: accept,
						},
					},
				},
			},
			args{
				[]byte{148, 0, 0, 1},
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
					"X-Version": []string{"1.0"},
				},
			},
			false,
		},
		{
			"token and IP refused",
			fields{
				[]*net.IPNet{
					{IP: []byte{10, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}},
					{IP: []byte{172, 69, 0, 0}, Mask: []byte{255, 255, 0, 0}},
					{IP: []byte{192, 168, 0, 0}, Mask: []byte{255, 255, 255, 0}},
				},
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "dev",
							action: accept,
						},
					},
				},
			},
			args{
				[]byte{149, 0, 0, 1},
				http.Header{},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, _ := logrustest.NewNullLogger()
			f := &Filter{
				restricted: tt.fields.restricted,
				token:      tt.fields.token,
				log:        log,
			}
			if got := f.IsConcerned(tt.args.ip, tt.args.parsedHeader); got != tt.want {
				t.Errorf("Filter.IsConcerned() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_isAuthorized(t *testing.T) {
	type fields struct {
		restricted []*net.IPNet
	}
	type args struct {
		ip net.IP
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantLog []string
	}{
		{
			"loopback",
			fields{[]*net.IPNet{}},
			args{[]byte{127, 0, 0, 1}},
			true,
			[]string{},
		},
		{
			"no restriction",
			fields{[]*net.IPNet{}},
			args{[]byte{192, 0, 0, 1}},
			true,
			[]string{},
		},
		{
			"not restricted",
			fields{[]*net.IPNet{
				{IP: []byte{10, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}},
				{IP: []byte{172, 69, 0, 0}, Mask: []byte{255, 255, 0, 0}},
				{IP: []byte{192, 168, 0, 0}, Mask: []byte{255, 255, 255, 0}},
			}},
			args{[]byte{192, 168, 0, 1}},
			true,
			[]string{},
		},
		{
			"restricted",
			fields{[]*net.IPNet{
				{IP: []byte{10, 0, 0, 0}, Mask: []byte{255, 0, 0, 0}},
				{IP: []byte{172, 69, 0, 0}, Mask: []byte{255, 255, 0, 0}},
				{IP: []byte{192, 168, 0, 0}, Mask: []byte{255, 255, 255, 0}},
			}},
			args{[]byte{8, 8, 8, 8}},
			false,
			[]string{"filter forbidden for this IP"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			f := &Filter{
				restricted: tt.fields.restricted,
				log:        log,
			}
			if got := f.isAuthorized(tt.args.ip); got != tt.want {
				t.Errorf("Filter.isAuthorized() = %v, want %v", got, tt.want)
			}
			verifyLogged("Filter.toFilter", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_isAccepted(t *testing.T) {
	type fields struct {
		token map[string][]headerConditions
	}
	type args struct {
		parsedHeader http.Header
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantLog []string
	}{
		{
			"minimal",
			fields{
				map[string][]headerConditions{},
			},
			args{
				http.Header{},
			},
			true,
			[]string{
				"Accepted",
			},
		},
		{
			"no corresponding accepted header",
			fields{
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "dev",
							action: accept,
						},
					},
				},
			},
			args{
				http.Header{},
			},
			false,
			[]string{
				"missing header for this filter",
			},
		},
		{
			"accepted headers",
			fields{
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "dev",
							action: accept,
						},
						{
							value:  "test",
							action: accept,
						},
					},
					"X-Authors": {
						{
							value:  "bob",
							action: accept,
						},
					},
				},
			},
			args{
				http.Header{
					"X-ENV":     []string{"test"},
					"X-Authors": []string{"alice", "bob"},
					"X-Version": []string{"1.0"},
				},
			},
			true,
			[]string{
				"lookup for condition",
				"lookup for condition",
				"Accepted",
			},
		},
		{
			"rejected headers",
			fields{
				map[string][]headerConditions{
					"X-Authors": {
						{
							value:  "bob",
							action: reject,
						},
					},
				},
			},
			args{
				http.Header{
					"X-ENV":     []string{"test"},
					"X-Authors": []string{"alice", "bob"},
					"X-Version": []string{"1.0"},
				},
			},
			false,
			[]string{
				"lookup for condition",
				"Refused",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)

			f := &Filter{
				token: tt.fields.token,
				log:   log,
			}
			if got := f.isAccepted(tt.args.parsedHeader); got != tt.want {
				t.Errorf("Filter.isAccepted() = %v, want %v", got, tt.want)
			}

			verifyLogged("Filter.toFilter", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_IsConditional(t *testing.T) {
	type fields struct {
		restricted []*net.IPNet
		token      map[string][]headerConditions
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			"no condition",
			fields{
				[]*net.IPNet{},
				map[string][]headerConditions{},
			},
			false,
		},
		{
			"restricted",
			fields{
				[]*net.IPNet{
					{
						IP:   []byte{192, 168, 1, 1},
						Mask: []byte{255, 255, 255, 0},
					},
				},
				map[string][]headerConditions{},
			},
			true,
		},
		{
			"token",
			fields{
				[]*net.IPNet{},
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "X-ENV",
							action: accept,
						},
					},
				},
			},
			true,
		},
		{
			"both conditions",
			fields{
				[]*net.IPNet{
					{
						IP:   []byte{192, 168, 1, 1},
						Mask: []byte{255, 255, 255, 0},
					},
				},
				map[string][]headerConditions{
					"X-ENV": {
						{
							value:  "X-ENV",
							action: accept,
						},
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Filter{
				restricted: tt.fields.restricted,
				token:      tt.fields.token,
			}
			if got := f.IsConditional(); got != tt.want {
				t.Errorf("Filter.IsConditional() = %v, want %v", got, tt.want)
			}
		})
	}
}
