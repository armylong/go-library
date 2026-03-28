package conf

import (
	"os"
	"sync"
)

var fsInitOnce sync.Once

const (
	UserAccessTokenCacheKey        = "feishu:user:access:token"
	UserAccessTokenRefreshCacheKey = "feishu:user:access:refresh:token"
)

type FsConfig struct {
	AppId                          string
	AppSecret                      string
	UserAccessTokenRefreshCacheKey string
}

var fsConfig *FsConfig

func GetFsConfig() *FsConfig {
	_initFsConfig()
	return fsConfig
}

func _initFsConfig() {
	fsInitOnce.Do(func() {
		appId := os.Getenv("FEISHU_ARMYLONG_APP_ID")
		appSecret := os.Getenv("FEISHU_ARMYLONG_APP_SECRET")
		if appId != "" && appSecret != "" {
			fsConfig = &FsConfig{
				AppId:                          appId,
				AppSecret:                      appSecret,
				UserAccessTokenRefreshCacheKey: UserAccessTokenRefreshCacheKey,
			}
			return
		}
	})
}
