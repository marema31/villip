package filterlist

import (
	"net"
	"net/http"
	"path"
	"strconv"
	"strings"
	"testing"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

type mockFilter struct {
	position    int
	toInsert    bool
	conditional bool
}

func (m *mockFilter) IsConcerned(ip net.IP, h http.Header) bool {
	return m.toInsert
}

func (m *mockFilter) IsConditional() bool {
	return m.conditional
}

func (m *mockFilter) Serve(res http.ResponseWriter, req *http.Request) {
}

func (m *mockFilter) PrefixReplace(URL string) string {
	return URL
}

func TestList_insert(t *testing.T) {
	type fields struct {
		filters map[string]map[uint8][]filter.FilteredServer
	}
	type args struct {
		port        string
		priority    uint8
		conditional bool
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		position int
	}{
		{
			"First insert",
			fields{
				make(map[string]map[uint8][]filter.FilteredServer),
			},
			args{
				"8080",
				1,
				false,
			},
			0,
		},
		{
			"First in port",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8081": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			args{
				"8080",
				1,
				false,
			},
			0,
		},
		{
			"First in priority",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			args{
				"8080",
				1,
				false,
			},
			0,
		},
		{
			"Insert after",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			args{
				"8080",
				10,
				false,
			},
			1,
		},
		{
			"Insert before",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			args{
				"8080",
				10,
				true,
			},
			0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := New()
			fl.filters = tt.fields.filters
			fl.insert(tt.args.port, tt.args.priority, &mockFilter{1, true, tt.args.conditional})

			_, ok := fl.filters[tt.args.port]
			if !ok {
				t.Error("The filter was not inserted for the right port")
				return
			}

			list, ok := fl.filters[tt.args.port][tt.args.priority]
			if !ok {
				t.Error("The filter was not inserted for the right priority")
				return
			}

			if len(list) < tt.position {
				t.Error("List is too short")
				return
			}

			if !list[tt.position].IsConcerned([]byte{127, 0, 0, 1}, http.Header{}) {
				t.Error("The filter was not inserted at the right position")
			}
		})
	}
}

func Test_sortFilter(t *testing.T) {
	type args struct {
		filters map[uint8][]filter.FilteredServer
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"Normal",
			args{
				map[uint8][]filter.FilteredServer{
					1: {
						&mockFilter{2, false, true},
						&mockFilter{3, false, false},
					},
					10: {
						&mockFilter{0, true, true},
						&mockFilter{1, true, false},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sortFilter(tt.args.filters)

			for i, f := range got {
				mf, ok := f.(*mockFilter)
				if !ok {
					t.Error("The list something else than a mock... weird")
					return
				}

				if mf.position != i {
					t.Errorf("element %d was not at the expected index but at %d", mf.position, i)
				}
			}
		})

	}
}

func TestList_CreateServers(t *testing.T) {
	type fields struct {
		filters map[string]map[uint8][]filter.FilteredServer
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{
			"normal",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
							&mockFilter{1, false, false},
						},
					},
					"8081": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			[]string{"8080", "8081"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := New()
			fl.filters = tt.fields.filters

			got := fl.CreateServers(logrus.New())

			if len(got) != len(tt.want) {
				t.Errorf("wrong number of server wanted %d, got %d", len(tt.want), len(got))
			}
			for _, port := range tt.want {
				if _, ok := got[port]; !ok {
					t.Errorf("no server for port %s", port)
				}
			}
		})
	}
}

func TestList_readConfigFiles(t *testing.T) {
	type fields struct {
		filters map[string]map[uint8][]filter.FilteredServer
	}
	type args struct {
		folderPath string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		expectFatal bool
		want        map[string]map[uint8]int
	}{
		{
			"no folder",
			fields{
				map[string]map[uint8][]filter.FilteredServer{},
			},
			args{
				"dummy",
			},
			true,
			map[string]map[uint8]int{},
		},
		{
			"normal",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
					"8081": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
			},
			args{
				"testdata",
			},
			false,
			map[string]map[uint8]int{
				"8080": {
					10: 1,
					1:  1,
				},
				"8081": {
					10: 3,
				},
				"8082": {
					11: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use logrus abilities to test log.Fatal
			log, hook := logrustest.NewNullLogger()
			log.ExitFunc = func(int) { return }
			defer func() { log.ExitFunc = nil }()
			log.SetLevel(logrus.DebugLevel)

			fl := New()
			fl.filters = tt.fields.filters
			fl.factory = &MockCreator{}

			fl.readConfigFiles(log, tt.args.folderPath)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			checkListFilters(t, fl.filters, tt.want)
		})
	}
}

