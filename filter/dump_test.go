package filter

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func TestFilter_dumpHTTPMessage(t *testing.T) {
	// provide temporary directory for dump tests
	tmpDir, err := os.MkdirTemp(os.TempDir(), "villipdumpHTTPMessage")
	if err != nil {
		t.Fatalf("Not able to create temporary directory")
	}
	defer os.RemoveAll(tmpDir)

	type fields struct {
		dumpFolder string
		dumpURLs   []*regexp.Regexp
	}
	type args struct {
		generateID           func() (string, error)
		requestID            string
		requestIDFromRequest string
		url                  string
		header               http.Header
		body                 string
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		kind        string
		expectFatal bool
		want        string
		wantLog     []string
		wantContent string
	}{
		{
			"ID generation fails",
			fields{
				"",
				[]*regexp.Regexp{},
			},
			args{
				func() (string, error) {
					return "123456789012", fmt.Errorf("fake error")
				},
				"",
				"",
				"http://localhost:8081",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"request",
			true,
			"123456789012",
			[]string{
				"X-ENV: dev",
				"This is a body",
			},
			"",
		},
		{
			"minimal request",
			fields{
				"",
				[]*regexp.Regexp{},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"",
				"http://localhost:8081",
				http.Header{"Server": []string{"remote"}, "X-ENV": []string{"dev"}},
				"This is a body",
			},
			"request",
			false,
			"123456789012",
			[]string{
				"Server: remote",
				"X-ENV: dev",
				"This is a body",
			},
			"",
		},
		{
			"minimal response",
			fields{
				"",
				[]*regexp.Regexp{},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"01234567890",
				"http://localhost:8081",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"response",
			false,
			"01234567890",
			[]string{
				"X-ENV: dev",
				"This is a body",
			},
			"",
		},
		{
			"filtered",
			fields{
				"",
				[]*regexp.Regexp{
					regexp.MustCompile("/youngster"),
					regexp.MustCompile("/children"),
				},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"01234567890",
				"http://localhost:8081/boomer",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"response",
			false,
			"01234567890",
			[]string{},
			"",
		},
		{
			"not filtered",
			fields{
				"",
				[]*regexp.Regexp{
					regexp.MustCompile("/youngster"),
					regexp.MustCompile("/children"),
				},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"01234567890",
				"http://localhost:8081/youngster/1",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"response",
			false,
			"01234567890",
			[]string{
				"X-ENV: dev",
				"This is a body",
			},
			"",
		},
		{
			"error dump files",
			fields{
				filepath.Join(tmpDir, "dump"),
				[]*regexp.Regexp{},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"01234567890",
				"http://localhost:8081",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"response",
			true,
			"01234567890",
			[]string{
				"X-ENV: dev",
				"This is a body",
			},
			"",
		},
		{
			"ok dump files",
			fields{
				tmpDir,
				[]*regexp.Regexp{},
			},
			args{
				func() (string, error) {
					return "123456789012", nil
				},
				"",
				"01234567890",
				"http://localhost:8081",
				http.Header{"X-ENV": []string{"dev"}},
				"This is a body",
			},
			"Request",
			false,
			"01234567890",
			[]string{},
			"URL: http://localhost:8081\nX-ENV: dev\n\nThis is a body",
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
				url:        "http://localhost:8080",
				log:        log,
				dumpFolder: tt.fields.dumpFolder,
				dumpURLs:   tt.fields.dumpURLs,
			}

			_generateID = tt.args.generateID

			got := f.dumpHTTPMessage(tt.args.requestID, tt.args.requestIDFromRequest, tt.args.url, tt.args.header, tt.args.body)

			fatal := HadErrorLevel(hook, logrus.FatalLevel)
			if fatal != tt.expectFatal {
				t.Errorf("Filter.dumpHTTPMessage() fatal got = %v, want %v", fatal, tt.expectFatal)
			}

			if fatal {
				return
			}

			verifyLogged("Filter.dumpHTTPMessage", tt.wantLog, hook, t)

			if got != tt.want {
				t.Errorf("Filter.dumpHTTPMessage() = %v, want %v", got, tt.want)
			}

			if tt.fields.dumpFolder != "" {
				filePath := filepath.Join(tt.fields.dumpFolder, got+".original"+tt.kind)
				if _, err := os.Stat(filePath); os.IsNotExist(err) {
					t.Errorf("Filter.dumpHTTPMessage() file %s was not created", filePath)
				} else {
					content, err := ioutil.ReadFile(filePath)
					if err != nil {
						t.Fatal(err)
					}
					if string(content) != tt.wantContent {
						t.Errorf("Filter.dumpHTTPMessage() dumped %s , want %s", content, tt.wantContent)
					}
				}
			}
		})
	}
}

func Test_generateID(t *testing.T) {
	tests := []struct {
		name    string
		want    int
		wantErr bool
	}{
		{
			"ok",
			24,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := generateID()
			if (err != nil) != tt.wantErr {
				t.Errorf("generateID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.want {
				t.Errorf("generateID() = %d, want %d", len(got), tt.want)
			}
		})
	}
}
