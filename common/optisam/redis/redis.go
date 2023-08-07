package redis

import (
	"github.com/go-redis/redis/v8"
)

// NewConnection returns a new database connection for the application.
func NewConnection(config Config) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     config.RedisHost,
		Password: config.RedisPassword,
		DB:       config.DB,
		Username: config.UserName,
	})
}

// NewConnection returns a new sentinel redis connection for the application.
func NewConnectionSentinel(config Config) *redis.Client {
	return redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    config.SentinelMasterName,
		Password:      config.RedisPassword,
		SentinelAddrs: []string{config.SentinelHost + ":" + config.SentinelPort},
		Username:      config.UserName,
		DB:            config.DB,
	})
}
