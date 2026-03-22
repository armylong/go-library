package conf

import (
	"os"
)

type RedisConfig struct {
	Host     string `json:"host"`
	Port     string `json:"port"`
	Password string `json:"password"`
}

func GetRedisConfig(serverName string) *RedisConfig {

	var host, port, password string

	if serverName == "" || serverName == "default" || serverName == "wsl" {
		host = os.Getenv("REDIS_HOST")
		port = os.Getenv("REDIS_PORT")
		password = os.Getenv("REDIS_PASSWORD")

	}

	if host == "" || port == "" || password == "" {
		return nil
	}

	config := &RedisConfig{
		Host:     host,
		Port:     port,
		Password: password,
	}
	return config
}
