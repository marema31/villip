package filter

// Configuration  for replacement.
type Creplacement struct {
	From string `yaml:"from" json:"from,omitempty"`
	To   string `yaml:"to" json:"to,omitempty"`
	// +kubebuilder:validation:Optional
	Urls []string `yaml:"urls" json:"urls,omitempty"`
}

// Configuration for dump log.
type Cdump struct {
	Folder string `yaml:"folder" json:"folder,omitempty"`
	// +kubebuilder:validation:Optional
	URLs []string `yaml:"urls" json:"urls,omitempty"`
}

// Configuration for header management.
type Cheader struct {
	Name  string `yaml:"name" json:"name,omitempty"`
	Value string `yaml:"value" json:"value,omitempty"`
	Force bool   `yaml:"force" json:"force,omitempty"`
	// +kubebuilder:default=false
	Add bool `yaml:"add" json:"add,omitempty"`
	// +kubebuilder:validation:Optional
	UUID bool `yaml:"uuid" json:"uuid,omitempty"`
}

// Configuration for request and response  management.
type Caction struct {
	Replace []Creplacement `yaml:"replace" json:"replace,omitempty"`
	Header  []Cheader      `yaml:"header" json:"header,omitempty"`
}

// Configuration for token management.
type CtokenAction struct {
	Header string `yaml:"header" json:"header,omitempty"`
	Value  string `yaml:"value" json:"value,omitempty"`
	// +kubebuilder:validation:Enum=accept;reject;notEmpty
	Action string `yaml:"action" json:"action,omitempty"`
}

// Rule configuration.
type Config struct {
	ContentTypes []string       `yaml:"content-types" json:"content-types,omitempty"` //nolint: tagliatelle
	Dump         Cdump          `yaml:"dump" json:"dump,omitempty"`
	Force        bool           `yaml:"force" json:"force,omitempty"`
	Insecure     bool           `yaml:"insecure" json:"insecure,omitempty"`
	Port         int            `yaml:"port" json:"port,omitempty"`
	Prefix       []Creplacement `yaml:"prefix" json:"prefix,omitempty"`
	Priority     uint8          `yaml:"priority" json:"priority,omitempty"`
	Replace      []Creplacement `yaml:"replace" json:"replace,omitempty"`
	Request      Caction        `yaml:"request" json:"request,omitempty"`
	Response     Caction        `yaml:"response" json:"response,omitempty"`
	Restricted   []string       `yaml:"restricted" json:"restricted,omitempty"`
	Status       []string       `yaml:"status" json:"status,omitempty"`
	Token        []CtokenAction `yaml:"token" json:"token,omitempty"`
	Type         string         `yaml:"type" json:"type,omitempty"`
	URL          string         `yaml:"url" json:"url,omitempty"`
}
