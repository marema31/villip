package filter

import (
	"encoding/json"
	"io/ioutil"
	"path/filepath"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// NewFromYAML instantiate a Filter object from the configuration file.
func NewFromYAML(upLog logrus.FieldLogger, filePath string) (string, uint8, *Filter) {
	log := upLog.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c Config

	err = yaml.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode YAML: %v", err)
	}

	return _newFromConfig(upLog, c)
}

// NewFromJSON instantiate a Filter object from the configuration file.
func NewFromJSON(upLog logrus.FieldLogger, filePath string) (string, uint8, *Filter) {
	log := upLog.WithField("file", filepath.Base(filePath))

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Cannot read file: %v", err)
	}

	var c Config

	err = json.Unmarshal(content, &c)
	if err != nil {
		log.Fatalf("Cannot decode JSON: %v", err)
	}

	return _newFromConfig(upLog, c)
}
