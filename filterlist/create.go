package filterlist

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/marema31/villip/filter"
	"github.com/marema31/villip/server"
	"github.com/marema31/villip/server/http"
	"github.com/marema31/villip/server/tcp"
	"github.com/sirupsen/logrus"
)

// List contains a list of filter.
type List struct {
	filters map[string]map[uint8][]filter.FilteredServer
	factory filter.Creator
	// Make os.LookupEnv mockable for unit test.
	lookupEnv func(string) (string, bool)
}

// New returns a new empty filter list.
func New() *List {
	return &List{lookupEnv: os.LookupEnv, filters: make(map[string]map[uint8][]filter.FilteredServer)}
}

func (fl *List) insert(port string, priority uint8, f filter.FilteredServer) {
	if _, ok := fl.filters[port]; !ok {
		fl.filters[port] = make(map[uint8][]filter.FilteredServer)
	}

	if _, ok := fl.filters[port][priority]; !ok {
		fl.filters[port][priority] = make([]filter.FilteredServer, 0, 1)
	}

	if !f.IsConditional() {
		fl.filters[port][priority] = append(fl.filters[port][priority], f)
	} else {
		// Prepending filter to the list using golang tricks (the first inserted f will be replaced by the copy)
		fl.filters[port][priority] = append(fl.filters[port][priority], f)
		copy(fl.filters[port][priority][1:], fl.filters[port][priority])
		fl.filters[port][priority][0] = f
	}
}

func sortFilter(filters map[uint8][]filter.FilteredServer) []filter.FilteredServer {
	fl := make([]filter.FilteredServer, 0)

	priorities := make([]uint8, 0, 1)
	for priority := range filters {
		priorities = append(priorities, priority)
	}

	// Sort the priorities in descent order.
	sort.Slice(priorities, func(i, j int) bool { return priorities[i] > priorities[j] })

	for _, priority := range priorities {
		fl = append(fl, filters[priority]...)
	}

	return fl
}

func createServer(filters map[uint8][]filter.FilteredServer, port string, upLog logrus.FieldLogger) server.Server {
	var (
		s server.Server
	)

	for _, f := range sortFilter(filters) {
		if s != nil {
			if f.Kind() != filter.HTTP {
				upLog.Fatal("Cannot add a non HTTP filter to the same port than a HTTP proxy")
			}

			s.Insert(f)
		} else {
			switch f.Kind() {
			case filter.HTTP:
				s = http.New(upLog, port, f)
			case filter.TCP:
				s = tcp.New(upLog, port, f)
			}
		}
	}

	return s
}

// CreateServers creates all the server corresponding to the filters of the list.
func (fl *List) CreateServers(upLog logrus.FieldLogger) map[string]server.Server {
	servers := make(map[string]server.Server)
	for port := range fl.filters {
		servers[port] = createServer(fl.filters[port], port, upLog)
	}

	return servers
}

func (fl *List) readConfigFiles(upLog logrus.FieldLogger, folderPath string) {
	files, err := ioutil.ReadDir(folderPath)
	if err != nil {
		upLog.Fatalf("Error getting list of configuration files: %v", err)
	}

	for _, file := range files {
		ext := filepath.Ext(file.Name())

		switch {
		case file.Mode().IsRegular() && (ext == ".yml" || ext == ".yaml"):
			port, priority, f := fl.factory.NewFromYAML(filepath.Join(folderPath, file.Name()))
			fl.insert(port, priority, f)

		case file.Mode().IsRegular() && (ext == ".json"):
			port, priority, f := fl.factory.NewFromJSON(filepath.Join(folderPath, file.Name()))
			fl.insert(port, priority, f)
		default:
			continue
		}
	}
}

// ReadConfig fill the list with filter from the provided configurations.
func (fl *List) ReadConfig(upLog logrus.FieldLogger) {
	if fl.factory == nil {
		fl.factory = filter.NewFactory(upLog)
	}

	if _, ok := fl.lookupEnv("VILLIP_URL"); ok {
		port, priority, f := fl.factory.NewFromEnv()
		fl.insert(port, priority, f)
	}

	if folderPath, ok := fl.lookupEnv("VILLIP_FOLDER"); ok {
		fl.readConfigFiles(upLog, folderPath)
	}
}
