package filter_test

import (
	"reflect"
	"testing"

	"github.com/marema31/villip/filter"
	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestNewFromYAML(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        filter.Config
	}{
		{
			"filenofound",
			args{"./testdata/nonexist.yaml"},
			true,
			filter.Config{},
		},
		{
			"notyaml",
			args{"./testdata/notyaml.yaml"},
			true,
			filter.Config{},
		},
		{
			"minimal",
			args{"./testdata/minimal.yaml"},
			false,
			filter.Config{
				URL: "http://localhost:8081",
			},
		},
		{
			"maximal",
			args{"./testdata/maximal.yaml"},
			false,
			filter.Config{
				ContentTypes: []string{"text/html", "application/json"},
				Dump: filter.Cdump{
					Folder: "/var/log/villip/dump",
					URLs:   []string{"/books/", "/movies/"},
				},
				Force:    true,
				Insecure: false,
				Port:     8081,
				Priority: 0,
				Replace:  []filter.Creplacement(nil),
				Request: filter.Caction{
					Replace: []filter.Creplacement{
						{
							From: "book",
							To:   "smartphone",
							Urls: []string{"/youngster/"},
						},
						{
							From: "dance",
							To:   "chat",
							Urls: []string{"/youngsters/", "/geeks/"},
						},
					},
					Header: []filter.Cheader{
						{
							Name:  "X-community",
							Value: "In real life",
							Force: false,
						},
					},
				},
				Response: filter.Caction{
					Replace: []filter.Creplacement{
						{
							From: "book",
							To:   "smartphone",
							Urls: []string{"/youngster/"},
						},
						{
							From: "dance",
							To:   "chat",
							Urls: []string{"/youngsters/", "/geeks/"},
						},
						{
							From: "meeting",
							To:   "texting",
							Urls: []string(nil),
						},
					}, Header: []filter.Cheader{
						{
							Name:  "X-community",
							Value: "In real life",
							Force: false,
						},
					},
				},
				Restricted: []string{"192.168.1.0/24", "192.168.8.0/24"},
				Token: []filter.CtokenAction{
					{
						Header: "X-MY-TOKEN",
						Value:  "123",
						Action: "accept",
					},
					{
						Header: "X-MY-TOKEN",
						Value:  "456",
						Action: "accept",
					},
					{
						Header: "X-MY-TOKEN",
						Value:  "789",
						Action: "reject",
					},
					{
						Header: "X-MY-SECONDTOKEN",
						Value:  "ABC", Action: "accept",
					},
					{
						Header: "X-MY-THIRDTOKEN",
						Value:  "",
						Action: "notempty",
					},
				},
				Type: "http",
				URL:  "http://localhost:1234/url1",
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

			factory.NewFromYAML(tt.args.filePath)

			fatal := filter.HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("NewFromYAML() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromYAML() \ngot  = %#v, \nwant = %#v", got, tt.want)
			}
		})
	}
}

func TestNewFromJSON(t *testing.T) {
	type args struct {
		filePath string
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        filter.Config
	}{
		{
			"filenofound",
			args{"./testdata/nonexist.json"},
			true,
			filter.Config{},
		},
		{
			"notyaml",
			args{"./testdata/notyaml.json"},
			true,
			filter.Config{},
		},
		{
			"minimal",
			args{"./testdata/minimal.json"},
			false,
			filter.Config{
				URL: "http://localhost:8081",
			},
		},
		{
			"maximal",
			args{"./testdata/maximal.json"},
			false,
			filter.Config{
				ContentTypes: []string{"text/html", "application/json"},
				Dump: filter.Cdump{
					Folder: "/var/log/villip/dump",
					URLs:   []string{"/books/", "/movies/"},
				},
				Force:    true,
				Insecure: false,
				Port:     8081,
				Priority: 0,
				Replace:  []filter.Creplacement(nil),
				Request: filter.Caction{
					Replace: []filter.Creplacement{
						{
							From: "book",
							To:   "smartphone",
							Urls: []string{"/youngster/"},
						},
						{
							From: "dance",
							To:   "chat",
							Urls: []string{"/youngsters/", "/geeks/"},
						},
					},
					Header: []filter.Cheader{
						{
							Name:  "X-community",
							Value: "In real life",
							Force: false,
						},
					},
				},
				Response: filter.Caction{
					Replace: []filter.Creplacement{
						{
							From: "book",
							To:   "smartphone",
							Urls: []string{"/youngster/"},
						},
						{
							From: "dance",
							To:   "chat",
							Urls: []string{"/youngsters/", "/geeks/"},
						},
						{
							From: "meeting",
							To:   "texting",
							Urls: []string(nil),
						},
					}, Header: []filter.Cheader{
						{
							Name:  "X-community",
							Value: "In real life",
							Force: false,
						},
					},
				},
				Restricted: []string{"192.168.1.0/24", "192.168.8.0/24"},
				Token: []filter.CtokenAction{
					{
						Header: "X-MY-TOKEN",
						Value:  "123",
						Action: "accept",
					},
					{
						Header: "X-MY-TOKEN",
						Value:  "456",
						Action: "accept",
					},
					{
						Header: "X-MY-TOKEN",
						Value:  "789",
						Action: "reject",
					},
					{
						Header: "X-MY-SECONDTOKEN",
						Value:  "ABC", Action: "accept",
					},
					{
						Header: "X-MY-THIRDTOKEN",
						Value:  "",
						Action: "notempty",
					},
				},
				Type: "http",
				URL:  "http://localhost:1234/url1",
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

			factory.NewFromJSON(tt.args.filePath)

			fatal := filter.HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("NewFromJSON() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFromJSON() \ngot  = %#v, \nwant = %#v", got, tt.want)
			}
		})
	}
}
