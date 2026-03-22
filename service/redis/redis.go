package redis

import (
	"context"
	"sync"
	"time"

	"github.com/armylong/go-library/service/conf"

	"github.com/redis/go-redis/v9"
)

type Client struct {
	*redis.Client
}

var (
	redisClientMap     = map[string]*redis.Client{}
	createDbLock       sync.Mutex
	dialTimeout        time.Duration
	poolSize           int
	minIdleConns       int
	maxConnAge         time.Duration
	poolTimeout        time.Duration
	idleTimeout        time.Duration
	idleCheckFrequency time.Duration
)

func init() {
	dialTimeout = 1 * time.Second        // 拨号超时1秒（合理，快速失败）
	poolSize = 10                        // 连接池大小10（通用场景足够）
	minIdleConns = 5                     // 最小空闲连接5（合理，避免频繁创建）
	maxConnAge = 30 * time.Minute        // 修正：30分钟（而非30秒）
	poolTimeout = 3 * time.Second        // 获取连接超时3秒（合理）
	idleTimeout = 5 * time.Minute        // 修正：空闲5分钟关闭（而非30秒）
	idleCheckFrequency = 1 * time.Minute // 修正：每分钟检查一次（而非1秒）
}

func Init(serverName string) {
	createDbLock.Lock()
	defer func() {
		createDbLock.Unlock()
	}()
	cnn := func(serverName string) {
		config := conf.GetRedisConfig(serverName)
		if config == nil {
			return
		}
		option := &redis.Options{
			Addr:         config.Host + ":" + config.Port,
			Password:     config.Password,
			DB:           0,
			DialTimeout:  dialTimeout,
			PoolSize:     poolSize,
			MinIdleConns: minIdleConns,
			PoolTimeout:  poolTimeout,
		}

		client := redis.NewClient(option)
		redisClientMap[serverName] = client
		// 注意：Ping失败不会清除客户端，允许调用者处理连接错误
		// 如果需要严格的连接检查，可以在使用前自行Ping
		_ = client.Ping(context.Background())
		return
	}
	client, _ := redisClientMap[serverName]
	if client == nil {
		cnn(serverName)
		client, _ = redisClientMap[serverName]
		if client != nil {
			go func(c *redis.Client) {
				ticker := time.NewTicker(time.Second * 5)
				defer ticker.Stop()
				for {
					<-ticker.C
					if _, err := c.Ping(context.Background()).Result(); err != nil {
						cnn(serverName)
					}
				}
			}(client)
		}
	}
	return
}

func GetClient(serverName string) *Client {
	client := GetRedisClient(serverName)
	if client == nil {
		return nil
	}
	return &Client{
		Client: client,
	}
}

func GetRedisClient(serverName string) *redis.Client {
	client, _ := redisClientMap[serverName]
	if client == nil {
		Init(serverName)
		client, _ = redisClientMap[serverName]
	}
	return client
}
