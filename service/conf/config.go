package conf

import (
	"os"
	"sync"

	"github.com/spf13/cast"
)

var initOnce sync.Once
var httpPort int

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

func GetHttpPort() int {
	_init()
	return httpPort
}

func _init() {
	initOnce.Do(func() {
		httpPort := cast.ToInt(os.Getenv(`HTTP_PORT`))
		if httpPort == 0 {
			httpPort = 80
		}
	})
}
