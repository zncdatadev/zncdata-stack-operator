package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type IRedisConnector interface {
	CheckConnection() (bool, error)
}
type Connector struct {
	client *redis.Client
}

func (r Connector) CheckConnection() (bool, error) {
	_, err := r.client.Ping(context.Background()).Result()
	if err != nil {
		return false, err
	}
	return true, nil
}

type Config struct {
	Username string
	Password string
	Addr     string
}

func NewRedisConnector(config *Config) *Connector {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Username:     config.Username,
		Password:     config.Password,
		DB:           0,
		DialTimeout:  1 * time.Minute,
		WriteTimeout: 1 * time.Minute,
		ReadTimeout:  10 * time.Second,
	})

	redisConnector := &Connector{
		client: client,
	}
	return redisConnector
}
