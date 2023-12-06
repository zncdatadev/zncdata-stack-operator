package redis

import (
	"os"
	"testing"
)

func NewConfigFromEnv() *Config {
	username := os.Getenv("REDIS_USERNAME")
	password := os.Getenv("REDIS_PASSWORD")
	addr := os.Getenv("REDIS_ADDR")

	return &Config{
		Username: username,
		Password: password,
		Addr:     addr,
	}
}
func NewLocalConfig() *Config {
	return &Config{
		Addr: "36.26.69.28:16379",
	}
}

func TestNewRDBClient(t *testing.T) {
	//env := NewConfigFromEnv()
	env := NewLocalConfig()
	connector := NewRedisConnector(env)
	connection, err := connector.CheckConnection()
	if err != nil {
		t.Error("connect redis", "err: ", err)
	}
	t.Log("connect redis", "check connection result: ", connection)
}
