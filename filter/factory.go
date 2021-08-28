package filter

import (
	"os"

	"github.com/sirupsen/logrus"
)

type fNewConfig func(logrus.FieldLogger, Config) (string, uint8, FilteredServer)

// Factory provides way to create a filters.
type Factory struct {
	log           logrus.FieldLogger
	lookupEnv     func(string) (string, bool)
	newFromConfig fNewConfig
}

// NewFactory returns a Filter Factory.
func NewFactory(upLog logrus.FieldLogger) Creator {
	return &Factory{log: upLog, lookupEnv: os.LookupEnv, newFromConfig: genNewFromConfig()}
}

// Creator allow mocking of Factory.
type Creator interface {
	NewFromYAML(string) (string, uint8, FilteredServer)
	NewFromJSON(string) (string, uint8, FilteredServer)
	NewFromEnv() (string, uint8, FilteredServer)
}