func TestList_ReadConfig(t *testing.T) {
	type fields struct {
		filters map[string]map[uint8][]filter.FilteredServer
		factory filter.Creator
	}
	type args struct {
		env map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   map[string]map[uint8]int
	}{
		{
			"No initial factory",
			fields{
				map[string]map[uint8][]filter.FilteredServer{},
				nil,
			},
			args{map[string]string{}},
			map[string]map[uint8]int{},
		},
		{
			"normal",
			fields{
				map[string]map[uint8][]filter.FilteredServer{
					"8080": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
					"8081": {
						10: []filter.FilteredServer{
							&mockFilter{0, false, false},
						},
					},
				},
				&MockCreator{},
			},
			args{map[string]string{
				"VILLIP_URL":    "http://localhost:8081",
				"VILLIP_FROM":   "boat",
				"VILLIP_TO":     "ship",
				"VILLIP_FROM_1": "car",
				"VILLIP_TO_1":   "char",
				"VILLIP_FROM_2": "plane",
				"VILLIP_FOLDER": "testdata",
			}},
			map[string]map[uint8]int{
				"8080": {
					10: 2,
					1:  1,
				},
				"8081": {
					10: 3,
				},
				"8082": {
					11: 2,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fl := New()
			fl.filters = tt.fields.filters
			fl.lookupEnv = func(key string) (string, bool) {
				value, ok := tt.args.env[key]
				return value, ok
			}

			fl.factory = tt.fields.factory
			fl.ReadConfig(logrus.New())

			if fl.factory == nil {
				t.Error("no factory was initialized")
			}

			checkListFilters(t, fl.filters, tt.want)
		})
	}
}

// helper for readConfig and fl.readConfigFiles tests
func checkListFilters(t *testing.T, filters map[string]map[uint8][]filter.FilteredServer, want map[string]map[uint8]int) {
	if len(filters) != len(want) {
		t.Errorf("wrong number of filter ports wanted %d, got %d", len(want), len(filters))
	}

	for port, priorities := range want {
		if _, ok := filters[port]; !ok {
			t.Errorf("no server for port %s", port)
		}

		if len(priorities) != len(filters[port]) {
			t.Errorf("wrong number of priorities for %s wanted %d, got %d", port, len(priorities), len(filters[port]))
		}

		for priority, count := range priorities {
			if _, ok := filters[port][priority]; !ok {
				t.Errorf("no server for port %s", port)
			}
			if count != len(filters[port][priority]) {
				t.Errorf("wrong number of filter for %s/%d wanted %d, got %d", port, priority, count, len(filters[port][priority]))
			}
		}
	}
}

type MockCreator struct {
}

func (mc *MockCreator) NewFromYAML(filepath string) (string, uint8, filter.FilteredServer) {
	_, filename := path.Split(filepath)
	elmt := strings.Split(filename, "_")
	port := elmt[0]
	priority, _ := strconv.Atoi(elmt[1][:strings.Index(elmt[1], ".")])
	return port, uint8(priority), &filter.Filter{}
}

func (mc *MockCreator) NewFromJSON(filepath string) (string, uint8, filter.FilteredServer) {
	_, filename := path.Split(filepath)
	elmt := strings.Split(filename, "_")
	port := elmt[0]
	priority, _ := strconv.Atoi(elmt[1][:strings.Index(elmt[1], ".")])
	return port, uint8(priority), &filter.Filter{}
}

func (mc *MockCreator) NewFromEnv() (string, uint8, filter.FilteredServer) {
	return "8080", 10, &filter.Filter{}
}
