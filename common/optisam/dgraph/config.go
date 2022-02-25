package dgraph

import "errors"

// Config holds information necessary for connecting to a database.
type Config struct {
	Hosts []string
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if len(c.Hosts) == 0 {
		return errors.New("dgraph-config: atleast one config is required")
	}
	return nil
}
