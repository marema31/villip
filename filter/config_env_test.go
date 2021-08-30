package filter_test

import (
	"reflect"
	"testing"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestNewFromEnv(t *testing.T) {
	type args struct {
		env map[string]string
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        filter.Config
	}{
		{
			"no URL",
			args{make(map[string]string)},
			true,
			filter.Config{},
		},
		{
			"string priority",
			args{map[string]string{
				"VILLIP_URL":      "http://localhost:8081",
				"VILLIP_PRIORITY": "high",
			}},
			true,
			filter.Config{},
		},
		{
			"negative priority",
			args{map[string]string{
				"VILLIP_URL":      "http://localhost:8081",
				"VILLIP_PRIORITY": "-1",
			}},
			true,
			filter.Config{},
		},
		{
			"to high priority",
			args{map[string]string{
				"VILLIP_URL":      "http://localhost:8081",
				"VILLIP_PRIORITY": "20000",
			}},
			true,
			filter.Config{},
		},
		{
			"string port",
			args{map[string]string{
				"VILLIP_URL":  "http://localhost:8081",
				"VILLIP_PORT": "boat",
			}},
			true,
			filter.Config{},
		},
		{
			"missing to",
			args{map[string]string{
				"VILLIP_URL":  "http://localhost:8081",
				"VILLIP_FROM": "boat",
			}},
			true,
			filter.Config{},
		},
		{
			"missing to",
			args{map[string]string{
				"VILLIP_URL":    "http://localhost:8081",
				"VILLIP_FROM":   "boat",
				"VILLIP_TO":     "ship",
				"VILLIP_FROM_1": "car",
				"VILLIP_TO_1":   "char",
				"VILLIP_FROM_2": "plane",
			}},
			true,
			filter.Config{},
		},
		{
			"minimal",
			args{map[string]string{
				"VILLIP_URL": "http://localhost:8081",
			}},
			false,
			filter.Config{
				Dump: filter.Cdump{
					Folder: "",
					URLs:   []string(nil),
				},
				Force:    false,
				Insecure: false,
				Port:     8080,
				Prefix:   []filter.Creplacement{},
				Priority: 0,
				Replace:  []filter.Creplacement{},
				Request: filter.Caction{
					Replace: []filter.Creplacement{},
					Header:  []filter.Cheader{},
				},
				Response: filter.Caction{
					Replace: []filter.Creplacement{},
					Header:  []filter.Cheader{},
				},
				Restricted: []string(nil),
				Token:      []filter.CtokenAction(nil),
				Type:       "",
				URL:        "http://localhost:8081",
			},
		},
		{
			"maximal",
			args{map[string]string{
				"VILLIP_URL":         "http://localhost:1234/url1",
				"VILLIP_PORT":        "8081",
				"VILLIP_PRIORITY":    "100",
				"VILLIP_FORCE":       "1",
				"VILLIP_INSECURE":    "1",
				"VILLIP_DUMPFOLDER":  "/var/log/villip/dump",
				"VILLIP_DUMPURLS":    "/books/,/movies/",
				"VILLIP_FROM":        "book",
				"VILLIP_TO":          "smartphone",
				"VILLIP_FOR":         "/youngsters/",
				"VILLIP_FROM_1":      "dance",
				"VILLIP_TO_1":        "chat",
				"VILLIP_FOR_1":       "/youngsters/,/geeks/",
				"VILLIP_TYPES":       "text/html,application/json",
				"VILLIP_RESTRICTED":  "192.168.1.0/24,192.168.8.0/24",
				"VILLIP_PREFIX_FROM": "/env/",
				"VILLIP_PREFIX_TO":   "/",
			}},
			false,
			filter.Config{
				ContentTypes: []string{"text/html", "application/json"},
				Dump: filter.Cdump{
					Folder: "/var/log/villip/dump",
					URLs:   []string{"/books/", "/movies/"},
				},
				Force:    true,
				Insecure: true,
				Port:     8081,
				Prefix: []filter.Creplacement{
					{
						From: "/env/",
						To:   "/",
						Urls: []string{},
					},
				},
				Priority: 100,
				Replace:  []filter.Creplacement{},
				Response: filter.Caction{
					Replace: []filter.Creplacement{
						{
							From: "book",
							To:   "smartphone",
							Urls: []string{"/youngsters/"},
						},
						{
							From: "dance",
							To:   "chat",
							Urls: []string{"/youngsters/", "/geeks/"},
						},
					},
					Header: []filter.Cheader{},
				},
				Request: filter.Caction{
					Replace: []filter.Creplacement{},
					Header:  []filter.Cheader{},
				},
				Restricted: []string{"192.168.1.0/24", "192.168.8.0/24"},
				Token:      []filter.CtokenAction(nil),
				Type:       "",
				URL:        "http://localhost:1234/url1",
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

			factory := filter.NewFactory(log).(*filter.Factory)
			// Mock newFromConfig
			var got filter.Config
			factory.MockNewFromConfig(func(log logrus.FieldLogger, c filter.Config) (string, uint8, filter.FilteredServer) {
				got = c
				return "", 0, &filter.Filter{}
			})

			// Mock os.LookupEnv
			factory.MockLookupEnv(func(key string) (string, bool) {
				value, ok := tt.args.env[key]
				return value, ok
			})

			factory.NewFromEnv()

			fatal := filter.HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("NewFromEnv() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromEnv() \ngot  = %#v \nwant = %#v", got, tt.want)
			}
		})
	}
}
