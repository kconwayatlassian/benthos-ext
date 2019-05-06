package benthosx

import (
	"github.com/Jeffail/benthos/lib/config"
	"github.com/Jeffail/benthos/lib/util/text"
	"gopkg.in/yaml.v2"
)

// NewConfig parses a Benthos YAML config file and replaces all environment
// variables.
func NewConfig(b []byte) (*config.Type, error) {
	conf := config.New()
	err := yaml.Unmarshal(text.ReplaceEnvVariables(b), &conf)
	return &conf, err
}
