package filter

import (
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func (fact *Factory) MockNewFromConfig(f fNewConfig) {
	fact.newFromConfig = f
}

func (fact *Factory) MockLookupEnv(f func(string) (string, bool)) {
	fact.lookupEnv = f
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
				prefix:   []replaceParameters{},
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
				response: response{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("^/youngster"),
								regexp.MustCompile("^/children"),
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
				response: response{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("^/youngster"),
								regexp.MustCompile("^/children"),
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
								regexp.MustCompile("^/youngster"),
								regexp.MustCompile("^/children"),
							},
						},
					},
					Header: []Cheader{},
				},
				prefix: []replaceParameters{},
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
				request: request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("^/youngster"),
								regexp.MustCompile("^/children"),
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
								regexp.MustCompile("^/boomer"),
								regexp.MustCompile("^/grandparent"),
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
			},
		},
		{
			"responserequestHeaderPrefix",
			args{
				Config{
					URL: "http://localhost:8081",
					Prefix: []Creplacement{
						{
							From: "/env",
							To:   "/dev/env",
							Urls: []string{"/env/admin", "/env/health"},
						},
					},
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
				prefix: []replaceParameters{
					{
						from: "/env",
						to:   "/dev/env",
						urls: []*regexp.Regexp{
							regexp.MustCompile("^/env/admin"),
							regexp.MustCompile("^/env/health"),
						},
					},
				},
				request: request{
					Replace: []replaceParameters{
						{
							from: "book",
							to:   "smartphone",
							urls: []*regexp.Regexp{
								regexp.MustCompile("^/youngster"),
								regexp.MustCompile("^/children"),
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
								regexp.MustCompile("^/boomer"),
								regexp.MustCompile("^/grandparent"),
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
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
				status:       []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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
				prefix:   []replaceParameters{},
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
				status:   []int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
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

			factory := NewFactory(log).(*Factory)
			got, got1, got2 := factory.newFromConfig(log, tt.args.c)

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
			g2 := got2.(*Filter)
			g2.log = nil

			if !reflect.DeepEqual(g2, tt.want2) {
				switch {
				case !reflect.DeepEqual(g2.response, tt.want2.response):
					t.Errorf("newFromConfig(response) \ngot2 = %#v,\nwant = %#v", g2.response, tt.want2.response)
					t.Errorf("newFromConfig(response.urls) \ngot2 = %#v,\nwant = %#v", g2.response.Replace[0].urls[0], tt.want2.response.Replace[0].urls[0])
				case !reflect.DeepEqual(g2.request, tt.want2.request):
					t.Errorf("newFromConfig(request) \ngot2 = %#v,\nwant = %#v", g2.request, tt.want2.request)
					t.Errorf("newFromConfig(request.urls) \ngot2 = %#v,\nwant = %#v", g2.request.Replace[0].urls[0], tt.want2.request.Replace[0].urls[0])
				case !reflect.DeepEqual(g2.prefix, tt.want2.prefix):
					t.Errorf("newFromConfig(prefix) \ngot2 = %#v,\nwant = %#v", g2.prefix, tt.want2.prefix)
				case !reflect.DeepEqual(g2.token, tt.want2.token):
					t.Errorf("newFromConfig(token) \ngot2 = %#v,\nwant = %#v", g2.token, tt.want2.token)
				default:
					t.Errorf("newFromConfig() \ngot2 = %#v,\nwant = %#v", g2, tt.want2)
				}
			}
		})
	}
}

func Test_parseReplaceConfig(t *testing.T) {
	type args struct {
		rep    []Creplacement
		prefix []replaceParameters
	}
	tests := []struct {
		name        string
		args        args
		expectFatal bool
		want        []replaceParameters
	}{
		{
			"simple",
			args{
				[]Creplacement{
					{
						From: "from",
						To:   "to",
						Urls: []string{"/"},
					},
				},
				[]replaceParameters{},
			},
			false,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{regexp.MustCompile("^/")},
				},
			},
		},
		{
			"multiple",
			args{
				[]Creplacement{
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
				},
				[]replaceParameters{},
			},
			false,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{
						regexp.MustCompile("^/test1/mul"),
						regexp.MustCompile("^/test2"),
					},
				},
				{
					from: "to",
					to:   "from",
					urls: []*regexp.Regexp{
						regexp.MustCompile("^/test3/mul"),
						regexp.MustCompile("^/test4"),
					},
				},
			},
		},
		{
			"error",
			args{
				[]Creplacement{
					{
						From: "from",
						To:   "to",
						Urls: []string{"/("},
					},
				},
				[]replaceParameters{},
			},
			true,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{regexp.MustCompile("/")},
				},
			},
		},
		{
			"prefixed",
			args{
				[]Creplacement{
					{
						From: "from",
						To:   "to",
						Urls: []string{"/test1/mul", "/test2/test2/mul"},
					},
					{
						From: "to",
						To:   "from",
						Urls: []string{"/test3/mul", "/test4/test3"},
					},
				},
				[]replaceParameters{{
					from: "/test3",
					to:   "",
					urls: []*regexp.Regexp{
						regexp.MustCompile("^/test3"),
					},
				},
					{
						from: "/test4",
						to:   "/test6",
						urls: []*regexp.Regexp{
							regexp.MustCompile("^/test4"),
						},
					},
					{
						from: "/test2",
						to:   "/dev/test2",
						urls: []*regexp.Regexp{
							regexp.MustCompile("^/test2"),
						},
					},
				},
			},
			false,
			[]replaceParameters{
				{
					from: "from",
					to:   "to",
					urls: []*regexp.Regexp{
						regexp.MustCompile("^/test1/mul"),
						regexp.MustCompile("^/dev/test2/test2/mul"),
					},
				},
				{
					from: "to",
					to:   "from",
					urls: []*regexp.Regexp{
						regexp.MustCompile("^/mul"),
						regexp.MustCompile("^/test6/test3"),
					},
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

			got := parseReplaceConfig(log, tt.args.rep, tt.args.prefix)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("parseReplaceConfig() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseReplaceConfig() = %v, want %v", got, tt.want)
				for i := range got {
					if !reflect.DeepEqual(got[i], tt.want[i]) {
						t.Errorf("parseReplaceConfig(i) = %v, want %v", got[i].urls, tt.want[i].urls)

					}
				}
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

func TestFilter_convertStatus(t *testing.T) {

	type args struct {
		statusList []string
	}
	tests := []struct {
		name        string
		args        args
		want        []int
		expectFatal bool
	}{
		{
			"default",
			args{
				[]string{},
			},
			[]int{http.StatusOK, http.StatusFound, http.StatusMovedPermanently},
			false,
		},
		{
			"correct",
			args{
				[]string{"404", "666"},
			},
			[]int{200, 302, 301, 404, 666},
			false,
		},
		{
			"negative status",
			args{
				[]string{"404", "-1", "666"},
			},
			[]int{},
			true,
		},
		{
			"wrong status",
			args{
				[]string{"404", "1234", "666"},
			},
			[]int{},
			true,
		},
		{
			"not convertible status",
			args{
				[]string{"404", "found", "666"},
			},
			[]int{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// Use logrus abilities to test log.Fatal
			log, hook := logrustest.NewNullLogger()
			log.ExitFunc = func(int) { return }
			defer func() { log.ExitFunc = nil }()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{
				log: log,
			}

			got := f.convertStatus(tt.args.statusList)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("convertStatus() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter.convertStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}
