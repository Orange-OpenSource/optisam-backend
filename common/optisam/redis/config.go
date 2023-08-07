package redis

// Config holds information necessary for connecting to a database.
type Config struct {
	RedisHost          string
	RedisPassword      string
	DB                 int
	UserName           string
	SentinelHost       string
	SentinelPort       string
	SentinelMasterName string
}

// Validate checks that the configuration is valid.
// func (c *Config) Validate() error {
// 	if c.RedisHost == "" {
// 		return errors.New("redis host is required")
// 	}
// 	return nil
// }
