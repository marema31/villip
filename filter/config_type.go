package filter

// Creplacement configuration parts for replacement.
type Creplacement struct {
	From string   `yaml:"from" json:"from"`
	To   string   `yaml:"to" json:"to"`
	Urls []string `yaml:"urls" json:"urls"`
}

// Cdump configuration parts for dump log.
type Cdump struct {
	Folder string   `yaml:"folder" json:"folder"`
	URLs   []string `yaml:"urls" json:"urls"`
}

// Cheader configuration parts for header management.
type Cheader struct {
	Name  string `yaml:"name" json:"name"`
	Value string `yaml:"value" json:"value"`
	Force bool   `yaml:"force" json:"force"`
	Add   bool   `yaml:"add" json:"add"`
	UUID  bool   `yaml:"uuid" json:"uuid"`
}

// Caction configuration parts for request and response  management.
type Caction struct {
	Replace []Creplacement `yaml:"replace" json:"replace"`
	Header  []Cheader      `yaml:"header" json:"header"`
}

// CtokenAction configuration parts for token management.
type CtokenAction struct {
	Header string `yaml:"header" json:"header"`
	Value  string `yaml:"value" json:"value"`
	Action string `yaml:"action" json:"action"`
}

// Config common structure for configuration regardless the source.
type Config struct {
	ContentTypes []string       `yaml:"content-types" json:"content-types"` //nolint: tagliatelle
	Dump         Cdump          `yaml:"dump" json:"dump"`
	Force        bool           `yaml:"force" json:"force"`
	Insecure     bool           `yaml:"insecure" json:"insecure"`
	Port         int            `yaml:"port" json:"port"`
	Prefix       []Creplacement `yaml:"prefix" json:"prefix"`
	Priority     uint8          `yaml:"priority" json:"priority"`
	Replace      []Creplacement `yaml:"replace" json:"replace"`
	Request      Caction        `yaml:"request" json:"request"`
	Response     Caction        `yaml:"response" json:"response"`
	Restricted   []string       `yaml:"restricted" json:"restricted"`
	Status       []string       `yaml:"status" json:"status"`
	Token        []CtokenAction `yaml:"token" json:"token"`
	Type         string         `yaml:"type" json:"type"`
	URL          string         `yaml:"url" json:"url"`
}
