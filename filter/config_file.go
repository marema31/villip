package filter

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// NewFromYAML instantiate a Filter object from the configuration file.
func (f *Factory) NewFromYAML(filePath string) (string, uint8, FilteredServer) {
	log := f.log.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c Config

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode YAML: %v", err)
	}

	return f.newFromConfig(log, c)
}

// NewFromJSON instantiate a Filter object from the configuration file.
func (f *Factory) NewFromJSON(filePath string) (string, uint8, FilteredServer) {
	log := f.log.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c Config

	err = json.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode JSON: %v", err)
	}

	return f.newFromConfig(log, c)
}
