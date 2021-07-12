package filter

import (
	"net"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func MockNewFromConfig(f func(logrus.FieldLogger, Config) (string, uint8, *Filter)) {
	_newFromConfig = f
}

func MockLookupEnv(f func(string) (string, bool)) {
	_LookupEnv = f
}

func Test_newFromConfig(t *testing.T) {
	// provide temporary directory for dump tests
	tmpDir, err := os.MkdirTemp(os.TempDir(), "villipNewFromConfigTest")
	if err != nil {
		t.Fatalf("Not able to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)

	type args struct {
		c Config
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        string
		want1       uint8
		want2       *Filter
	}{
		{
			"NoUrl",
			args{Config{
				URL: "",
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"NegativePort",
			args{Config{
				URL:  "http://localhost:8080",
				Port: -1,
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"WrongPort",
			args{Config{
				URL:  "http://localhost:8080",
				Port: 67890,
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"WrongPort",
			args{Config{
				URL:  "http://localhost:8080",
				Port: 67890,
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"WrongDumpFolder",
			args{Config{
				URL: "http://localhost:8080",
				Dump: Cdump{
					Folder: "./testdata/emptyfile/dump",
				},
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"WrongDumpUrls",
			args{Config{
				URL: "http://localhost:8080",
				Dump: Cdump{
					Folder: filepath.Join(tmpDir, "WrongDumpUrls"),
					URLs:   []string{"/student", "/(teacher"},
				},
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"ResponseAndReplace",
			args{Config{
				URL: "http://localhost:8081",
				Replace: []Creplacement{
					{
						From: "",
						To:   "",
						Urls: []string{},
					},
				},
				Response: Caction{
					Replace: []Creplacement{
						{
							From: "",
							To:   "",
							Urls: []string{},
						},
					},
				},
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"WrongRestricted",
			args{Config{
				URL:        "http://localhost:8080",
				Restricted: []string{"192.168.1/24", "172.1.2.3/34", "192.168.2.1"},
			}},
			true,
			"8080",
			0,
			&Filter{},
		},
		{
			"default",
			args{Config{
				URL: "http://localhost:8081",
			}},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				response: response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				request: request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"insecureContentTypePortPriority",
			args{Config{
				URL:          "http://localhost:8081",
				Port:         9090,
				Priority:     100,
				ContentTypes: []string{"text/xml", "appplication/xsl"},
				Insecure:     true,
				Type:         "HTTP",
			}},
			false,
			"9090",
			100,
			&Filter{
				insecure: true,
				force:    false,
				response: response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				request: request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/xml", "appplication/xsl"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "9090",
				priority:     "100",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"replaceTypeForce",
			args{Config{
				URL: "http://localhost:8081",
				Replace: []Creplacement{
					{
						From: "book",
						To:   "smartphone",
						Urls: []string{"/youngster", "/children"},
					},
				},
				Force: true,
				Type:  "tcp",
			}},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    true,
				response: response{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				request: request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"responseReplace",
			args{
				Config{
					URL: "http://localhost:8081",
					Response: Caction{
						Replace: []Creplacement{
							{
								From: "book",
								To:   "smartphone",
								Urls: []string{"/youngster", "/children"},
							},
						},
					},
					Type: "uDp",
				},
			},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				response: response{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				request: request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"requestReplace",
			args{
				Config{
					URL: "http://localhost:8081",
					Request: Caction{
						Replace: []Creplacement{
							{
								From: "book",
								To:   "smartphone",
								Urls: []string{"/youngster", "/children"},
							},
						},
					},
				},
			},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				request: request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				response: response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"responserequestReplace",
			args{
				Config{
					URL: "http://localhost:8081",
					Request: Caction{
						Replace: []Creplacement{
							{
								From: "book",
								To:   "smartphone",
								Urls: []string{"/youngster", "/children"},
							},
						},
					},
					Response: Caction{
						Replace: []Creplacement{
							{
								From: "videogame",
								To:   "boardgame",
								Urls: []string{"/boomer", "/grandparent"},
							},
						},
					},
				},
			},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				request: request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				response: response{
					Replace: []replaceParameters{
						{
							from: "videogame",
							to:   "boardgame",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/boomer"),
								regexp.MustCompile("/grandparent"),
							},
						},
					},
					Header: []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"responserequestHeader",
			args{
				Config{
					URL: "http://localhost:8081",
					Request: Caction{
						Replace: []Creplacement{
							{
								From: "book",
								To:   "smartphone",
								Urls: []string{"/youngster", "/children"},
							},
						},
						Header: []Cheader{
							{
								Name:  "X-ENV",
								Value: "dev",
								Force: false,
							},
							{
								Name:  "X-Version",
								Value: "1.2",
								Force: true,
							},
						},
					},
					Response: Caction{
						Replace: []Creplacement{
							{
								From: "videogame",
								To:   "boardgame",
								Urls: []string{"/boomer", "/grandparent"},
							},
						},
						Header: []Cheader{
							{
								Name:  "X-TEST",
								Value: "valid",
							},
							{
								Name:  "X-Author",
								Value: "bob",
								Force: true,
							},
						},
					},
				},
			},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				request: request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/youngster"),
								regexp.MustCompile("/children"),
							},
						},
					},
					Header: []Cheader{
						{
							Name:  "X-ENV",
							Value: "dev",
							Force: false,
						},
						{
							Name:  "X-Version",
							Value: "1.2",
							Force: true,
						},
					},
				},
				response: response{
					Replace: []replaceParameters{
						{
							from: "videogame",
							to:   "boardgame",
							urls: []*regexp.Regexp{
								regexp.MustCompile("/boomer"),
								regexp.MustCompile("/grandparent"),
							},
						},
					},
					Header: []Cheader{
						{
							Name:  "X-TEST",
							Value: "valid",
							Force: false,
						},
						{
							Name:  "X-Author",
							Value: "bob",
							Force: true,
						},
					},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"onlyheader",
			args{
				Config{
					URL: "http://localhost:8081",
					Request: Caction{
						Header: []Cheader{
							{
								Name:  "X-ENV",
								Value: "dev",
								Force: false,
							},
						},
					},
					Response: Caction{
						Header: []Cheader{
							{
								Name:  "X-TEST",
								Value: "valid",
							},
							{
								Name:  "X-Author",
								Value: "bob",
								Force: true,
							},
							{
								Name:  "X-Version",
								Value: "1.2",
								Force: true,
							},
						},
					},
				},
			},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				request: request{
					Replace: []replaceParameters{},
					Header: []Cheader{
						{
							Name:  "X-ENV",
							Value: "dev",
							Force: false,
						},
					},
				},
				response: response{
					Replace: []replaceParameters{},
					Header: []Cheader{
						{
							Name:  "X-TEST",
							Value: "valid",
							Force: false,
						},
						{
							Name:  "X-Author",
							Value: "bob",
							Force: true,
						},
						{
							Name:  "X-Version",
							Value: "1.2",
							Force: true,
						},
					},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token:        map[string][]headerConditions{},
				url:          "http://localhost:8081",
				port:         "8080",
				priority:     "0",
				dumpURLs:     []*regexp.Regexp{},
			},
		},
		{
			"token",
			args{Config{
				URL: "http://localhost:8080",
				Token: []CtokenAction{
					{
						Header: "X-ENV",
						Value:  "dev",
						Action: "accept",
					},
					{
						Header: "X-ENV",
						Value:  "test",
						Action: "accept",
					},
					{
						Header: "X-Author",
						Value:  "bob",
						Action: "reject",
					},
				},
			}},
			false,
			"8080",
			0,
			&Filter{
				insecure: false,
				force:    false,
				response: response{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				request: request{
					Replace: []replaceParameters{},
					Header:  []Cheader{},
				},
				contentTypes: []string{"text/html", "text/css", "application/javascript"},
				restricted:   []*net.IPNet{},
				token: map[string][]headerConditions{
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
					"X-Author": {
						{
							value:  "bob",
							action: reject,
						},
					},
				},
				url:      "http://localhost:8080",
				port:     "8080",
				priority: "0",
				dumpURLs: []*regexp.Regexp{},
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
			tt.want2.log = nil

			got, got1, got2 := newFromConfig(log, tt.args.c)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("parseReplaceConfig() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if got != tt.want {
				t.Errorf("newFromConfig() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("newFromConfig() got1 = %v, want %v", got1, tt.want1)
			}

			// This field is not interesting for our tests and make DeepEqual impossible
			got2.log = nil

			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("newFromConfig() \ngot2 = %#v,\nwant = %#v", got2, tt.want2)
			}
		})
	}
}

func Test_parseReplaceConfig(t *testing.T) {
	type args struct {
		rep []Creplacement
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        []replaceParameters
	}{
		{
			"simple",
			args{[]Creplacement{
				{
					From: "from",
					To:   "to",
					Urls: []string{"/"},
				},
			}},
			false,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{regexp.MustCompile("/")},
				},
			},
		},
		{
			"multiple",
			args{[]Creplacement{
				{
					From: "from",
					To:   "to",
					Urls: []string{"/test1/mul", "/test2"},
				},
				{
					From: "to",
					To:   "from",
					Urls: []string{"/test3/mul", "/test4"},
				},
			}},
			false,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{
						regexp.MustCompile("/test1/mul"),
						regexp.MustCompile("/test2"),
					},
				},
				{
					from: "to",
					to:   "from",
					urls: []*regexp.Regexp{
						regexp.MustCompile("/test3/mul"),
						regexp.MustCompile("/test4"),
					},
				},
			},
		},
		{
			"error",
			args{[]Creplacement{
				{
					From: "from",
					To:   "to",
					Urls: []string{"/("},
				},
			}},
			true,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{regexp.MustCompile("/")},
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

			got := parseReplaceConfig(log, tt.args.rep)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("parseReplaceConfig() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseReplaceConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseTokenConfig(t *testing.T) {
	type args struct {
		tokenConfig CtokenAction
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        string
		want1       headerConditions
	}{
		{
			"accept",
			args{tokenConfig: CtokenAction{
				Header: "X-ENV",
				Value:  "test",
				Action: "Accept",
			}},
			false,
			"X-ENV",
			headerConditions{value: "test", action: accept},
		},
		{
			"Reject",
			args{tokenConfig: CtokenAction{
				Header: "ENV",
				Value:  "try",
				Action: "REJECT",
			}},
			false,
			"ENV",
			headerConditions{value: "try", action: reject},
		},
		{
			"NotEmpty",
			args{tokenConfig: CtokenAction{
				Header: "X-ENV",
				Value:  "test",
				Action: "NotEmpty",
			}},
			false,
			"X-ENV",
			headerConditions{value: "test", action: notEmpty},
		},
		{
			"empty",
			args{tokenConfig: CtokenAction{
				Header: "",
				Value:  "test",
				Action: "accept",
			}},
			true,
			"X-ENV",
			headerConditions{value: "test", action: accept},
		},
		{
			"dummy",
			args{tokenConfig: CtokenAction{
				Header: "X-ENV",
				Value:  "test",
				Action: "dummy",
			}},
			true,
			"X-ENV",
			headerConditions{value: "test", action: accept},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use logrus abilities to test log.Fatal
			log, hook := logrustest.NewNullLogger()
			log.ExitFunc = func(int) { return }
			defer func() { log.ExitFunc = nil }()
			log.SetLevel(logrus.DebugLevel)

			got, got1 := parseTokenConfig(log, tt.args.tokenConfig)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("parseTokenConfig() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if got != tt.want {
				t.Errorf("parseTokenConfig() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("parseTokenConfig() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
