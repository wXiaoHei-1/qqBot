package utils

import (
	"log"
	"testing"
)

func TestConfig(t *testing.T) {
	t.Run("test Get configuration information from the yaml file", func(t *testing.T) {
		NewConfig()
		if ConfigInfo.AppID == 0 {
			log.Println("加载配置出错")
		}
		log.Println("加载配置成功", ConfigInfo)
	})
}
