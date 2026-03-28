package conf

import (
	"os"
	"sync"
)

var fsInitOnce sync.Once

type FsConfig struct {
	AppId     string
	AppSecret string
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
				AppId:     appId,
				AppSecret: appSecret,
			}
			return
		}
	})
}
