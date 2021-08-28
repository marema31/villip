package filter

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	logrustest "github.com/sirupsen/logrus/hooks/test"
)

func Test_do(t *testing.T) {
	type args struct {
		url string
		s   string
		rep []replaceParameters
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"nothing",
			args{
				"http://localhost:8080/youngster",
				"take your book,\ntry to dance\n sing often",
				[]replaceParameters{},
			},
			"take your book,\ntry to dance\n sing often",
		},
		{
			"replaced",
			args{
				"http://localhost:8080/youngster",
				"take your book,\ntry to dance\n sing often",
				[]replaceParameters{
					{
						from: "videogame",
						to:   "boardgame",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/boomer"),
							regexp.MustCompile("/grandparent"),
						},
					},
					{
						from: "sing",
						to:   "chat",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/youngster"),
							regexp.MustCompile("/children"),
						},
					},
					{
						from: "book",
						to:   "smartphone",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/youngster"),
							regexp.MustCompile("/children"),
						},
					},
				},
			},
			"take your smartphone,\ntry to dance\n chat often",
		},
		{
			"replaced",
			args{
				"http://localhost:8080/parent",
				"take your book,\ntry to dance\n sing often",
				[]replaceParameters{
					{
						from: "videogame",
						to:   "boardgame",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/boomer"),
							regexp.MustCompile("/grandparent"),
						},
					},
					{
						from: "sing",
						to:   "chat",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/youngster"),
							regexp.MustCompile("/children"),
						},
					},
					{
						from: "book",
						to:   "smartphone",
						urls: []*regexp.Regexp{
							regexp.MustCompile("/youngster"),
							regexp.MustCompile("/children"),
						},
					},
				},
			},
			"take your book,\ntry to dance\n sing often",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := do(tt.args.url, tt.args.s, tt.args.rep); got != tt.want {
				t.Errorf("do() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_headerReplace(t *testing.T) {
	type args struct {
		parsedHeader http.Header
		headerConfig []Cheader
	}
	tests := []struct {
		name    string
		args    args
		want    http.Header
		wantLog []string
	}{
		{
			"empty",
			args{
				http.Header{},
				[]Cheader{},
			},
			http.Header{},
			[]string{"Checking if need to replace header"},
		},
		{
			"no config",
			args{
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
				[]Cheader{},
			},
			http.Header{
				"X-ENV":     []string{"dev"},
				"X-Authors": []string{"alice", "bob"},
			},
			[]string{"Checking if need to replace header"},
		},
		{
			"no replacement",
			args{
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
				[]Cheader{
					{
						Name:  "X-Authors",
						Value: "Charly",
						Force: false,
					},
				},
			},
			http.Header{
				"X-ENV":     []string{"dev"},
				"X-Authors": []string{"alice", "bob"},
			},
			[]string{"Checking if need to replace header"},
		},
		{
			"force replacement",
			args{
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
				[]Cheader{
					{
						Name:  "X-Authors",
						Value: "Charly",
						Force: true,
					},
					{
						Name:  "X-VERSION",
						Value: "1.0",
					},
					{
						Name:  "X-TIME",
						Value: "123456",
						Force: false,
					},
					{
						Name:  "X-ENV",
						Value: "prod",
					},
				},
			},
			http.Header{
				"X-Authors": []string{"Charly"},
				"X-VERSION": []string{"1.0"},
				"X-TIME":    []string{"123456"},
				"X-ENV":     []string{"prod"},
			},
			[]string{
				"Checking if need to replace header",
				"Set header X-Authors with value :  Charly",
				"Set header X-VERSION with value :  1.0",
				"Set header X-TIME with value :  123456",
				"Set header X-ENV with value :  prod",
			},
		},
		{
			"add",
			args{
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
				[]Cheader{
					{
						Name:  "X-Authors",
						Value: "Charly",
						Add:   true,
					},
					{
						Name:  "X-VERSION",
						Value: "1.0",
						Add:   true,
					},
					{
						Name:  "X-TIME",
						Value: "123456",
						Force: false,
					},
					{
						Name:  "X-ENV",
						Value: "prod",
						Add:   false,
					},
				},
			},
			http.Header{
				"X-Authors": []string{"alice", "bob", "Charly"},
				"X-VERSION": []string{"1.0"},
				"X-TIME":    []string{"123456"},
				"X-ENV":     []string{"prod"},
			},
			[]string{
				"Checking if need to replace header",
				"Adding to header X-Authors with value :  Charly",
				"Adding to header X-VERSION with value :  1.0",
				"Set header X-TIME with value :  123456",
				"Set header X-ENV with value :  prod",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{}
			f.headerReplace(log, tt.args.parsedHeader, tt.args.headerConfig)

			if !reflect.DeepEqual(tt.args.parsedHeader, tt.want) {
				t.Errorf("Filter.headerReplace() \ngot = %#v,\nwant = %#v", tt.args.parsedHeader, tt.want)
			}

			verifyLogged("Filter.headerReplace", tt.wantLog, hook, t)
		})
	}
}

func TestFilter_readAndReplaceBody(t *testing.T) {
	type args struct {
		bod          io.ReadCloser
		newbod       string
		parsedHeader http.Header
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   io.ReadCloser
		want2   string
		wantErr bool
		wantLog []string
	}{
		{
			"plain",
			args{
				ioutil.NopCloser(strings.NewReader("hello world")),
				"dlrow olleh. hello earth",
				http.Header{
					"X-ENV":     []string{"dev"},
					"X-Authors": []string{"alice", "bob"},
				},
			},
			24,
			ioutil.NopCloser(bytes.NewBufferString("dlrow olleh. hello earth")),
			"hello world",
			false,
			[]string{"Body before the replacement : hello world", "Body after the replacement : dlrow olleh. hello earth"},
		},
		{
			"wrong gzip",
			args{
				ioutil.NopCloser(strings.NewReader("hello world")),
				"",
				http.Header{
					"Content-Encoding": []string{"gzip"},
					"X-Authors":        []string{"alice", "bob"},
				},
			},
			0,
			nil,
			"",
			true,
			[]string{"Impossible to decompress: gzip: invalid header"},
		},
		{
			"gzip",
			args{
				ioutil.NopCloser(
					bytes.NewBuffer(
						[]byte{
							0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0xca, 0x48, 0xcd, 0xc9, 0xc9, 0x57,
							0x28, 0xcf, 0x2f, 0xca, 0x49, 0x1, 0x0, 0x0, 0x0, 0xff, 0xff, 0x1, 0x0, 0x0, 0xff, 0xff,
							0x85, 0x11, 0x4a, 0xd, 0xb, 0x0, 0x0, 0x0,
						},
					),
				),
				"dlrow olleh. hello earth",
				http.Header{
					"Content-Encoding": []string{"gzip"},
					"X-Authors":        []string{"alice", "bob"},
				},
			},
			53,
			ioutil.NopCloser(
				bytes.NewBuffer(
					[]byte{
						0x1f, 0x8b, 0x8, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0xff, 0x4a, 0xc9, 0x29, 0xca, 0x2f, 0x57,
						0xc8, 0xcf, 0xc9, 0x49, 0xcd, 0xd0, 0x53, 0xc8, 0x48, 0xcd, 0xc9, 0xc9, 0x57, 0x48, 0x4d,
						0x2c, 0x2a, 0xc9, 0x0, 0x0, 0x0, 0x0, 0xff, 0xff, 0x1, 0x0, 0x0, 0xff, 0xff, 0xf8, 0xc5,
						0x6e, 0x16, 0x18, 0x0, 0x0, 0x0,
					},
				),
			),
			"hello world",
			false,
			[]string{"Body before the replacement : hello world", "Body after the replacement : dlrow olleh. hello earth"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, hook := logrustest.NewNullLogger()
			log.SetLevel(logrus.DebugLevel)
			f := &Filter{
				log: log,
			}

			oldDo := _do
			_do = func(url string, s string, rep []replaceParameters) string {
				return tt.args.newbod
			}
			defer func() { _do = oldDo }()

			got, got1, got2, got3, err := f.readAndReplaceBody("http://localhost:8080", []replaceParameters{}, tt.args.bod, tt.args.parsedHeader)

			if (err != nil) != tt.wantErr {
				t.Errorf("Filter.readAndReplaceBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("Filter.readAndReplaceBody() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Filter.readAndReplaceBody() got1 = %v, want %#v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("Filter.readAndReplaceBody() got2 = %v, want %v", got2, tt.want2)
			}
			if got3 != tt.args.newbod {
				t.Errorf("Filter.readAndReplaceBody() got3 = %v, want %v", got3, tt.args.newbod)
			}

			verifyLogged("Filter.readAndReplaceBody", tt.wantLog, hook, t)
		})
	}
}
